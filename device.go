package dspi

import (
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

	vendorInterfaceOutRequest = gousb.ControlOut | gousb.ControlVendor | gousb.ControlInterface

	// https://github.com/WeebLabs/DSPi/blob/5c71c5d2a09b25761abf3013781aa6a905cc001c/firmware/DSPi/config.h
	reqGetStatus       = 0x50
	reqGetPlatform     = 0x7F
	reqClearClips      = 0x83
	reqSetMasterVolume = 0xD2
	reqGetMasterVolume = 0xD3
)

// Device wraps a USB connection to a DSPi.
type Device struct {
	ctx      *gousb.Context
	device   *gousb.Device
	platform Platform
	serial   string
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

// SetMasterVolume sets the device-side master volume in dB.
// Range: -128 (mute sentinel) to 0 dB. Typical range: -127 to 0 dB.
func (d *Device) SetMasterVolume(db float64) error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(db)))
	_, err := d.device.Control(vendorInterfaceOutRequest, reqSetMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MASTER_VOLUME: %w", err)
	}

	return nil
}

// GetMasterVolume reads the current device-side master volume in dB.
func (d *Device) GetMasterVolume() (float64, error) {
	if d.device == nil {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.device.Control(vendorInterfaceInRequest, reqGetMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_MASTER_VOLUME: %w", err)
	}

	db := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))

	return db, nil
}

func normalize(raw uint16) float64 {
	// maxInt16 is the maximum value of a 16-bit signed integer,
	// used to normalize raw ADC readings to the 0.0–1.0 range.
	const maxInt16 = 32767

	return float64(raw) / maxInt16
}
