package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SetPreamp sets the global preamp gain (applied before per-channel preamp).
// Range is typically -24 to +24 dB.
func (d *Device) SetPreamp(db float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(db)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetPreamp, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_PREAMP: %w", err)
	}

	return nil
}

// GetPreamp returns the global preamp gain in dB.
func (d *Device) GetPreamp() (Gain, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetPreamp, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_PREAMP: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}
