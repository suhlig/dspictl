package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

const (
	xoverBandBase    = 20
	maxXoverBands    = 4
	totalBandIndices = xoverBandBase + maxXoverBands // 24

	// peqActiveBands is the number of PEQ bands the current firmware accepts
	// via REQ_SET_EQ_PARAM. The bulk header reports MAX_BANDS (12), which is
	// storage depth; the firmware vendor handlers only allow 0..9 today.
	peqActiveBands = 10
)

// FilterType identifies the shape of an EQ band.
type FilterType int

const (
	FilterTypeFlat FilterType = iota
	FilterTypePeaking
	FilterTypeLowShelf
	FilterTypeHighShelf
	FilterTypeLowPass
	FilterTypeHighPass
)

// String returns the human-readable name of the filter type.
func (t FilterType) String() string {
	switch t {
	case FilterTypeFlat:
		return "flat"
	case FilterTypePeaking:
		return "peak"
	case FilterTypeLowShelf:
		return "lowshelf"
	case FilterTypeHighShelf:
		return "highshelf"
	case FilterTypeLowPass:
		return "lowpass"
	case FilterTypeHighPass:
		return "highpass"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ParseFilterType converts a string to a FilterType.
func ParseFilterType(s string) (FilterType, error) {
	switch strings.ToLower(s) {
	case "flat":
		return FilterTypeFlat, nil
	case "peak":
		return FilterTypePeaking, nil
	case "lowshelf":
		return FilterTypeLowShelf, nil
	case "highshelf":
		return FilterTypeHighShelf, nil
	case "lowpass":
		return FilterTypeLowPass, nil
	case "highpass":
		return FilterTypeHighPass, nil
	default:
		return 0, fmt.Errorf("unknown filter type: %s", s)
	}
}

// EQBand describes a single parametric EQ filter band.
type EQBand struct {
	Channel       int
	Band          int
	Type          FilterType
	Freq          float64 // Hz
	QualityFactor float64
	Gain          float64 // dB
}

// Validate checks that the band parameters are within acceptable ranges.
func (b *EQBand) Validate(maxChannel, maxBand int) error {
	if b.Channel < 0 || b.Channel > maxChannel {
		return fmt.Errorf("channel %d out of range (0-%d)", b.Channel, maxChannel)
	}

	err := validateBandIndex(b.Band, maxBand)
	if err != nil {
		return err
	}

	if b.Type != FilterTypeFlat && b.Freq <= 0 {
		return fmt.Errorf("frequency must be > 0")
	}

	if b.Type != FilterTypeFlat && b.QualityFactor <= 0 {
		return fmt.Errorf("quality factor must be > 0")
	}

	return nil
}

// validateBandIndex checks a band index against the firmware band map:
//
//	0..maxBand-1              = PEQ (valid)
//	maxBand..xoverBandBase-1  = out of range for PEQ
//	20..23                    = crossover (rejected — not yet supported)
//	24+                       = out of range overall
func validateBandIndex(band, maxBand int) error {
	if band < 0 || band >= totalBandIndices {
		return fmt.Errorf("band %d out of range (0-%d)", band, totalBandIndices-1)
	}

	if band < maxBand {
		return nil
	}

	if band >= xoverBandBase && band < xoverBandBase+maxXoverBands {
		return fmt.Errorf("band %d is a crossover band (%d-%d); crossover is not yet supported", band, xoverBandBase, xoverBandBase+maxXoverBands-1)
	}

	return fmt.Errorf("band %d out of range (0-%d)", band, maxBand-1)
}

// SetEQBand uploads a single EQ filter band to the device.
func (d *Device) SetEQBand(band *EQBand) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	maxChannel := 6
	if d.platform == PlatformRP2350 {
		maxChannel = 10
	}

	maxBand, err := d.MaxBands()
	if err != nil {
		return fmt.Errorf("getting max bands: %w", err)
	}

	err = band.Validate(maxChannel, maxBand)

	if err != nil {
		return fmt.Errorf("validating EQ band: %w", err)
	}

	buf := make([]byte, 16)
	buf[0] = byte(band.Channel)
	buf[1] = byte(band.Band)
	buf[2] = byte(band.Type)
	binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(float32(band.Freq)))
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(float32(band.QualityFactor)))
	binary.LittleEndian.PutUint32(buf[12:16], math.Float32bits(float32(band.Gain)))

	_, err = d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetEQParam, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_EQ_PARAM: %w", err)
	}

	return nil
}

