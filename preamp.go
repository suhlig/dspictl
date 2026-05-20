package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

func (d *Device) SetPreampChannel(channel int, gain Gain) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(gain.DB())))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetPreampCh, uint16(channel), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_PREAMP_CH: %w", err)
	}

	return nil
}

func (d *Device) GetPreampChannel(channel int) (Gain, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetPreampCh, uint16(channel), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_PREAMP_CH: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}
