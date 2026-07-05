package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	// chunkSize is the maximum bytes per chunked control transfer, matching the
	// Linux usbfs 4096-byte limit (rounded down for safety).
	chunkSize = 4096

	// wireBulkSizeV16 is the V16 WireBulkParams payload size in bytes.
	wireBulkSizeV16 = 5864
)

// V16 field offsets within the WireBulkParams payload (5864 bytes).
const (
	fieldHeader       = 0    // 16 bytes
	fieldGlobal       = 16   // 16 bytes
	fieldCrossfeed    = 32   // 16 bytes
	fieldLegacy       = 48   // 16 bytes
	fieldDelays       = 64   // 68 bytes (17 × float32)
	fieldCrosspoints  = 132  // 576 bytes (8×9 × 8)
	fieldOutputs      = 708  // 108 bytes (9 × 12)
	fieldPins         = 816  // 8 bytes
	fieldEQ           = 824  // 3264 bytes (17×12 × 16)
	fieldChannelNames = 4088 // 544 bytes (17×32)
	fieldI2SConfig    = 4632 // 16 bytes
	fieldLeveller     = 4648 // 16 bytes
	fieldPreamp       = 4664 // 32 bytes (8 × float32)
	fieldMasterVolume = 4696 // 16 bytes
	fieldInputConfig  = 4712 // 16 bytes
	fieldLgSoundSync  = 4728 // 16 bytes
	fieldUserVolume   = 4744 // 16 bytes
	fieldDacHwMute    = 4760 // 16 bytes
	fieldCrossovers   = 4776 // 1088 bytes (17×4 × 16)
)

// fieldEntry describes a named field in the bulk payload by its offset and size.
type fieldEntry struct {
	Offset int
	Size   int
}

// fieldRegistry maps field names to their offset and size in the bulk payload.
var fieldRegistry = map[string]fieldEntry{
	"global":        {fieldGlobal, 16},
	"crossfeed":     {fieldCrossfeed, 16},
	"legacy":        {fieldLegacy, 16},
	"delays":        {fieldDelays, 68},
	"crosspoints":   {fieldCrosspoints, 576},
	"outputs":       {fieldOutputs, 108},
	"pins":          {fieldPins, 8},
	"eq":            {fieldEQ, 3264},
	"channel_names": {fieldChannelNames, 544},
	"i2s_config":    {fieldI2SConfig, 16},
	"leveller":      {fieldLeveller, 16},
	"preamp":        {fieldPreamp, 32},
	"master_volume": {fieldMasterVolume, 16},
	"input_config":  {fieldInputConfig, 16},
	"lg_sound_sync": {fieldLgSoundSync, 16},
	"user_volume":   {fieldUserVolume, 16},
	"dac_hw_mute":   {fieldDacHwMute, 16},
	"crossovers":    {fieldCrossovers, 1088},
}

// BulkHeader is the 16-byte header parsed from a WireBulkParams payload.
type BulkHeader struct {
	FormatVersion    uint8
	Platform         Platform
	NumChannels      int
	NumOutputs       int
	NumInputChannels int // NEW in V16 (byte 4)
	MaxBands         int
	PayloadLength    int
	FWMajor          uint16
	FWMinor          uint16
	FWPatch          uint16
}

// BulkParams holds a complete DSP state snapshot from the device.
type BulkParams struct {
	Header BulkHeader
	Raw    []byte // Full payload, suitable for restoration
}

