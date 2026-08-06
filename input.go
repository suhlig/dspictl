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
	case InputSourceADAT:
		return "ADAT"
	case InputSourceSPDIF2:
		return "S/PDIF 2"
	case InputSourceSPDIF3:
		return "S/PDIF 3"
	case InputSourceSPDIF4:
		return "S/PDIF 4"
	default:
		return fmt.Sprintf("Unknown(%d)", source)
	}
}

// AdatInputState represents the ADAT input receiver lock state.
type AdatInputState int

const (
	AdatInputInactive  AdatInputState = 0
	AdatInputAcquiring AdatInputState = 1
	AdatInputSyncing   AdatInputState = 2
	AdatInputLocked    AdatInputState = 3
	AdatInputRelocking AdatInputState = 4
)

// String returns a human-readable name for the ADAT input state.
func (s AdatInputState) String() string {
	switch s {
	case AdatInputInactive:
		return "inactive"
	case AdatInputAcquiring:
		return "acquiring"
	case AdatInputSyncing:
		return "syncing"
	case AdatInputLocked:
		return "locked"
	case AdatInputRelocking:
		return "relocking"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// AdatInputStatus holds the decoded 20-byte status packet from REQ_GET_ADAT_INPUT_STATUS.
type AdatInputStatus struct {
	State        AdatInputState
	ClockMode    int
	Enabled      bool
	Pin          uint8
	RateOK       bool
	LockCount    uint8
	LossCount    uint8
	SlipCount    uint8
	HeaderErrors uint16
	DetectedRate uint32
	MeasuredHz   uint32
}

// AdatClockModeName returns a human-readable name for the ADAT input clock mode.
func AdatClockModeName(mode int) string {
	switch mode {
	case AdatClockModeMaster:
		return "master"
	case AdatClockModeSlave:
		return "slave"
	default:
		return fmt.Sprintf("unknown(%d)", mode)
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

// SetI2SRxPin sets the I2S RX data GPIO pin for pair 0.
// This is a convenience wrapper around the pair-aware SetI2SRxPinPair.
func (d *Device) SetI2SRxPin(pin int) error {
	return d.SetI2SRxPinPair(0, pin)
}

// SetI2SRxPinPair sets the I2S RX data GPIO pin for the given I2S pair.
// wValue is encoded as (pair << 8) | pin.
func (d *Device) SetI2SRxPinPair(pair, pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	wValue := uint16(pair)<<8 | uint16(pin)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetI2SRxPin, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_RX_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_I2S_RX_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetI2SRxPin returns the current I2S RX data GPIO pin for pair 0.
// This is a convenience wrapper around the pair-aware GetI2SRxPinPair.
func (d *Device) GetI2SRxPin() (int, error) {
	return d.GetI2SRxPinPair(0)
}

// GetI2SRxPinPair returns the current I2S RX data GPIO pin for the given I2S pair.
func (d *Device) GetI2SRxPinPair(pair int) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SRxPin, uint16(pair), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_RX_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// SetI2SInputChannels sets the number of I2S input channels.
// Valid values: 2, 4, 6, 8.
func (d *Device) SetI2SInputChannels(channels int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetI2SInputChannels, 0, vendorInterface, []byte{byte(channels)})

	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_INPUT_CHANNELS: %w", err)
	}

	return nil
}

// GetI2SInputChannels returns the number of I2S input channels.
// Returns 0 if the field is absent.
func (d *Device) GetI2SInputChannels() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SInputChannels, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_INPUT_CHANNELS: %w", err)
	}

	return int(buf[0]), nil
}

// GetAdatInputStatus reads the 20-byte ADAT input status packet.
func (d *Device) GetAdatInputStatus() (*AdatInputStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 20)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatInputStatus, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_ADAT_INPUT_STATUS: %w", err)
	}

	return &AdatInputStatus{
		State:        AdatInputState(buf[0]),
		ClockMode:    int(buf[1]),
		Enabled:      buf[2] != 0,
		Pin:          buf[3],
		RateOK:       buf[4] != 0,
		LockCount:    buf[5],
		LossCount:    buf[6],
		SlipCount:    buf[7],
		HeaderErrors: binary.LittleEndian.Uint16(buf[8:10]),
		DetectedRate: binary.LittleEndian.Uint32(buf[12:16]),
		MeasuredHz:   binary.LittleEndian.Uint32(buf[16:20]),
	}, nil
}

