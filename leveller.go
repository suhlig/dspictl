package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// LevellerStatus holds the current leveller (dynamic range compressor) settings.
type LevellerStatus struct {
	Enabled   bool
	Amount    float64 // compression amount
	Speed     int     // attack/release speed
	MaxGain   float64 // maximum gain reduction in dB
	Lookahead bool    // lookahead enabled
	Gate      float64 // noise gate threshold in dB
}

// SetLeveller enables or disables the leveller.
func (d *Device) SetLeveller(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLeveller, 0, vendorInterface, []byte{val})
	if err != nil {
		return fmt.Errorf("REQ_SET_LEVELLER: %w", err)
	}

	return nil
}

// GetLeveller returns whether the leveller is currently enabled.
func (d *Device) GetLeveller() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLeveller, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_LEVELLER: %w", err)
	}

	return buf[0] != 0, nil
}

// SetLevellerAmount sets the leveller compression amount.
func (d *Device) SetLevellerAmount(amount float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(amount)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLevellerAmount, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_LEVELLER_AMOUNT: %w", err)
	}

	return nil
}

// GetLevellerAmount returns the leveller compression amount.
func (d *Device) GetLevellerAmount() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLevellerAmount, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_LEVELLER_AMOUNT: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}

// SetLevellerSpeed sets the leveller attack/release speed.
func (d *Device) SetLevellerSpeed(speed int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLevellerSpeed, 0, vendorInterface, []byte{byte(speed)})
	if err != nil {
		return fmt.Errorf("REQ_SET_LEVELLER_SPEED: %w", err)
	}

	return nil
}

// GetLevellerSpeed returns the leveller attack/release speed.
func (d *Device) GetLevellerSpeed() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLevellerSpeed, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_LEVELLER_SPEED: %w", err)
	}

	return int(buf[0]), nil
}

// SetLevellerMaxGain sets the maximum gain reduction in dB.
func (d *Device) SetLevellerMaxGain(maxGain float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(maxGain)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLevellerMaxGain, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_LEVELLER_MAXGAIN: %w", err)
	}

	return nil
}

// GetLevellerMaxGain returns the maximum gain reduction in dB.
func (d *Device) GetLevellerMaxGain() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLevellerMaxGain, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_LEVELLER_MAXGAIN: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}

// SetLevellerLookahead enables or disables lookahead for the leveller.
func (d *Device) SetLevellerLookahead(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLevellerLookahead, 0, vendorInterface, []byte{val})
	if err != nil {
		return fmt.Errorf("REQ_SET_LEVELLER_LOOKAHEAD: %w", err)
	}

	return nil
}

// GetLevellerLookahead returns whether lookahead is enabled.
func (d *Device) GetLevellerLookahead() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLevellerLookahead, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_LEVELLER_LOOKAHEAD: %w", err)
	}

	return buf[0] != 0, nil
}

// SetLevellerGate sets the leveller noise gate threshold in dB.
func (d *Device) SetLevellerGate(gate float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(gate)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLevellerGate, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_LEVELLER_GATE: %w", err)
	}

	return nil
}

// GetLevellerGate returns the leveller noise gate threshold in dB.
func (d *Device) GetLevellerGate() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLevellerGate, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_LEVELLER_GATE: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}