// GetAllParams performs REQ_GET_ALL_PARAMS using chunked transfers (0xA2) and
// returns the complete device state. The firmware requires sequential offsets
// starting from 0; a STALL means restart from 0.
func (d *Device) GetAllParams() (*BulkParams, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	// Determine expected payload size from a minimal first-chunk response.
	// We request the first chunk (offset=0, 16 bytes) to read the header,
	// then calculate the full payload length.
	scratch := make([]byte, chunkSize)
	n, err := d.chunkedTransfer(vendorInterfaceInRequest, ReqGetAllParamsChunk, 0, 16, scratch)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_ALL_PARAMS: initial chunk: %w", err)
	}
	if n < 16 {
		return nil, fmt.Errorf("REQ_GET_ALL_PARAMS: response too short for header (got %d bytes)", n)
	}

	h := ParseBulkHeader(scratch[:n])
	payloadLen := h.PayloadLength
	if payloadLen < 16 || payloadLen > wireBulkSizeV16 {
		payloadLen = wireBulkSizeV16
	}

	// Allocate full payload buffer and copy the first chunk data.
	full := make([]byte, payloadLen)
	copy(full, scratch[:n])

	// Fetch remaining chunks sequentially, starting from where we left off.
	offset := n
	for offset < payloadLen {
		remaining := payloadLen - offset
		reqSize := min(remaining, chunkSize)

		n, err := d.chunkedTransfer(vendorInterfaceInRequest, ReqGetAllParamsChunk, uint16(offset), reqSize, scratch[:reqSize])
		if err != nil {
			return nil, fmt.Errorf("REQ_GET_ALL_PARAMS: chunk at offset %d: %w", offset, err)
		}

		copy(full[offset:], scratch[:n])
		offset += n
	}

	return &BulkParams{
		Header: h,
		Raw:    full,
	}, nil
}

// SetAllParams restores the complete device state via REQ_SET_ALL_PARAMS (0xA3 chunks).
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

	payload := params.Raw
	offset := 0
	for offset < len(payload) {
		remaining := len(payload) - offset
		segSize := min(remaining, chunkSize)

		_, err := d.chunkedTransfer(vendorInterfaceOutRequest, ReqSetAllParamsChunk, uint16(offset), segSize, payload[offset:offset+segSize])
		if err != nil {
			return fmt.Errorf("REQ_SET_ALL_PARAMS: chunk at offset %d: %w", offset, err)
		}

		offset += segSize
	}

	return nil
}

// chunkedTransfer performs a single chunk of a bulk GET or SET operation.
// It sends a control transfer with the given parameters and chunk metadata
// encoded as wValue=offset and data length = segSize.
func (d *Device) chunkedTransfer(bmRequestType, bRequest uint8, offset uint16, segSize int, data []byte) (int, error) {
	if segSize > len(data) {
		segSize = len(data)
	}

	outData := data
	if bmRequestType == vendorInterfaceOutRequest {
		// Use only the segment we want to send
		outData = data[:segSize]
	}

	return d.usb.ControlTransfer(bmRequestType, bRequest, offset, vendorInterface, outData)
}

// ParseBulkHeader parses a BulkHeader from the first 16 bytes of a raw payload.
func ParseBulkHeader(raw []byte) BulkHeader {
	payloadLen := 0
	if len(raw) >= 8 {
		payloadLen = int(binary.LittleEndian.Uint16(raw[6:8]))
	}

	return BulkHeader{
		FormatVersion:    raw[0],
		Platform:         Platform(raw[1]),
		NumChannels:      int(raw[2]),
		NumOutputs:       int(raw[3]),
		NumInputChannels: int(raw[4]), // V16: active input channels
		MaxBands:         int(raw[5]),
		PayloadLength:    payloadLen,
		FWMajor:          binary.LittleEndian.Uint16(raw[8:10]),
		FWMinor:          binary.LittleEndian.Uint16(raw[10:12]),
	}
}

// --- Field accessors for BulkParams ---

// validField checks that the named field exists and fits within the raw payload.
func (bp *BulkParams) validField(name string, dataSize int) (int, bool) {
	e, ok := fieldRegistry[name]
	if !ok {
		return 0, false
	}
	if e.Offset+dataSize > len(bp.Raw) {
		return 0, false
	}
	return e.Offset, true
}

// GetU8 reads a uint8 from a named field at the given sub-offset.
func (bp *BulkParams) GetU8(field string, subOffset int) (uint8, bool) {
	off, ok := bp.validField(field, subOffset+1)
	if !ok {
		return 0, false
	}
	return bp.Raw[off+subOffset], true
}

// SetU8 writes a uint8 to a named field at the given sub-offset.
func (bp *BulkParams) SetU8(field string, subOffset int, v uint8) {
	off, ok := bp.validField(field, subOffset+1)
	if !ok {
		return
	}
	bp.Raw[off+subOffset] = v
}

