package dspi

import (
	"encoding/binary"
	"fmt"
)

// UartCtrlConfig is the 8-byte UART control interface configuration
// (REQ_SET/GET_UART_CONFIG).  Fixed 8N1 framing; only the baud rate is
// configurable besides the pins and enable.  SET is USB-only: the firmware
// rejects it over UART or I2C so an external controller can never
// reconfigure its own transport.
type UartCtrlConfig struct {
	Enabled      bool   // 0/1
	TxPin        uint8  // GPIO with UARTx TX mux (pin%4 == 0)
	RxPin        uint8  // GPIO with UARTx RX mux (pin%4 == 1), same instance
	NotifyEnable bool   // 0/1: push async notification frames (type 0x40)
	Baud         uint32 // UART_CTRL_BAUD_MIN..MAX (9600..1000000)
}

// Encode serialises the config into the 8-byte wire layout.
func (c *UartCtrlConfig) Encode() []byte {
	buf := make([]byte, 8)
	if c.Enabled {
		buf[0] = 1
	}
	buf[1] = c.TxPin
	buf[2] = c.RxPin
	if c.NotifyEnable {
		buf[3] = 1
	}
	binary.LittleEndian.PutUint32(buf[4:8], c.Baud)
	return buf
}

// DecodeUartCtrlConfig parses an 8-byte UART config payload.
func DecodeUartCtrlConfig(raw []byte) (*UartCtrlConfig, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("UART config payload too short (got %d, need 8)", len(raw))
	}
	return &UartCtrlConfig{
		Enabled:      raw[0] != 0,
		TxPin:        raw[1],
		RxPin:        raw[2],
		NotifyEnable: raw[3] != 0,
		Baud:         binary.LittleEndian.Uint32(raw[4:8]),
	}, nil
}

// I2cCtrlConfig is the 8-byte I2C target control interface configuration
// (REQ_SET/GET_I2C_CONFIG).  SET is USB-only.
type I2cCtrlConfig struct {
	Enabled bool  // 0/1
	SdaPin  uint8 // GPIO with I2Cx SDA mux (pin%2 == 0)
	SclPin  uint8 // GPIO with I2Cx SCL mux (pin%2 == 1), same instance
	Address uint8 // 7-bit target address, 0x08..0x77
}

// Encode serialises the config into the 8-byte wire layout.
func (c *I2cCtrlConfig) Encode() []byte {
	buf := make([]byte, 8)
	if c.Enabled {
		buf[0] = 1
	}
	buf[1] = c.SdaPin
	buf[2] = c.SclPin
	buf[3] = c.Address
	return buf
}

// DecodeI2cCtrlConfig parses an 8-byte I2C config payload.
func DecodeI2cCtrlConfig(raw []byte) (*I2cCtrlConfig, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("I2C config payload too short (got %d, need 8)", len(raw))
	}
	return &I2cCtrlConfig{
		Enabled: raw[0] != 0,
		SdaPin:  raw[1],
		SclPin:  raw[2],
		Address: raw[3],
	}, nil
}

// CtrlIfaceStatus is the 8-byte REQ_GET_CTRL_IFACE_STATUS response.
// last_status is the PIN_CONFIG_* result of the most recent SET; live
// reflects whether the interface is actually up (differs from config.enabled
// after a boot-time pin collision kept it down).
type CtrlIfaceStatus struct {
	UartLastStatus uint8
	UartLive       bool
	I2cLastStatus  uint8
	I2cLive        bool
	ProtoVersion   uint8
}

// DecodeCtrlIfaceStatus parses the 8-byte status payload.
func DecodeCtrlIfaceStatus(raw []byte) (*CtrlIfaceStatus, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("control interface status payload too short (got %d, need 8)", len(raw))
	}
	return &CtrlIfaceStatus{
		UartLastStatus: raw[0],
		UartLive:       raw[1] != 0,
		I2cLastStatus:  raw[2],
		I2cLive:        raw[3] != 0,
		ProtoVersion:   raw[4],
	}, nil
}

// SetUartConfig configures the UART control interface (REQ_SET_UART_CONFIG;
// USB-only).  The PIN_CONFIG_* outcome is read back via GetCtrlIfaceStatus.
func (d *Device) SetUartConfig(cfg *UartCtrlConfig) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetUartConfig, 0, vendorInterface, cfg.Encode())
	if err != nil {
		return fmt.Errorf("REQ_SET_UART_CONFIG: %w", err)
	}

	return nil
}

// GetUartConfig returns the persisted UART control interface configuration.
func (d *Device) GetUartConfig() (*UartCtrlConfig, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetUartConfig, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_UART_CONFIG: %w", err)
	}

	return DecodeUartCtrlConfig(buf[:n])
}

// SetI2CConfig configures the I2C target control interface (REQ_SET_I2C_CONFIG;
// USB-only).  The PIN_CONFIG_* outcome is read back via GetCtrlIfaceStatus.
func (d *Device) SetI2CConfig(cfg *I2cCtrlConfig) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetI2CConfig, 0, vendorInterface, cfg.Encode())
	if err != nil {
		return fmt.Errorf("REQ_SET_I2C_CONFIG: %w", err)
	}

	return nil
}

// GetI2CConfig returns the persisted I2C target control interface configuration.
func (d *Device) GetI2CConfig() (*I2cCtrlConfig, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2CConfig, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_I2C_CONFIG: %w", err)
	}

	return DecodeI2cCtrlConfig(buf[:n])
}

// GetCtrlIfaceStatus reads the UART/I2C control interface live status.
func (d *Device) GetCtrlIfaceStatus() (*CtrlIfaceStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCtrlIfaceStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_CTRL_IFACE_STATUS: %w", err)
	}

	return DecodeCtrlIfaceStatus(buf[:n])
}
