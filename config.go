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

// SetI2SBckPin sets the shared I2S BCK GPIO (master/unified pair, role 0).
// This is a convenience wrapper around the role-aware SetI2SBckPinRole.
func (d *Device) SetI2SBckPin(pin int) error {
	return d.SetI2SBckPinRole(I2SBckRoleMaster, pin)
}

// SetI2SBckPinRole sets the I2S BCK GPIO for a clock role (0 = master/unified
// pair, 1 = slave pair).  wValue is encoded as (role << 8) | pin; pass
// AdatInputPinUnset (0xFF) as the pin to restore the role's platform default.
// The firmware returns a PIN_CONFIG_* status byte.
func (d *Device) SetI2SBckPinRole(role, pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	wValue := uint16(role)<<8 | uint16(pin)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetI2SBckPin, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_BCK_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_I2S_BCK_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetI2SBckPin returns the shared I2S BCK GPIO (master/unified pair, role 0).
// This is a convenience wrapper around the role-aware GetI2SBckPinRole.
func (d *Device) GetI2SBckPin() (int, error) {
	return d.GetI2SBckPinRole(I2SBckRoleMaster)
}

// GetI2SBckPinRole returns the I2S BCK GPIO for a clock role
// (0 = master/unified pair, 1 = slave pair).
func (d *Device) GetI2SBckPinRole(role int) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SBckPin, uint16(role), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_BCK_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// SetI2SClockPinMode selects whether master and slave clock modes share one
// BCK/LRCLK pair (0 = unified) or each have their own (1 = split;
// REQ_SET_I2S_CLOCK_PIN_MODE).  The firmware returns a PIN_CONFIG_* status byte.
func (d *Device) SetI2SClockPinMode(mode int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetI2SClockPinMode, uint16(mode), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_CLOCK_PIN_MODE: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_I2S_CLOCK_PIN_MODE: status 0x%02X", buf[0])
	}

	return nil
}

// GetI2SClockPinMode returns the live I2S clock-pin mode (0 = unified, 1 = split).
func (d *Device) GetI2SClockPinMode() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SClockPinMode, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_CLOCK_PIN_MODE: %w", err)
	}

	return int(buf[0]), nil
}

// SetMCKEnable enables or disables the I2S master clock output.
// The enable flag is carried in wValue and the firmware returns a 1-byte status.
func (d *Device) SetMCKEnable(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	wValue := uint16(0)
	if enabled {
		wValue = 1
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetMCKEnable, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MCK_ENABLE: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_MCK_ENABLE: status 0x%02X", buf[0])
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
// The pin is carried in wValue and the firmware returns a 1-byte status.
func (d *Device) SetMCKPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetMCKPin, uint16(pin), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MCK_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_MCK_PIN: status 0x%02X", buf[0])
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
// The multiplier is carried in wValue and the firmware returns a 1-byte status.
func (d *Device) SetMCKMultiplier(multiplier int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetMCKMultiplier, uint16(multiplier), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MCK_MULTIPLIER: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_MCK_MULTIPLIER: status 0x%02X", buf[0])
	}

	return nil
}

// SaveOutputConfig saves the current output pin/type configuration to flash.
func (d *Device) SaveOutputConfig() error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSaveOutputConfig, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SAVE_OUTPUT_CONFIG: %w", err)
	}

	if len(buf) > 0 && buf[0] != 0 {
		return fmt.Errorf("REQ_SAVE_OUTPUT_CONFIG: status 0x%02X", buf[0])
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