// SetAdatInputEnable enables or disables the ADAT input.
// The firmware returns a PIN_CONFIG_* status byte.
func (d *Device) SetAdatInputEnable(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	wValue := uint16(0)
	if enabled {
		wValue = 1
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetAdatInputEnable, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_ADAT_INPUT_ENABLE: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_ADAT_INPUT_ENABLE: status 0x%02X", buf[0])
	}

	return nil
}

// GetAdatInputEnable returns the configured ADAT input enable state.
func (d *Device) GetAdatInputEnable() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatInputEnable, 0, vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_ADAT_INPUT_ENABLE: %w", err)
	}

	return buf[0] != 0, nil
}

// SetAdatInputPin sets the ADAT input RX GPIO pin (or 0xFF to clear).
// The firmware returns a PIN_CONFIG_* status byte.
func (d *Device) SetAdatInputPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetAdatInputPin, uint16(pin), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_ADAT_INPUT_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_ADAT_INPUT_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetAdatInputPin returns the configured ADAT input RX GPIO pin.
// Returns 0xFF when unset.
func (d *Device) GetAdatInputPin() (uint8, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatInputPin, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_ADAT_INPUT_PIN: %w", err)
	}

	return buf[0], nil
}

// SetAdatInputClockMode sets the ADAT input clock mode (0=master, 1=slave).
// The firmware returns a PIN_CONFIG_* status byte; the change is deferred.
func (d *Device) SetAdatInputClockMode(mode int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetAdatInputClockMode, uint16(mode), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_ADAT_INPUT_CLOCK_MODE: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_ADAT_INPUT_CLOCK_MODE: status 0x%02X", buf[0])
	}

	return nil
}

// GetAdatInputClockMode returns the live ADAT input clock mode.
func (d *Device) GetAdatInputClockMode() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAdatInputClockMode, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_ADAT_INPUT_CLOCK_MODE: %w", err)
	}

	return int(buf[0]), nil
}

// I2sSlaveState describes the I2S input external-clock lock state.
type I2sSlaveState int

const (
	I2sSlaveInactive  I2sSlaveState = 0 // not in slave role (or input stopped)
	I2sSlaveAcquiring I2sSlaveState = 1 // measuring external clocks, no lock yet
	I2sSlaveRelocking I2sSlaveState = 2 // clocks lost or rate changed; output muted
	I2sSlaveLocked    I2sSlaveState = 3 // locked to a supported external rate
)

// String returns a human-readable name for the I2S slave state.
func (s I2sSlaveState) String() string {
	switch s {
	case I2sSlaveInactive:
		return "inactive"
	case I2sSlaveAcquiring:
		return "acquiring"
	case I2sSlaveRelocking:
		return "relocking"
	case I2sSlaveLocked:
		return "locked"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// I2sSlaveStatus is the decoded 16-byte I2S slave status packet from
// REQ_GET_I2S_SLAVE_STATUS.
type I2sSlaveStatus struct {
	State        I2sSlaveState
	ClockMode    int
	LockCount    uint8
	LossCount    uint8
	DetectedRate uint32
	MeasuredHz   uint32
	SlipCount    uint8
}

// GetI2sSlaveStatus reads the 16-byte I2S slave input status packet.
func (d *Device) GetI2sSlaveStatus() (*I2sSlaveStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SSlaveStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_I2S_SLAVE_STATUS: %w", err)
	}
	if n < 16 {
		return nil, fmt.Errorf("REQ_GET_I2S_SLAVE_STATUS: response too short (%d bytes)", n)
	}

	return &I2sSlaveStatus{
		State:        I2sSlaveState(buf[0]),
		ClockMode:    int(buf[1]),
		LockCount:    buf[2],
		LossCount:    buf[3],
		DetectedRate: binary.LittleEndian.Uint32(buf[4:8]),
		MeasuredHz:   binary.LittleEndian.Uint32(buf[8:12]),
		SlipCount:    buf[12],
	}, nil
}

// SetI2SClockMode sets the I2S input clock mode (0 = master, 1 = slave;
// REQ_SET_I2S_CLOCK_MODE).  The change is deferred to the main loop; the
// GET returns the live mode until then.
func (d *Device) SetI2SClockMode(mode int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetI2SClockMode, 0, vendorInterface, []byte{byte(mode)})
	if err != nil {
		return fmt.Errorf("REQ_SET_I2S_CLOCK_MODE: %w", err)
	}

	return nil
}

// GetI2SClockMode returns the live I2S input clock mode (0 = master, 1 = slave).
func (d *Device) GetI2SClockMode() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetI2SClockMode, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_I2S_CLOCK_MODE: %w", err)
	}

	return int(buf[0]), nil
}