// GetEQBand reads a single EQ filter band from the device.
// The firmware returns one parameter per request (4 bytes each), so this
// method performs four USB control transfers and assembles the result.
func (d *Device) GetEQBand(channel, band int) (*EQBand, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	maxBand, err := d.MaxBands()
	if err != nil {
		return nil, fmt.Errorf("getting max bands: %w", err)
	}

	err = validateBandIndex(band, maxBand)
	if err != nil {
		return nil, err
	}

	// Firmware packs wValue as: bits[15:8]=channel, bits[7:3]=band (5 bits),
	// bits[2:0]=param (0=type, 1=freq, 2=Q, 3=gain, 4=bypass).
	base := uint16(channel)<<8 | uint16(band)<<3

	// param 0: type (uint32)
	buf := make([]byte, 4)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM type: %w", err)
	}

	filterType := FilterType(binary.LittleEndian.Uint32(buf))

	// param 1: freq (float32)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base+1, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM freq: %w", err)
	}

	freq := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))

	// param 2: Q (float32)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base+2, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM Q: %w", err)
	}

	q := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))

	// param 3: gain (float32)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base+3, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM gain: %w", err)
	}

	gain := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))

	b := &EQBand{
		Channel:       channel,
		Band:          band,
		Type:          filterType,
		Freq:          freq,
		QualityFactor: q,
		Gain:          gain,
	}

	return b, nil
}

// SetMasterEQBypass enables or disables the master EQ bypass.
func (d *Device) SetMasterEQBypass(bypass bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)

	if bypass {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetBypass, 0, vendorInterface, []byte{val})

	if err != nil {
		return fmt.Errorf("REQ_SET_BYPASS: %w", err)
	}

	return nil
}

// GetMasterEQBypass reads the current master EQ bypass state.
func (d *Device) GetMasterEQBypass() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetBypass, 0, vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_BYPASS: %w", err)
	}

	return buf[0] != 0, nil
}

// MaxEQChannel returns the highest valid EQ channel index for the device's platform.
func (d *Device) MaxEQChannel() int {
	if d.platform == PlatformRP2350 {
		return 10
	}

	return 6
}

// MaxBands returns the number of active PEQ bands the device accepts.
// It reads the raw storage depth from the bulk header and caps it at the
// known firmware active count (10), since the firmware only allows bands
// 0..9 via REQ_SET_EQ_PARAM today.
func (d *Device) MaxBands() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	if d.maxBands > 0 {
		return d.maxBands, nil
	}

	bp, err := d.GetAllParams()
	if err != nil {
		return 0, fmt.Errorf("getting all params: %w", err)
	}

	raw := bp.Header.MaxBands

	if raw <= 0 {
		return 0, fmt.Errorf("invalid max bands reported by device: %d", raw)
	}

	d.maxBands = min(raw, peqActiveBands)

	return d.maxBands, nil
}

// SetBandBypass enables or disables bypass for a single EQ band.
func (d *Device) SetBandBypass(channel, band int, bypass bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	maxBand, err := d.MaxBands()
	if err != nil {
		return fmt.Errorf("getting max bands: %w", err)
	}

	err = validateBandIndex(band, maxBand)
	if err != nil {
		return err
	}

	var val byte
	if bypass {
		val = 1
	}

	wValue := uint16(channel)<<8 | uint16(band)
	_, err = d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetBandBypass, wValue, vendorInterface, []byte{val})

	if err != nil {
		return fmt.Errorf("REQ_SET_BAND_BYPASS: %w", err)
	}

	return nil
}

// GetBandBypass reads the bypass state of a single EQ band.
func (d *Device) GetBandBypass(channel, band int) (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	maxBand, err := d.MaxBands()
	if err != nil {
		return false, fmt.Errorf("getting max bands: %w", err)
	}

	err = validateBandIndex(band, maxBand)
	if err != nil {
		return false, err
	}

	wValue := uint16(channel)<<8 | uint16(band)
	buf := make([]byte, 1)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetBandBypass, wValue, vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_BAND_BYPASS: %w", err)
	}

	return buf[0] != 0, nil
}
