package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

func (d *Device) SetOutputGain(output int, gain Gain) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(gain.DB())))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetOutputGain, uint16(output), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_GAIN: %w", err)
	}

	return nil
}

func (d *Device) GetOutputGain(output int) (Gain, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputGain, uint16(output), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_OUTPUT_GAIN: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}

func (d *Device) SetOutputMute(output int, muted bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)

	if muted {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetOutputMute, uint16(output), vendorInterface, []byte{val})

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_MUTE: %w", err)
	}

	return nil
}

func (d *Device) GetOutputMute(output int) (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputMute, uint16(output), vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_OUTPUT_MUTE: %w", err)
	}

	return buf[0] != 0, nil
}

func (d *Device) SetOutputDelay(output int, ms float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(ms)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetOutputDelay, uint16(output), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_DELAY: %w", err)
	}

	return nil
}

func (d *Device) GetOutputDelay(output int) (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputDelay, uint16(output), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_OUTPUT_DELAY: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}

func (d *Device) SetOutputEnable(output int, enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)

	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetOutputEnable, uint16(output), vendorInterface, []byte{val})

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_ENABLE: %w", err)
	}

	return nil
}

func (d *Device) GetOutputEnable(output int) (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputEnable, uint16(output), vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_OUTPUT_ENABLE: %w", err)
	}

	return buf[0] != 0, nil
}
