package dspi

import (
	"encoding/binary"
	"fmt"
)

const (
	// safeMaxPayload is a generous upper bound for a WireBulkParams transfer.
	// The firmware payload is ~2896 bytes; this gives ample headroom.
	// Linux usbfs limits control transfers to 4096 bytes, so we stay within
	// that bound to avoid libusb: invalid param [code -2] on Linux hosts.
	safeMaxPayload = 4096
)

// BulkHeader is the 16-byte header parsed from a WireBulkParams payload.
type BulkHeader struct {
	FormatVersion uint8
	Platform      Platform
	NumChannels   int
	NumOutputs    int
	PayloadLength int
	FWMajor       uint16
	FWMinor       uint16
	FWPatch       uint16
}

// BulkParams holds a complete DSP state snapshot from the device.
type BulkParams struct {
	Header BulkHeader
	Raw    []byte // Full payload, suitable for restoration
}

// GetAllParams performs REQ_GET_ALL_PARAMS and returns the complete device state.
// It trusts the payload_length reported in the header rather than a hardcoded size.
func (d *Device) GetAllParams() (*BulkParams, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, safeMaxPayload)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAllParams, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_ALL_PARAMS: %w", err)
	}
	if n < 16 {
		return nil, fmt.Errorf("REQ_GET_ALL_PARAMS: response too short for header (got %d bytes)", n)
	}

	raw := append([]byte(nil), buf[:n]...)
	h := ParseBulkHeader(raw)

	return &BulkParams{
		Header: h,
		Raw:    raw,
	}, nil
}

// SetAllParams restores the complete device state via REQ_SET_ALL_PARAMS.
func (d *Device) SetAllParams(params *BulkParams) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	if params == nil || len(params.Raw) == 0 {
		return fmt.Errorf("no params to restore")
	}
	if params.Header.Platform != d.platform {
		return fmt.Errorf("platform mismatch (snapshot is %s, device is %s)", params.Header.Platform, d.platform)
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetAllParams, 0, vendorInterface, params.Raw)
	if err != nil {
		return fmt.Errorf("REQ_SET_ALL_PARAMS: %w", err)
	}

	return nil
}

// ParseBulkHeader parses a BulkHeader from the first 16 bytes of a raw payload.
func ParseBulkHeader(raw []byte) BulkHeader {
	return BulkHeader{
		FormatVersion: raw[0],
		Platform:      Platform(raw[1]),
		NumChannels:   int(raw[2]),
		NumOutputs:    int(raw[3]),
		PayloadLength: int(binary.LittleEndian.Uint16(raw[6:8])),
		FWMajor:       binary.LittleEndian.Uint16(raw[8:10]),
		FWMinor:       binary.LittleEndian.Uint16(raw[10:12]),
	}
}
