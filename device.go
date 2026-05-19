package dspi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/google/gousb"
)

const (
	// vendorInterfaceInRequest is the USB bmRequestType for vendor-specific,
	// device-to-host (IN) control transfers addressed to an interface.
	vendorInterfaceInRequest = gousb.ControlIn | gousb.ControlVendor | gousb.ControlInterface

	// vendorInterfaceOutRequest is the USB bmRequestType for vendor-specific,
	// host-to-device (OUT) control transfers addressed to an interface.
	vendorInterfaceOutRequest = gousb.ControlOut | gousb.ControlVendor | gousb.ControlInterface
)

// Device wraps a USB connection to a DSPi.
type Device struct {
	ctx      *gousb.Context
	device   *gousb.Device
	platform Platform
	serial   string
	bus      int
	address  int
}

// Open opens a specific DSPi device identified by info.
func Open(info DeviceInfo) (*Device, error) {
	ctx := gousb.NewContext()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(dspiVID) && desc.Product == gousb.ID(dspiPID)
	})

	if err != nil {
		_ = ctx.Close()

		return nil, fmt.Errorf("opening DSPi device: %w", err)
	}

	var target *gousb.Device

	for _, dev := range devs {
		serial, err := dev.SerialNumber()

		if err != nil {
			_ = dev.Close()

			continue
		}

		if serial == info.Serial {
			target = dev
		} else {
			_ = dev.Close()
		}
	}

	if target == nil {
		_ = ctx.Close()

		return nil, fmt.Errorf("DSPi device with serial %s not found", info.Serial)
	}

	d := &Device{
		ctx:     ctx,
		device:  target,
		serial:  info.Serial,
		bus:     info.Bus,
		address: info.Address,
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
	var errs []error

	for _, info := range infos {
		dev, err := Open(info)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", info.Serial, err))

			continue
		}

		devices = append(devices, dev)
	}

	if len(devices) == 0 {
		if len(errs) > 0 {
			return nil, fmt.Errorf("no DSPi device could be opened: %w", errors.Join(errs...))
		}

		return nil, fmt.Errorf("no DSPi device could be opened")
	}

	return devices, nil
}

// Close releases the USB device.
func (d *Device) Close() {
	if d.device != nil {
		_ = d.device.Close()
		d.device = nil
	}
	if d.ctx != nil {
		_ = d.ctx.Close()
		d.ctx = nil
	}
}

// Platform returns the detected hardware platform.
func (d *Device) Platform() Platform { return d.platform }

// Serial returns the unique serial number of the device.
func (d *Device) Serial() string { return d.serial }

// Bus returns the USB bus number of the device.
func (d *Device) Bus() int { return d.bus }

// Address returns the USB device address on the bus.
func (d *Device) Address() int { return d.address }

func (d *Device) detectPlatform() (Platform, error) {
	buf := make([]byte, 4)
	_, err := d.device.Control(vendorInterfaceInRequest, reqGetPlatform, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_PLATFORM: %w", err)
	}
	if len(buf) < 1 {
		return 0, fmt.Errorf("REQ_GET_PLATFORM: empty response")
	}

	return Platform(buf[0]), nil
}

// ReadMeter polls the device for combined status (wValue=9).
func (d *Device) ReadMeter() MeterSnapshot {
	var snap MeterSnapshot

	if d.device == nil {
		snap.err = fmt.Errorf("device is closed")

		return snap
	}

	buf := make([]byte, maxChannels*2+5)

	n, err := d.device.Control(vendorInterfaceInRequest, reqGetStatus, 9, vendorInterface, buf)

	if err != nil {
		snap.err = fmt.Errorf("REQ_GET_STATUS: %w", err)

		return snap
	}

	// Determine channel count from the response length.
	// Response format (no header): [peak pairs...] [CPU0(1)] [CPU1(1)] [ClipFlags(2)]
	// That is channels*2 + 4 bytes.
	// Some firmware versions may include a 1-byte header, making it channels*2 + 5.
	switch {
	case n >= 4 && (n-4)%2 == 0:
		snap.Channels = (n - 4) / 2
	case n >= 6 && (n-5)%2 == 0:
		snap.Channels = (n - 5) / 2
	default:
		snap.err = fmt.Errorf("REQ_GET_STATUS: unexpected response length %d", n)

		return snap
	}

	actualCh := min(snap.Channels, maxChannels)

	needed := actualCh*2 + 4

	if n < needed {
		snap.err = fmt.Errorf("REQ_GET_STATUS: response too short (got %d, need %d)", n, needed)

		return snap
	}

	for i := range actualCh {
		raw := binary.LittleEndian.Uint16(buf[i*2:])
		snap.Peaks[i] = NewLevel(normalize(raw))
	}

	offset := actualCh * 2
	snap.CPU0 = int(buf[offset])
	snap.CPU1 = int(buf[offset+1])
	snap.ClipFlags = binary.LittleEndian.Uint16(buf[offset+2:])

	return snap
}

