package dspi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/gousb"
)

const (
	// chunkSize is the maximum bytes per chunked control transfer, matching the
	// Linux usbfs 4096-byte limit (rounded down for safety).
	chunkSize = 4096

	// wireBulkSize is the V28 WireBulkParams payload size in bytes.
	// V28 (fourth selectable SPDIF input): WireInputConfig's spdif_rx_pin_ext grows
	// 2→3 entries, shifting the fields below it down one byte (section size unchanged).
	// V27 (upmixer centre OFF) and V26 (upmixer presence) are enum/byte-reuse only.
	// V25 (Stereo Upmixer) appends a 44-byte WireUpmixParams section at offset 5900.
	// V24 (ADAT input) reuses reserved bytes inside WireInputConfig; total size unchanged from V23.
	// V23 (Psychoacoustic Bass) appends a 24-byte WirePsybassParams section at offset 5876.
	// V20 adds: loudness_output_mask in global (V19), crossfeed output_pair_mask (V20),
	// leveller detector/apply masks growing from 16→20 bytes (V18), and ADAT output section (V17).
	wireBulkSize = 5944
)

// V28 field offsets within the WireBulkParams payload (5944 bytes).
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
	fieldLeveller     = 4648 // 20 bytes (V18: +detector/apply masks at +16/+17)
	fieldPreamp       = 4668 // 32 bytes (8 × float32)
	fieldMasterVolume = 4700 // 16 bytes
	fieldInputConfig  = 4716 // 16 bytes
	fieldLgSoundSync  = 4732 // 16 bytes
	fieldUserVolume   = 4748 // 16 bytes
	fieldDacHwMute    = 4764 // 16 bytes
	fieldCrossovers   = 4780 // 1088 bytes (17×4 × 16)
	fieldADAT         = 5868 // 8 bytes (ADAT output)
	fieldPsybass      = 5876 // 24 bytes (V23)
	fieldUpmix        = 5900 // 44 bytes (V25: stereo upmixer)
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
	"leveller":      {fieldLeveller, 20},
	"preamp":        {fieldPreamp, 32},
	"master_volume": {fieldMasterVolume, 16},
	"input_config":  {fieldInputConfig, 16},
	"lg_sound_sync": {fieldLgSoundSync, 16},
	"user_volume":   {fieldUserVolume, 16},
	"dac_hw_mute":   {fieldDacHwMute, 16},
	"crossovers":    {fieldCrossovers, 1088},
	"adat":          {fieldADAT, 8},
	"psybass":       {fieldPsybass, 24},
	"upmix":         {fieldUpmix, 44},
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

// bulkRetryBackoff is the delay between retries of a bulk transfer that the
// firmware STALLed.  The firmware refuses bulk access while the main-loop
// apply of a previous SET is still running (bulk_params_pending) or another
// transport owns the shared bulk buffer; the Console retries the same way
// (Commands.swift fetchAllParamsRetrying, 0.15/0.3/0.6 s).
var bulkRetryBackoff = []time.Duration{150 * time.Millisecond, 300 * time.Millisecond, 600 * time.Millisecond}

// isBulkStall reports whether err is a USB STALL (pipe error), the
// firmware's "busy, retry" signal for the shared bulk buffer.
func isBulkStall(err error) bool {
	var usbErr gousb.Error
	return errors.As(err, &usbErr) && usbErr == gousb.ErrorPipe
}

// GetAllParams performs REQ_GET_ALL_PARAMS using chunked transfers (0xA2) and
// returns the complete device state.  The firmware requires sequential offsets
// starting from 0; a STALL means restart from 0.  STALLs are retried with
// backoff because the firmware refuses bulk access while the main-loop apply
// of a previous SET is still running.
func (d *Device) GetAllParams() (*BulkParams, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	for attempt := 0; ; attempt++ {
		bp, err := d.getAllParamsOnce()
		if err == nil {
			return bp, nil
		}
		if !isBulkStall(err) || attempt >= len(bulkRetryBackoff) {
			return nil, err
		}
		time.Sleep(bulkRetryBackoff[attempt])
	}
}

