package dspi

import (
	"encoding/binary"
	"fmt"
)

// SpdifRxStatus holds the decoded S/PDIF RX status.
type SpdifRxStatus struct {
	Raw     []byte
	Locked  bool
	Audio   bool
	Rate    uint32 // sample rate in Hz if available
	RateRaw uint32 // raw rate field from the device
}

// SpdifRxChannelStatus holds the S/PDIF channel status data (24 bytes).
type SpdifRxChannelStatus struct {
	Raw []byte
}

// SetSpdifRxPin sets the GPIO pin for S/PDIF RX input 1 (index 0).
func (d *Device) SetSpdifRxPin(pin int) error {
	return d.SetSpdifRxPinForIndex(0, pin)
}

// GetSpdifRxPin returns the current GPIO pin for S/PDIF RX input 1 (index 0).
func (d *Device) GetSpdifRxPin() (int, error) {
	return d.GetSpdifRxPinForIndex(0)
}

// SetSpdifRxPinForIndex sets the GPIO pin for a S/PDIF RX input (index 0..3).
// wValue is encoded as (index << 8) | pin; pass AdatInputPinUnset (0xFF) as
// the pin to restore that input's platform default.  The firmware returns a
// PIN_CONFIG_* status byte.
func (d *Device) SetSpdifRxPinForIndex(index, pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	wValue := uint16(index)<<8 | uint16(pin)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetSpdifRxPin, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_SPDIF_RX_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_SPDIF_RX_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetSpdifRxPinForIndex returns the GPIO pin for a S/PDIF RX input (index 0..3).
func (d *Device) GetSpdifRxPinForIndex(index int) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetSpdifRxPin, uint16(index), vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_SPDIF_RX_PIN: %w", err)
	}

	return int(buf[0]), nil
}

// SetSpdifInputEnable enables or disables an optional S/PDIF input
// (REQ_SET_SPDIF_INPUT_ENABLE; index 1..3).  Input 1 is always enabled;
// disabling an input that is the live or pending source is refused.  The
// firmware returns a PIN_CONFIG_* status byte.
func (d *Device) SetSpdifInputEnable(index int, enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	var enable uint16
	if enabled {
		enable = 1
	}

	buf := make([]byte, 1)
	wValue := uint16(index)<<8 | enable
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetSpdifInputEnable, wValue, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_SPDIF_INPUT_ENABLE: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_SPDIF_INPUT_ENABLE: status 0x%02X", buf[0])
	}

	return nil
}

// SpdifInputConfig describes the S/PDIF input inventory from
// REQ_GET_SPDIF_INPUT_CONFIG: the input count, an enable mask over ALL inputs
// (bit 0 = input 1, always set), and one GPIO per input.
type SpdifInputConfig struct {
	Count      int
	EnableMask uint8
	Pins       []uint8 // one GPIO per input (Count entries)
}

// GetSpdifInputConfig reads the S/PDIF input inventory.  Older firmware that
// predates the fourth input answers with a short payload; the count byte is
// authoritative and the pin list is read as short.
func (d *Device) GetSpdifInputConfig() (*SpdifInputConfig, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 6)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetSpdifInputConfig, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_SPDIF_INPUT_CONFIG: %w", err)
	}
	if n < 2 {
		return nil, fmt.Errorf("REQ_GET_SPDIF_INPUT_CONFIG: response too short (%d bytes)", n)
	}

	cfg := &SpdifInputConfig{
		Count:      int(buf[0]),
		EnableMask: buf[1],
	}
	if cfg.Count > n-2 {
		cfg.Count = n - 2
	}
	cfg.Pins = make([]uint8, cfg.Count)
	copy(cfg.Pins, buf[2:2+cfg.Count])

	return cfg, nil
}

// GetSpdifRxStatus reads the S/PDIF RX status and decodes it.
// The response contains:
//
//	byte 0: 0 = no lock, 1 = locked
//	byte 1: 0 = non-audio, 1 = audio
//	bytes 2-5: sample rate as uint32 (little-endian), 0 if unknown
func (d *Device) GetSpdifRxStatus() (*SpdifRxStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetSpdifRxStatus, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_SPDIF_RX_STATUS: %w", err)
	}

	status := &SpdifRxStatus{
		Raw: buf[:n],
	}

	if n >= 1 {
		status.Locked = buf[0] != 0
	}
	if n >= 2 {
		status.Audio = buf[1] != 0
	}
	if n >= 6 {
		status.RateRaw = binary.LittleEndian.Uint32(buf[2:6])
		// Valid sample rates are typically 32000, 44100, 48000, 88200, 96000, 176400, 192000
		if status.RateRaw >= 8000 && status.RateRaw <= 384000 {
			status.Rate = status.RateRaw
		}
	}

	return status, nil
}

// GetSpdifRxChannelStatus reads the S/PDIF RX channel status (24 bytes).
func (d *Device) GetSpdifRxChannelStatus() (*SpdifRxChannelStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 24)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetSpdifRxChStatus, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_SPDIF_RX_CH_STATUS: %w", err)
	}

	return &SpdifRxChannelStatus{Raw: buf[:n]}, nil
}