// ClearClips sends REQ_CLEAR_CLIPS (0x83) to reset the clip bitmask on the device.
func (d *Device) ClearClips() error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 2)
	_, err := d.device.Control(vendorInterfaceInRequest, reqClearClips, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("clearing clips: %w", err)
	}

	return nil
}

// SetMasterVolume sets the device-side master volume.
// Range: -128 (mute sentinel) to 0 dB. Typical range: -127 to 0 dB.
func (d *Device) SetMasterVolume(g Gain) error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(g.DB())))
	_, err := d.device.Control(vendorInterfaceOutRequest, reqSetMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MASTER_VOLUME: %w", err)
	}

	return nil
}

// GetMasterVolume reads the current device-side master volume.
func (d *Device) GetMasterVolume() (Gain, error) {
	if d.device == nil {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.device.Control(vendorInterfaceInRequest, reqGetMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_MASTER_VOLUME: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}

// ChannelName fetches the name of a channel from the device hardware.
// The channelIndex should be 0-based.
func (d *Device) ChannelName(channelIndex int) (string, error) {
	if d.device == nil {
		return "", fmt.Errorf("device is closed")
	}

	buf := make([]byte, 64)
	n, err := d.device.Control(vendorInterfaceInRequest, reqGetChannelName, uint16(channelIndex), vendorInterface, buf)

	if err != nil {
		return "", fmt.Errorf("REQ_GET_CHANNEL_NAME: %w", err)
	}

	// Find null terminator or use full buffer
	end := bytes.IndexByte(buf[:n], 0)
	if end == -1 {
		end = n
	}

	return string(bytes.TrimSpace(buf[:end])), nil
}

// Channels queries the device for all channel names and returns a slice of ChannelInfo.
// The number of channels is determined by the platform (7 for RP2040, 11 for RP2350).
func (d *Device) Channels() ([]ChannelInfo, error) {
	if d.device == nil {
		return nil, fmt.Errorf("device is closed")
	}

	channelCount := 7
	if d.platform == PlatformRP2350 {
		channelCount = 11
	}

	channels := make([]ChannelInfo, 0, channelCount)

	for i := 0; i < channelCount; i++ {
		name, err := d.ChannelName(i)
		if err != nil {
			return nil, fmt.Errorf("reading channel %d name: %w", i, err)
		}

		// Fallback to default name if device returns empty
		if name == "" {
			name = defaultChannelName(i, d.platform)
		}

		channels = append(channels, ChannelInfo{
			Index: i,
			Name:  name,
			Group: channelGroup(i, d.platform),
		})
	}

	return channels, nil
}

// defaultChannelName returns a default name for a channel based on its index and platform.
func defaultChannelName(index int, platform Platform) string {
	if platform == PlatformRP2350 {
		switch index {
		case 0:
			return "USB L"
		case 1:
			return "USB R"
		case 2:
			return "SPDIF 1 L"
		case 3:
			return "SPDIF 1 R"
		case 4:
			return "SPDIF 2 L"
		case 5:
			return "SPDIF 2 R"
		case 6:
			return "SPDIF 3 L"
		case 7:
			return "SPDIF 3 R"
		case 8:
			return "SPDIF 4 L"
		case 9:
			return "SPDIF 4 R"
		case 10:
			return "PDM Sub"
		}
	} else {
		switch index {
		case 0:
			return "USB L"
		case 1:
			return "USB R"
		case 2:
			return "SPDIF 1 L"
		case 3:
			return "SPDIF 1 R"
		case 4:
			return "SPDIF 2 L"
		case 5:
			return "SPDIF 2 R"
		case 6:
			return "PDM Sub"
		}
	}

	return fmt.Sprintf("Ch %d", index)
}

// channelGroup returns the group name for a channel based on its index and platform.
func channelGroup(index int, platform Platform) string {
	maxIndex := 6
	if platform == PlatformRP2350 {
		maxIndex = 10
	}

	switch {
	case index <= 1:
		return "USB Input"
	case index < maxIndex:
		return "S/PDIF Output"
	default:
		return "PDM Sub"
	}
}

func normalize(raw uint16) float64 {
	// maxInt16 is the maximum value of a 16-bit signed integer,
	// used to normalize raw ADC readings to the 0.0–1.0 range.
	const maxInt16 = 32767

	return float64(raw) / maxInt16
}
