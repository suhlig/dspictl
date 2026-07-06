package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SetChannelGain sets the gain for an individual channel.
func (d *Device) SetChannelGain(channel int, gain Gain) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(gain.DB())))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetChannelGain, uint16(channel), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_CHANNEL_GAIN: %w", err)
	}

	return nil
}

// GetChannelGain returns the gain for an individual channel.
func (d *Device) GetChannelGain(channel int) (Gain, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetChannelGain, uint16(channel), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_CHANNEL_GAIN: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}

// SetChannelMute mutes or unmutes an individual channel.
func (d *Device) SetChannelMute(channel int, muted bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if muted {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetChannelMute, uint16(channel), vendorInterface, []byte{val})

	if err != nil {
		return fmt.Errorf("REQ_SET_CHANNEL_MUTE: %w", err)
	}

	return nil
}

// GetChannelMute returns whether an individual channel is muted.
func (d *Device) GetChannelMute(channel int) (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetChannelMute, uint16(channel), vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_CHANNEL_MUTE: %w", err)
	}

	return buf[0] != 0, nil
}

// SetChannelDelay sets the delay for an individual channel in milliseconds.
func (d *Device) SetChannelDelay(channel int, ms float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(ms)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetDelay, uint16(channel), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_DELAY: %w", err)
	}

	return nil
}

// GetChannelDelay returns the delay for an individual channel in milliseconds.
func (d *Device) GetChannelDelay(channel int) (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetDelay, uint16(channel), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_DELAY: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}
