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

// SetSpdifRxPin sets the GPIO pin for S/PDIF RX input.
func (d *Device) SetSpdifRxPin(pin int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSetSpdifRxPin, uint16(pin), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_SPDIF_RX_PIN: %w", err)
	}

	if buf[0] != PinConfigSuccess {
		return fmt.Errorf("REQ_SET_SPDIF_RX_PIN: status 0x%02X", buf[0])
	}

	return nil
}

// GetSpdifRxPin returns the current GPIO pin for S/PDIF RX input.
func (d *Device) GetSpdifRxPin() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetSpdifRxPin, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_SPDIF_RX_PIN: %w", err)
	}

	return int(buf[0]), nil
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