// GetU16 reads a little-endian uint16 from a named field at the given sub-offset.
func (bp *BulkParams) GetU16(field string, subOffset int) (uint16, bool) {
	off, ok := bp.validField(field, subOffset+2)
	if !ok {
		return 0, false
	}
	return binary.LittleEndian.Uint16(bp.Raw[off+subOffset:]), true
}

// SetU16 writes a little-endian uint16 to a named field at the given sub-offset.
func (bp *BulkParams) SetU16(field string, subOffset int, v uint16) {
	off, ok := bp.validField(field, subOffset+2)
	if !ok {
		return
	}
	binary.LittleEndian.PutUint16(bp.Raw[off+subOffset:], v)
}

// GetU32 reads a little-endian uint32 from a named field at the given sub-offset.
func (bp *BulkParams) GetU32(field string, subOffset int) (uint32, bool) {
	off, ok := bp.validField(field, subOffset+4)
	if !ok {
		return 0, false
	}
	return binary.LittleEndian.Uint32(bp.Raw[off+subOffset:]), true
}

// SetU32 writes a little-endian uint32 to a named field at the given sub-offset.
func (bp *BulkParams) SetU32(field string, subOffset int, v uint32) {
	off, ok := bp.validField(field, subOffset+4)
	if !ok {
		return
	}
	binary.LittleEndian.PutUint32(bp.Raw[off+subOffset:], v)
}

// GetFloat32 reads a float32 from a named field at the given sub-offset.
func (bp *BulkParams) GetFloat32(field string, subOffset int) (float32, bool) {
	off, ok := bp.validField(field, subOffset+4)
	if !ok {
		return 0, false
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(bp.Raw[off+subOffset:])), true
}

// SetFloat32 writes a float32 to a named field at the given sub-offset.
func (bp *BulkParams) SetFloat32(field string, subOffset int, v float32) {
	off, ok := bp.validField(field, subOffset+4)
	if !ok {
		return
	}
	binary.LittleEndian.PutUint32(bp.Raw[off+subOffset:], math.Float32bits(v))
}

// Convenience methods for the input_config field (offset 4712, 16 bytes V16).

// InputSource returns the active input source from the bulk parameters.
// Valid for V16 wire format.
func (bp *BulkParams) InputSource() (int, bool) {
	v, ok := bp.GetU8("input_config", 0)
	return int(v), ok
}

// SetInputSource updates the input source in the bulk parameters.
func (bp *BulkParams) SetInputSource(v int) {
	bp.SetU8("input_config", 0, uint8(v))
}

// I2SRxPin returns the I2S RX data pin (pair 0) from the bulk parameters.
func (bp *BulkParams) I2SRxPin() (int, bool) {
	v, ok := bp.GetU8("input_config", 2)
	return int(v), ok
}

// SetI2SRxPin updates the I2S RX data pin in the bulk parameters.
func (bp *BulkParams) SetI2SRxPin(v int) {
	bp.SetU8("input_config", 2, uint8(v))
}

// I2SInputRate returns the I2S input rate enum from the bulk parameters.
// 0=44100, 1=48000, 2=96000.
func (bp *BulkParams) I2SInputRate() (int, bool) {
	v, ok := bp.GetU8("input_config", 3)
	return int(v), ok
}

// SetI2SInputRate updates the I2S input rate enum in the bulk parameters.
func (bp *BulkParams) SetI2SInputRate(v int) {
	bp.SetU8("input_config", 3, uint8(v))
}

// I2SInputChannels returns the number of I2S input channels from the bulk parameters.
// Valid values: 0=absent, 2, 4, 6, 8.
func (bp *BulkParams) I2SInputChannels() (int, bool) {
	v, ok := bp.GetU8("input_config", 4)
	return int(v), ok
}

// SetI2SInputChannels updates the I2S input channel count in the bulk parameters.
func (bp *BulkParams) SetI2SInputChannels(v int) {
	bp.SetU8("input_config", 4, uint8(v))
}
