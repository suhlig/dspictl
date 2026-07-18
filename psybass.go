package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// PsybassParams holds the psychoacoustic bass configuration.
type PsybassParams struct {
	Enabled   bool
	Mask      uint16
	Cutoff    float32
	Harmonics float32
	Drive     float32
	Character float32
	Original  float32
}

// GetPsybass reads the psychoacoustic bass enable state and mask.
func (d *Device) GetPsybass() (bool, uint16, error) {
	if d.closed {
		return false, 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetPsybass, 0, vendorInterface, buf)
	if err != nil {
		return false, 0, fmt.Errorf("REQ_GET_PSYBASS: %w", err)
	}
	enabled := buf[0] != 0

	mask, err := d.GetPsybassMask()
	if err != nil {
		return false, 0, err
	}

	return enabled, mask, nil
}

// SetPsybass enables or disables psychoacoustic bass and sets its output mask.
func (d *Device) SetPsybass(enabled bool, mask uint16) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetPsybass, 0, vendorInterface, []byte{boolToByte(enabled)})
	if err != nil {
		return fmt.Errorf("REQ_SET_PSYBASS: %w", err)
	}

	if err := d.SetPsybassMask(mask); err != nil {
		return err
	}

	return nil
}

// GetPsybassMask returns the psychoacoustic bass output channel mask.
func (d *Device) GetPsybassMask() (uint16, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 2)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetPsybassMask, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_PSYBASS_MASK: %w", err)
	}

	return binary.LittleEndian.Uint16(buf), nil
}

// SetPsybassMask sets the psychoacoustic bass output channel mask.
func (d *Device) SetPsybassMask(mask uint16) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, mask)

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetPsybassMask, 0, vendorInterface, payload)
	if err != nil {
		return fmt.Errorf("REQ_SET_PSYBASS_MASK: %w", err)
	}

	return nil
}

// GetPsybassCutoff returns the speaker LF limit in Hz.
func (d *Device) GetPsybassCutoff() (float32, error) {
	return d.getPsybassFloat(ReqGetPsybassCutoff, "REQ_GET_PSYBASS_CUTOFF")
}

// SetPsybassCutoff sets the speaker LF limit in Hz.
func (d *Device) SetPsybassCutoff(v float32) error {
	return d.setPsybassFloat(ReqSetPsybassCutoff, "REQ_SET_PSYBASS_CUTOFF", v)
}

// GetPsybassHarmonics returns the harmonic mix level in dB.
func (d *Device) GetPsybassHarmonics() (float32, error) {
	return d.getPsybassFloat(ReqGetPsybassHarmonics, "REQ_GET_PSYBASS_HARMONICS")
}

// SetPsybassHarmonics sets the harmonic mix level in dB.
func (d *Device) SetPsybassHarmonics(v float32) error {
	return d.setPsybassFloat(ReqSetPsybassHarmonics, "REQ_SET_PSYBASS_HARMONICS", v)
}

// GetPsybassDrive returns the odd-path clipper drive in dB.
func (d *Device) GetPsybassDrive() (float32, error) {
	return d.getPsybassFloat(ReqGetPsybassDrive, "REQ_GET_PSYBASS_DRIVE")
}

// SetPsybassDrive sets the odd-path clipper drive in dB.
func (d *Device) SetPsybassDrive(v float32) error {
	return d.setPsybassFloat(ReqSetPsybassDrive, "REQ_SET_PSYBASS_DRIVE", v)
}

// GetPsybassCharacter returns the even/odd harmonic blend percentage.
func (d *Device) GetPsybassCharacter() (float32, error) {
	return d.getPsybassFloat(ReqGetPsybassCharacter, "REQ_GET_PSYBASS_CHARACTER")
}

// SetPsybassCharacter sets the even/odd harmonic blend percentage.
func (d *Device) SetPsybassCharacter(v float32) error {
	return d.setPsybassFloat(ReqSetPsybassCharacter, "REQ_SET_PSYBASS_CHARACTER", v)
}

// GetPsybassOriginal returns the original low-band level in dB.
func (d *Device) GetPsybassOriginal() (float32, error) {
	return d.getPsybassFloat(ReqGetPsybassOriginal, "REQ_GET_PSYBASS_ORIGINAL")
}

// SetPsybassOriginal sets the original low-band level in dB.
func (d *Device) SetPsybassOriginal(v float32) error {
	return d.setPsybassFloat(ReqSetPsybassOriginal, "REQ_SET_PSYBASS_ORIGINAL", v)
}

func (d *Device) getPsybassFloat(req uint8, label string) (float32, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, req, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", label, err)
	}

	return math.Float32frombits(binary.LittleEndian.Uint32(buf)), nil
}

func (d *Device) setPsybassFloat(req uint8, label string, v float32) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, math.Float32bits(v))

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, req, 0, vendorInterface, payload)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}

	return nil
}

func boolToByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}