func (d *Device) getAllParamsOnce() (*BulkParams, error) {
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
	if payloadLen < 16 || payloadLen > wireBulkSize {
		payloadLen = wireBulkSize
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
// STALLs are retried with backoff (see GetAllParams); a mid-sequence STALL
// restarts the whole upload from offset 0, which the firmware accepts.
func (d *Device) SetAllParams(params *BulkParams) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	if params == nil || len(params.Raw) == 0 {
		return fmt.Errorf("no params to restore")
	}
	// The firmware only accepts the exact current wire size (v1.1.5 = V28,
	// 5944 bytes); anything else is rejected with a STALL.  Fail early with a
	// clear message instead (e.g. importing an older 5900-byte export).
	if len(params.Raw) != wireBulkSize {
		return fmt.Errorf("snapshot is %d bytes, device expects %d (wire V28)", len(params.Raw), wireBulkSize)
	}
	if params.Header.Platform != d.platform {
		return fmt.Errorf("platform mismatch (snapshot is %s, device is %s)", params.Header.Platform, d.platform)
	}

	for attempt := 0; ; attempt++ {
		err := d.setAllParamsOnce(params)
		if err == nil {
			return nil
		}
		if !isBulkStall(err) || attempt >= len(bulkRetryBackoff) {
			if isBulkStall(err) {
				return fmt.Errorf("%w (device stayed busy after %d attempts; is another host or transport using it?)", err, attempt+1)
			}
			return err
		}
		time.Sleep(bulkRetryBackoff[attempt])
	}
}

func (d *Device) setAllParamsOnce(params *BulkParams) error {
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

// LoudnessOutputMask returns the loudness output bitmask from the bulk parameters.
// Bit k = loudness compensates output channel k (V19+; default 0xFFFF = all outputs).
func (bp *BulkParams) LoudnessOutputMask() (uint16, bool) {
	return bp.GetU16("global", 6)
}

// SetLoudnessOutputMask updates the loudness output bitmask in the bulk parameters.
func (bp *BulkParams) SetLoudnessOutputMask(v uint16) {
	bp.SetU16("global", 6, v)
}

// LoudnessRefSPL returns the loudness reference SPL in dB from the bulk parameters.
func (bp *BulkParams) LoudnessRefSPL() (float32, bool) {
	return bp.GetFloat32("global", 8)
}

// SetLoudnessRefSPL updates the loudness reference SPL in dB in the bulk parameters.
func (bp *BulkParams) SetLoudnessRefSPL(v float32) {
	bp.SetFloat32("global", 8, v)
}

// LoudnessIntensityPct returns the loudness compensation intensity percentage
// from the bulk parameters.
func (bp *BulkParams) LoudnessIntensityPct() (float32, bool) {
	return bp.GetFloat32("global", 12)
}

// SetLoudnessIntensityPct updates the loudness compensation intensity percentage
// in the bulk parameters.
func (bp *BulkParams) SetLoudnessIntensityPct(v float32) {
	bp.SetFloat32("global", 12, v)
}

// CrossfeedOutputPairMask returns the crossfeed output-pair bitmask from the bulk parameters.
// Bit p = crossfeed runs on output pair p (outputs 2p / 2p+1) (V20+; default 0x01).
func (bp *BulkParams) CrossfeedOutputPairMask() (uint8, bool) {
	return bp.GetU8("crossfeed", 3)
}

// SetCrossfeedOutputPairMask updates the crossfeed output-pair bitmask in the bulk parameters.
func (bp *BulkParams) SetCrossfeedOutputPairMask(v uint8) {
	bp.SetU8("crossfeed", 3, v)
}

// LevellerDetectorMask returns the leveller detector channel mask from the bulk parameters.
// Bit k = input channel k feeds the shared RMS detector (V18+).
func (bp *BulkParams) LevellerDetectorMask() (uint8, bool) {
	return bp.GetU8("leveller", 16)
}

// SetLevellerDetectorMask updates the leveller detector channel mask in the bulk parameters.
func (bp *BulkParams) SetLevellerDetectorMask(v uint8) {
	bp.SetU8("leveller", 16, v)
}

// LevellerApplyMask returns the leveller apply channel mask from the bulk parameters.
// Bit k = gain is applied to input channel k (V18+).
func (bp *BulkParams) LevellerApplyMask() (uint8, bool) {
	return bp.GetU8("leveller", 17)
}

// SetLevellerApplyMask updates the leveller apply channel mask in the bulk parameters.
func (bp *BulkParams) SetLevellerApplyMask(v uint8) {
	bp.SetU8("leveller", 17, v)
}

// SpdifRxPinExt returns the optional SPDIF input RX GPIO (index 1..3; V28+).
// 0 means absent (keep live value).  Index 0 addresses the primary input,
// which lives in the separate spdif_rx_pin field.
func (bp *BulkParams) SpdifRxPinExt(index int) (uint8, bool) {
	if index < 1 || index > 3 {
		return 0, false
	}
	return bp.GetU8("input_config", 7+index)
}

// SetSpdifRxPinExt updates the optional SPDIF input RX GPIO (index 1..3; V28+).
func (bp *BulkParams) SetSpdifRxPinExt(index int, v uint8) {
	if index < 1 || index > 3 {
		return
	}
	bp.SetU8("input_config", 7+index, v)
}

// SpdifRxEnabledExtP1 returns the optional SPDIF inputs enable mask +1 encoded
// field from the bulk parameters (V28+).  0 = absent (keep live), 1 = all
// disabled, 2 = SPDIF2, 3 = SPDIF2+3, 4 = SPDIF2+3+4.
func (bp *BulkParams) SpdifRxEnabledExtP1() (uint8, bool) {
	return bp.GetU8("input_config", 11)
}

// SetSpdifRxEnabledExtP1 updates the optional SPDIF inputs enable mask +1 field.
func (bp *BulkParams) SetSpdifRxEnabledExtP1(v uint8) {
	bp.SetU8("input_config", 11, v)
}

// I2SClockMode returns the I2S clock mode from the bulk parameters (V21+).
// 0 = master, 1 = slave.
func (bp *BulkParams) I2SClockMode() (uint8, bool) {
	return bp.GetU8("input_config", 12)
}

// SetI2SClockMode updates the I2S clock mode in the bulk parameters.
func (bp *BulkParams) SetI2SClockMode(v uint8) {
	bp.SetU8("input_config", 12, v)
}

// AdatInputPin returns the configured ADAT input RX GPIO from the bulk parameters (V28+).
// 0 means absent (keep live value), 0xFF is never stored on the wire.
func (bp *BulkParams) AdatInputPin() (uint8, bool) {
	return bp.GetU8("input_config", 13)
}

// SetAdatInputPin updates the ADAT input RX GPIO in the bulk parameters.
func (bp *BulkParams) SetAdatInputPin(v uint8) {
	bp.SetU8("input_config", 13, v)
}

// AdatInputEnabledP1 returns the ADAT input enable +1 encoded field from the bulk parameters.
// 0 = absent (keep live), 1 = disabled, 2 = enabled.
func (bp *BulkParams) AdatInputEnabledP1() (uint8, bool) {
	return bp.GetU8("input_config", 14)
}

// SetAdatInputEnabledP1 updates the ADAT input enable +1 encoded field in the bulk parameters.
func (bp *BulkParams) SetAdatInputEnabledP1(v uint8) {
	bp.SetU8("input_config", 14, v)
}

// AdatInputClockModeP1 returns the ADAT input clock mode +1 encoded field from the bulk parameters.
// 0 = absent (keep live), 1 = master, 2 = slave.
func (bp *BulkParams) AdatInputClockModeP1() (uint8, bool) {
	return bp.GetU8("input_config", 15)
}

// SetAdatInputClockModeP1 updates the ADAT input clock mode +1 encoded field in the bulk parameters.
func (bp *BulkParams) SetAdatInputClockModeP1(v uint8) {
	bp.SetU8("input_config", 15, v)
}

// PsybassEnabled returns the psychoacoustic bass enabled flag from the bulk parameters.
func (bp *BulkParams) PsybassEnabled() (bool, bool) {
	v, ok := bp.GetU8("psybass", 0)
	return v != 0, ok
}

// SetPsybassEnabled updates the psychoacoustic bass enabled flag in the bulk parameters.
func (bp *BulkParams) SetPsybassEnabled(v bool) {
	val := uint8(0)
	if v {
		val = 1
	}
	bp.SetU8("psybass", 0, val)
}

// PsybassOutputMask returns the psychoacoustic bass output channel mask from the bulk parameters.
// Bit k = psybass processes output channel k.
func (bp *BulkParams) PsybassOutputMask() (uint16, bool) {
	return bp.GetU16("psybass", 2)
}

// SetPsybassOutputMask updates the psychoacoustic bass output channel mask in the bulk parameters.
func (bp *BulkParams) SetPsybassOutputMask(v uint16) {
	bp.SetU16("psybass", 2, v)
}

// PsybassCutoff returns the psychoacoustic bass cutoff frequency in Hz.
func (bp *BulkParams) PsybassCutoff() (float32, bool) {
	return bp.GetFloat32("psybass", 4)
}

// SetPsybassCutoff updates the psychoacoustic bass cutoff frequency in Hz.
func (bp *BulkParams) SetPsybassCutoff(v float32) {
	bp.SetFloat32("psybass", 4, v)
}

// PsybassHarmonics returns the psychoacoustic bass harmonics mix level in dB.
func (bp *BulkParams) PsybassHarmonics() (float32, bool) {
	return bp.GetFloat32("psybass", 8)
}

// SetPsybassHarmonics updates the psychoacoustic bass harmonics mix level in dB.
func (bp *BulkParams) SetPsybassHarmonics(v float32) {
	bp.SetFloat32("psybass", 8, v)
}

// PsybassDrive returns the psychoacoustic bass drive level in dB.
func (bp *BulkParams) PsybassDrive() (float32, bool) {
	return bp.GetFloat32("psybass", 12)
}

// SetPsybassDrive updates the psychoacoustic bass drive level in dB.
func (bp *BulkParams) SetPsybassDrive(v float32) {
	bp.SetFloat32("psybass", 12, v)
}

// PsybassCharacter returns the psychoacoustic bass even/odd harmonic blend percentage.
func (bp *BulkParams) PsybassCharacter() (float32, bool) {
	return bp.GetFloat32("psybass", 16)
}

// SetPsybassCharacter updates the psychoacoustic bass even/odd harmonic blend percentage.
func (bp *BulkParams) SetPsybassCharacter(v float32) {
	bp.SetFloat32("psybass", 16, v)
}

// PsybassOriginal returns the psychoacoustic bass original low-band level in dB.
func (bp *BulkParams) PsybassOriginal() (float32, bool) {
	return bp.GetFloat32("psybass", 20)
}

// SetPsybassOriginal updates the psychoacoustic bass original low-band level in dB.
func (bp *BulkParams) SetPsybassOriginal(v float32) {
	bp.SetFloat32("psybass", 20, v)
}

// UpmixEnabled returns the stereo upmixer enabled flag from the bulk parameters (V25+).
func (bp *BulkParams) UpmixEnabled() (bool, bool) {
	v, ok := bp.GetU8("upmix", 0)
	return v != 0, ok
}

// SetUpmixEnabled updates the stereo upmixer enabled flag in the bulk parameters.
func (bp *BulkParams) SetUpmixEnabled(v bool) {
	val := uint8(0)
	if v {
		val = 1
	}
	bp.SetU8("upmix", 0, val)
}

// UpmixCenterMode returns the upmixer centre engine mode from the bulk parameters.
// 0 = passive, 1 = adaptive, 2 = off (V27+).
func (bp *BulkParams) UpmixCenterMode() (uint8, bool) {
	return bp.GetU8("upmix", 1)
}

// SetUpmixCenterMode updates the upmixer centre engine mode in the bulk parameters.
func (bp *BulkParams) SetUpmixCenterMode(v uint8) {
	bp.SetU8("upmix", 1, v)
}

// UpmixSurroundMode returns the upmixer surround engine mode from the bulk parameters.
// 0 = off, 1 = passive, 2 = adaptive.
func (bp *BulkParams) UpmixSurroundMode() (uint8, bool) {
	return bp.GetU8("upmix", 2)
}

// SetUpmixSurroundMode updates the upmixer surround engine mode in the bulk parameters.
func (bp *BulkParams) SetUpmixSurroundMode(v uint8) {
	bp.SetU8("upmix", 2, v)
}

// UpmixPresenceQ1 returns the upmixer centre presence bell gain as the raw wire
// int8 (dB × 2, V26+).  Convert with float32(v) / 2 for dB.
func (bp *BulkParams) UpmixPresenceQ1() (int8, bool) {
	v, ok := bp.GetU8("upmix", 3)
	return int8(v), ok
}

// SetUpmixPresenceQ1 updates the upmixer centre presence bell gain (wire int8,
// dB × 2, V26+).
func (bp *BulkParams) SetUpmixPresenceQ1(v int8) {
	bp.SetU8("upmix", 3, uint8(v))
}

// UpmixStrengthPct returns the upmixer strength percentage from the bulk parameters.
func (bp *BulkParams) UpmixStrengthPct() (float32, bool) {
	return bp.GetFloat32("upmix", 4)
}

// SetUpmixStrengthPct updates the upmixer strength percentage in the bulk parameters.
func (bp *BulkParams) SetUpmixStrengthPct(v float32) {
	bp.SetFloat32("upmix", 4, v)
}

// UpmixCenterWidthPct returns the upmixer centre width percentage from the bulk parameters.
func (bp *BulkParams) UpmixCenterWidthPct() (float32, bool) {
	return bp.GetFloat32("upmix", 8)
}

// SetUpmixCenterWidthPct updates the upmixer centre width percentage in the bulk parameters.
func (bp *BulkParams) SetUpmixCenterWidthPct(v float32) {
	bp.SetFloat32("upmix", 8, v)
}

// UpmixCorrThresholdPct returns the upmixer correlation threshold percentage
// from the bulk parameters.
func (bp *BulkParams) UpmixCorrThresholdPct() (float32, bool) {
	return bp.GetFloat32("upmix", 12)
}

// SetUpmixCorrThresholdPct updates the upmixer correlation threshold percentage
// in the bulk parameters.
func (bp *BulkParams) SetUpmixCorrThresholdPct(v float32) {
	bp.SetFloat32("upmix", 12, v)
}

// UpmixAttackMs returns the upmixer attack time in ms from the bulk parameters.
func (bp *BulkParams) UpmixAttackMs() (float32, bool) {
	return bp.GetFloat32("upmix", 16)
}

// SetUpmixAttackMs updates the upmixer attack time in ms in the bulk parameters.
func (bp *BulkParams) SetUpmixAttackMs(v float32) {
	bp.SetFloat32("upmix", 16, v)
}

// UpmixReleaseMs returns the upmixer release time in ms from the bulk parameters.
func (bp *BulkParams) UpmixReleaseMs() (float32, bool) {
	return bp.GetFloat32("upmix", 20)
}

// SetUpmixReleaseMs updates the upmixer release time in ms in the bulk parameters.
func (bp *BulkParams) SetUpmixReleaseMs(v float32) {
	bp.SetFloat32("upmix", 20, v)
}

// UpmixDetectorHpfHz returns the upmixer detector high-pass frequency from the
// bulk parameters.
func (bp *BulkParams) UpmixDetectorHpfHz() (float32, bool) {
	return bp.GetFloat32("upmix", 24)
}

// SetUpmixDetectorHpfHz updates the upmixer detector high-pass frequency in the
// bulk parameters.
func (bp *BulkParams) SetUpmixDetectorHpfHz(v float32) {
	bp.SetFloat32("upmix", 24, v)
}

// UpmixSurroundDelayMs returns the upmixer surround delay in ms from the bulk parameters.
func (bp *BulkParams) UpmixSurroundDelayMs() (float32, bool) {
	return bp.GetFloat32("upmix", 28)
}

// SetUpmixSurroundDelayMs updates the upmixer surround delay in ms in the bulk parameters.
func (bp *BulkParams) SetUpmixSurroundDelayMs(v float32) {
	bp.SetFloat32("upmix", 28, v)
}

// UpmixSurroundHpfHz returns the upmixer surround high-pass frequency from the
// bulk parameters.
func (bp *BulkParams) UpmixSurroundHpfHz() (float32, bool) {
	return bp.GetFloat32("upmix", 32)
}

// SetUpmixSurroundHpfHz updates the upmixer surround high-pass frequency in the
// bulk parameters.
func (bp *BulkParams) SetUpmixSurroundHpfHz(v float32) {
	bp.SetFloat32("upmix", 32, v)
}

// UpmixSurroundLpfHz returns the upmixer surround low-pass frequency from the
// bulk parameters.
func (bp *BulkParams) UpmixSurroundLpfHz() (float32, bool) {
	return bp.GetFloat32("upmix", 36)
}

// SetUpmixSurroundLpfHz updates the upmixer surround low-pass frequency in the
// bulk parameters.
func (bp *BulkParams) SetUpmixSurroundLpfHz(v float32) {
	bp.SetFloat32("upmix", 36, v)
}

// UpmixDecorrPct returns the upmixer decorrelation percentage from the bulk parameters.
func (bp *BulkParams) UpmixDecorrPct() (float32, bool) {
	return bp.GetFloat32("upmix", 40)
}

// SetUpmixDecorrPct updates the upmixer decorrelation percentage in the bulk parameters.
func (bp *BulkParams) SetUpmixDecorrPct(v float32) {
	bp.SetFloat32("upmix", 40, v)
}
