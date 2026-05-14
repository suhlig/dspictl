package dspi

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

// ChannelTable returns the channel mapping for the given platform.
func ChannelTable(platform Platform) []ChannelInfo {
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

func (m MeterSnapshot) Err() error { return m.err }

// DBFS converts a linear peak (0.0–1.0) to dBFS. Returns -inf for zero.
func DBFS(linear float64) string {
	if linear <= 0 {
		return "-inf"
	}
	dbfs := 20 * math.Log10(linear)

	return fmt.Sprintf("%.1f", dbfs)
}

// Device wraps a USB connection to a DSPi.
type Device struct {
	ctx      *gousb.Context
	device   *gousb.Device
	platform Platform
	serial   string
}

// DeviceInfo describes a discovered DSPi device without an open connection.
type DeviceInfo struct {
	Serial  string
	Bus     int
	Address int
}

// List enumerates all connected DSPi devices.
func List() ([]DeviceInfo, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(dspiVID) && desc.Product == gousb.ID(dspiPID)
	})

	if err != nil {
		return nil, fmt.Errorf("enumerating DSPi devices: %w", err)
	}

	infos := make([]DeviceInfo, 0, len(devs))

	for _, dev := range devs {
		serial, err := dev.SerialNumber()

		if err != nil {
			dev.Close()

			continue
		}

		desc := dev.Desc
		infos = append(infos, DeviceInfo{
			Serial:  serial,
			Bus:     desc.Bus,
			Address: desc.Address,
		})
		dev.Close()
	}

	return infos, nil
}

// Open opens a specific DSPi device identified by info.
func Open(info DeviceInfo) (*Device, error) {
	ctx := gousb.NewContext()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(dspiVID) && desc.Product == gousb.ID(dspiPID)
	})

	if err != nil {
		ctx.Close()

		return nil, fmt.Errorf("opening DSPi device: %w", err)
	}

	var target *gousb.Device

	for _, dev := range devs {
		serial, err := dev.SerialNumber()

		if err != nil {
			dev.Close()

			continue
		}

		if serial == info.Serial {
			target = dev
		} else {
			dev.Close()
		}
	}

	if target == nil {
		ctx.Close()

		return nil, fmt.Errorf("DSPi device with serial %s not found", info.Serial)
	}

	d := &Device{
		ctx:    ctx,
		device: target,
		serial: info.Serial,
	}

	plat, err := d.detectPlatform()

	if err != nil {
		d.Close()

		return nil, fmt.Errorf("detect platform: %w", err)
	}

	d.platform = plat

	return d, nil
}

// OpenAll opens all connected DSPi devices.
func OpenAll() ([]*Device, error) {
	infos, err := List()

	if err != nil {
		return nil, err
	}

	if len(infos) == 0 {
		return nil, fmt.Errorf("no DSPi device found")
	}

	devices := make([]*Device, 0, len(infos))

	for _, info := range infos {
		dev, err := Open(info)

		if err != nil {
			continue
		}

		devices = append(devices, dev)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no DSPi device could be opened")
	}

	return devices, nil
}

// Close releases the USB device.
func (d *Device) Close() {
	if d.device != nil {
		_ = d.device.Close()
	}
	if d.ctx != nil {
		_ = d.ctx.Close()
	}
}

// Platform returns the detected hardware platform.
func (d *Device) Platform() Platform { return d.platform }

// Serial returns the unique serial number of the device.
func (d *Device) Serial() string { return d.serial }

func (d *Device) detectPlatform() (Platform, error) {
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
func (d *Device) ReadMeter() MeterSnapshot {
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
func (d *Device) ClearClips() error {
	buf := make([]byte, 2)
	_, err := d.device.Control(0xC1, reqClearClips, 0, macOSVendorInterface, buf)

	if err != nil {
		return fmt.Errorf("clearing clips: %w", err)
	}

	return nil
}
