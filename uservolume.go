package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SetUserVolume sets the UAC/user volume in dB.
// This is the volume control exposed via the USB Audio Class interface.
// Range: -60 to 0 dB.
func (d *Device) SetUserVolume(db float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(db)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetUserVolume, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_USER_VOLUME: %w", err)
	}

	return nil
}

// GetUserVolume returns the current UAC/user volume in dB.
func (d *Device) GetUserVolume() (Gain, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetUserVolume, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_USER_VOLUME: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}
