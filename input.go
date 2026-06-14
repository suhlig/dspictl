package dspi

import (
	"encoding/binary"
	"fmt"
)

// InputSourceName returns a human-readable name for an input source value.
func InputSourceName(source int) string {
	switch source {
	case InputSourceUSB:
		return "USB"
	case InputSourceSPDIF:
		return "S/PDIF"
	case InputSourceI2S:
		return "I2S"
	default:
		return fmt.Sprintf("Unknown(%d)", source)
	}
}

// InputRate holds the sample rate information from REQ_GET_INPUT_RATE.
type InputRate struct {
	PipelineHz uint32
	SelectedHz uint32
}

// I2SInputRateEnum converts a Hz value to the wire enum (0=44100, 1=48000, 2=96000).
func I2SInputRateEnum(hz uint32) int {
	switch hz {
	case 44100:
		return 0
	case 48000:
		return 1
	case 96000:
		return 2
	default:
		return -1
	}
}

// I2SInputRateHz converts a wire enum to Hz.
func I2SInputRateHz(enum int) uint32 {
	switch enum {
	case 0:
		return 44100
	case 1:
		return 48000
	case 2:
		return 96000
	default:
		return 0
	}
}

// SetInputSource selects the active input source.
func (d *Device) SetInputSource(source int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetInputSource, 0, vendorInterface, []byte{byte(source)})

	if err != nil {
		return fmt.Errorf("REQ_SET_INPUT_SOURCE: %w", err)
	}

	return nil
}

// GetInputSource returns the active input source.
func (d *Device) GetInputSource() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetInputSource, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_INPUT_SOURCE: %w", err)
	}

	return int(buf[0]), nil
}

// SetInputRate sets the I2S input sample rate.
// Valid values: 44100, 48000, 96000 Hz.
func (d *Device) SetInputRate(hz uint32) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, hz)

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetInputRate, 0, vendorInterface, payload)

	if err != nil {
		return fmt.Errorf("REQ_SET_INPUT_RATE: %w", err)
	}

	return nil
}

// GetInputRate returns the current pipeline rate and the selected I2S rate.
func (d *Device) GetInputRate() (InputRate, error) {
	if d.closed {
		return InputRate{}, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetInputRate, 0, vendorInterface, buf)

	if err != nil {
		return InputRate{}, fmt.Errorf("REQ_GET_INPUT_RATE: %w", err)
	}

	return InputRate{
		PipelineHz: binary.LittleEndian.Uint32(buf[0:4]),
		SelectedHz: binary.LittleEndian.Uint32(buf[4:8]),
	}, nil
}

// SetI2SRxPin sets the I2S RX data GPIO pin.
func (d *Device) SetI2SRxPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetI2SRxPin, uint16(pin), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_RX_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_I2S_RX_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetI2SRxPin returns the current I2S RX data GPIO pin.
func (d *Device) GetI2SRxPin() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SRxPin, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_RX_PIN: %w", err)
	}

	return int(buf[0]), nil
}
