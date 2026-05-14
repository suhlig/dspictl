package dspi

import (
	"encoding/binary"
	"fmt"

	"github.com/google/gousb"
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
	buf := make([]byte, maxChannels*2+5)

	n, err := d.device.Control(0xC1, reqGetStatus, 9, macOSVendorInterface, buf)

	if err != nil {
		snap.err = fmt.Errorf("REQ_GET_STATUS: %w", err)

		return snap
	}

	if (n-5) > 0 && (n-5)%2 == 0 {
		snap.Channels = (n - 5) / 2
	} else if (n-4) > 0 && (n-4)%2 == 0 {
		snap.Channels = (n - 4) / 2
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
