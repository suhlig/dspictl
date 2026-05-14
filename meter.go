package main

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/google/gousb"
)

const (
	dspiVID              = 0x2e8b // Raspberry Pi (Pico vendor)
	dspiPID              = 0xfeaa // DSPi product ID
	reqGetStatus         = 0x50
	reqGetPlatform       = 0x7F
	reqClearClips        = 0x83
	macOSVendorInterface = 0
	maxChannels          = 11
)

// Platform identifies the DSPi hardware platform.
type Platform int

const (
	PlatformRP2040 Platform = 0
	PlatformRP2350 Platform = 1
)

func (p Platform) String() string {
	switch p {
	case PlatformRP2040:
		return "RP2040"
	case PlatformRP2350:
		return "RP2350"
	default:
		return "Unknown"
	}
}

// numChannels returns the number of metered channels for the platform.
func (p Platform) numChannels() int {
	if p == PlatformRP2350 {
		return 11
	}
	return 7
}

// ChannelInfo describes one metered channel.
type ChannelInfo struct {
	Index int
	Name  string
	Group string
}

// channelTable returns the channel mapping for the given platform.
func channelTable(platform Platform) []ChannelInfo {
	if platform == PlatformRP2350 {
		return []ChannelInfo{
			{0, "USB L", "USB Input"},
			{1, "USB R", "USB Input"},
			{2, "SPDIF 1 L", "S/PDIF Output"},
			{3, "SPDIF 1 R", "S/PDIF Output"},
			{4, "SPDIF 2 L", "S/PDIF Output"},
			{5, "SPDIF 2 R", "S/PDIF Output"},
			{6, "SPDIF 3 L", "S/PDIF Output"},
			{7, "SPDIF 3 R", "S/PDIF Output"},
			{8, "SPDIF 4 L", "S/PDIF Output"},
			{9, "SPDIF 4 R", "S/PDIF Output"},
			{10, "PDM Sub", "PDM Sub"},
		}
	}
	return []ChannelInfo{
		{0, "USB L", "USB Input"},
		{1, "USB R", "USB Input"},
		{2, "SPDIF 1 L", "S/PDIF Output"},
		{3, "SPDIF 1 R", "S/PDIF Output"},
		{4, "SPDIF 2 L", "S/PDIF Output"},
		{5, "SPDIF 2 R", "S/PDIF Output"},
		{6, "PDM Sub", "PDM Sub"},
	}
}

// MeterSnapshot contains a single poll of all telemetry data.
type MeterSnapshot struct {
	Peaks     [maxChannels]float64 // 0.0 – 1.0 linear scale
	CPU0      int                  // Core 0 load 0–100%
	CPU1      int                  // Core 1 load 0–100%
	ClipFlags uint16               // sticky per-channel clip bitmask
	Channels  int                  // actual channel count for current platform
	err       error
}

// Err returns any error that occurred during the poll.
func (m MeterSnapshot) Err() error { return m.err }

// DBFS converts a linear peak (0.0–1.0) to dBFS. Returns -inf for zero.
func DBFS(linear float64) string {
	if linear <= 0 {
		return "-inf"
	}
	dbfs := 20 * math.Log10(linear)
	return fmt.Sprintf("%.1f", dbfs)
}

// DSPiDevice wraps a USB connection to a DSPi.
type DSPiDevice struct {
	ctx      *gousb.Context
	device   *gousb.Device
	platform Platform
}

// OpenDSPi finds and opens the first connected DSPi device.
func OpenDSPi() (*DSPiDevice, error) {
	ctx := gousb.NewContext()

	dev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(dspiVID), gousb.ID(dspiPID))
	if err != nil {
		ctx.Close()
		return nil, fmt.Errorf("open DSPi: %w", err)
	}
	if dev == nil {
		ctx.Close()
		return nil, fmt.Errorf("no DSPi device found (VID %04x PID %04x)\n"+
			"Make sure the DSPi is connected via USB", dspiVID, dspiPID)
	}

	d := &DSPiDevice{
		ctx:    ctx,
		device: dev,
	}

	plat, err := d.detectPlatform()
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("detect platform: %w", err)
	}
	d.platform = plat

	return d, nil
}

// Close releases the USB device.
func (d *DSPiDevice) Close() {
	if d.device != nil {
		d.device.Close()
	}
	if d.ctx != nil {
		d.ctx.Close()
	}
}

// Platform returns the detected hardware platform.
func (d *DSPiDevice) Platform() Platform { return d.platform }

// detectPlatform queries the device for its platform type via REQ_GET_PLATFORM (0x7F).
func (d *DSPiDevice) detectPlatform() (Platform, error) {
	buf := make([]byte, 4)
	_, err := d.device.Control(0xC1, reqGetPlatform, 0, macOSVendorInterface, buf)
	if err != nil {
		return PlatformRP2040, fmt.Errorf("REQ_GET_PLATFORM: %w", err)
	}
	if len(buf) < 1 {
		return PlatformRP2040, nil
	}
	return Platform(buf[0]), nil
}

// ReadMeter polls the device for combined status (wValue=9).
func (d *DSPiDevice) ReadMeter() MeterSnapshot {
	var snap MeterSnapshot
	nc := d.platform.numChannels()
	snap.Channels = nc

	reqLen := nc*2 + 5
	buf := make([]byte, reqLen)

	n, err := d.device.Control(0xC1, reqGetStatus, 9, macOSVendorInterface, buf)
	if err != nil {
		snap.err = fmt.Errorf("REQ_GET_STATUS: %w", err)
		return snap
	}

	respLen := n
	if (respLen-5) > 0 && (respLen-5)%2 == 0 {
		snap.Channels = (respLen - 5) / 2
	} else if (respLen-4) > 0 && (respLen-4)%2 == 0 {
		snap.Channels = (respLen - 4) / 2
	}

	actualCh := snap.Channels
	if actualCh > maxChannels {
		actualCh = maxChannels
	}

	for i := 0; i < actualCh; i++ {
		raw := binary.LittleEndian.Uint16(buf[i*2:])
		snap.Peaks[i] = float64(raw) / 32767.0
	}

	offset := actualCh * 2
	snap.CPU0 = int(buf[offset])
	snap.CPU1 = int(buf[offset+1])
	snap.ClipFlags = binary.LittleEndian.Uint16(buf[offset+2:])

	return snap
}

// ClearClips sends REQ_CLEAR_CLIPS (0x83) to reset the clip bitmask on the device.
func (d *DSPiDevice) ClearClips() {
	buf := make([]byte, 2)
	_, _ = d.device.Control(0xC1, reqClearClips, 0, macOSVendorInterface, buf)
}
