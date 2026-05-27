package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SetLoudness enables or disables loudness compensation on the device.
func (d *Device) SetLoudness(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	var val byte
	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLoudness, 0, vendorInterface, []byte{val})
	if err != nil {
		return fmt.Errorf("REQ_SET_LOUDNESS: %w", err)
	}

	return nil
}

// GetLoudness returns whether loudness compensation is currently enabled.
func (d *Device) GetLoudness() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLoudness, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_LOUDNESS: %w", err)
	}

	return buf[0] != 0, nil
}

// SetLoudnessReference sets the reference SPL for loudness compensation (40–100 dB).
func (d *Device) SetLoudnessReference(spl float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(spl)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLoudnessReference, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_LOUDNESS_REF: %w", err)
	}

	return nil
}

// GetLoudnessReference returns the current reference SPL (dB).
func (d *Device) GetLoudnessReference() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLoudnessReference, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_LOUDNESS_REF: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}

// SetLoudnessIntensity sets the loudness compensation intensity (0–200%).
func (d *Device) SetLoudnessIntensity(pct float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(pct)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLoudnessIntensity, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_LOUDNESS_INTENSITY: %w", err)
	}

	return nil
}

// GetLoudnessIntensity returns the current loudness compensation intensity (%).
func (d *Device) GetLoudnessIntensity() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLoudnessIntensity, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_LOUDNESS_INTENSITY: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}
