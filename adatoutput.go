package dspi

import (
	"encoding/binary"
	"fmt"
)

// AdatOutputStatus is the 8-byte ADAT bulk output status from
// REQ_GET_ADAT_OUTPUT_STATUS.  RP2350 only; RP2040 returns zeros.
type AdatOutputStatus struct {
	Enabled     bool   // configured enable (persisted intent)
	Active      bool   // stream currently running
	Pin         uint8  // configured data GPIO
	RateOK      bool   // current sample rate is 44.1/48 kHz
	ResyncCount uint16 // stream restarts since boot
	SlipCount   uint16 // emergency local resyncs since boot
}

// DecodeAdatOutputStatus parses an 8-byte status payload.
func DecodeAdatOutputStatus(raw []byte) (*AdatOutputStatus, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("ADAT output status payload too short (got %d, need 8)", len(raw))
	}
	return &AdatOutputStatus{
		Enabled:     raw[0] != 0,
		Active:      raw[1] != 0,
		Pin:         raw[2],
		RateOK:      raw[3] != 0,
		ResyncCount: binary.LittleEndian.Uint16(raw[4:6]),
		SlipCount:   binary.LittleEndian.Uint16(raw[6:8]),
	}, nil
}

// SetAdatOutputEnable enables or disables the ADAT bulk output
// (REQ_SET_ADAT_OUTPUT_ENABLE; RP2350 only).  The value travels in wValue
// and the firmware returns a PIN_CONFIG_* status byte.
func (d *Device) SetAdatOutputEnable(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	wValue := uint16(0)
	if enabled {
		wValue = 1
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetAdatOutputEnable, wValue, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_ADAT_OUTPUT_ENABLE: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_ADAT_OUTPUT_ENABLE: status 0x%02X", buf[0])
	}

	return nil
}

// GetAdatOutputEnable returns the configured ADAT output enable state.
func (d *Device) GetAdatOutputEnable() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatOutputEnable, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_ADAT_OUTPUT_ENABLE: %w", err)
	}

	return buf[0] != 0, nil
}

// SetAdatOutputPin sets the ADAT output data GPIO (REQ_SET_ADAT_OUTPUT_PIN;
// RP2350 only).  Pass AdatInputPinUnset (0xFF) to restore the platform
// default.  Re-routing is allowed even while enabled; the deferred apply
// moves the stream under mute.
func (d *Device) SetAdatOutputPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetAdatOutputPin, uint16(pin), vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_ADAT_OUTPUT_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_ADAT_OUTPUT_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetAdatOutputPin returns the configured ADAT output GPIO.
func (d *Device) GetAdatOutputPin() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatOutputPin, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_ADAT_OUTPUT_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// GetAdatOutputStatus reads the 8-byte ADAT output status packet.
func (d *Device) GetAdatOutputStatus() (*AdatOutputStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatOutputStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_ADAT_OUTPUT_STATUS: %w", err)
	}

	return DecodeAdatOutputStatus(buf[:n])
}
