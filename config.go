package dspi

import "fmt"

// SetOutputType sets the output type for a slot.
// outputType: 0=S/PDIF, 1=I2S.
func (d *Device) SetOutputType(slot int, outputType int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetOutputType, uint16(slot), vendorInterface, []byte{byte(outputType)})

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_TYPE: %w", err)
	}

	return nil
}

// GetOutputType returns the output type for a slot.
// 0=S/PDIF, 1=I2S.
func (d *Device) GetOutputType(slot int) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputType, uint16(slot), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_OUTPUT_TYPE: %w", err)
	}

	return int(buf[0]), nil
}

// SetOutputPin changes the GPIO pin assignment for an output.
// The protocol returns a status byte; non-zero statuses are returned as errors.
func (d *Device) SetOutputPin(output int, pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	wValue := uint16(pin)<<8 | uint16(output)
	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetOutputPin, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_OUTPUT_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetOutputPin returns the current GPIO pin for an output.
func (d *Device) GetOutputPin(output int) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputPin, uint16(output), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_OUTPUT_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// SetI2SBckPin sets the shared I2S BCK GPIO.
// The value is sent in wValue and the firmware returns a 1-byte status.
func (d *Device) SetI2SBckPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetI2SBckPin, uint16(pin), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_BCK_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_I2S_BCK_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetI2SBckPin returns the shared I2S BCK GPIO.
func (d *Device) GetI2SBckPin() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SBckPin, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_BCK_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// SetMCKEnable enables or disables the I2S master clock output.
func (d *Device) SetMCKEnable(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if enabled {
		val = 1
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetMCKEnable, uint16(val), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MCK_ENABLE: %w", err)
	}

	return nil
}

// GetMCKEnable returns whether the I2S master clock output is enabled.
func (d *Device) GetMCKEnable() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetMCKEnable, 0, vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_MCK_ENABLE: %w", err)
	}

	return buf[0] != 0, nil
}

// SetMCKPin sets the MCK GPIO pin.
func (d *Device) SetMCKPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetMCKPin, uint16(pin), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MCK_PIN: %w", err)
	}

	return nil
}

// GetMCKPin returns the MCK GPIO pin.
func (d *Device) GetMCKPin() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetMCKPin, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_MCK_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// SetMCKMultiplier sets the MCK multiplier.
// multiplier: 0=128x, 1=256x.
func (d *Device) SetMCKMultiplier(multiplier int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetMCKMultiplier, uint16(multiplier), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MCK_MULTIPLIER: %w", err)
	}

	return nil
}

// GetMCKMultiplier returns the MCK multiplier.
// 0=128x, 1=256x.
func (d *Device) GetMCKMultiplier() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetMCKMultiplier, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_MCK_MULTIPLIER: %w", err)
	}

	return int(buf[0]), nil
}
