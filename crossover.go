package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// CrossoverFilterType identifies a crossover filter family, order, and shape.
type CrossoverFilterType int

const (
	CrossoverTypeLR2LP  CrossoverFilterType = 32 // Linkwitz-Riley 2nd-order low-pass
	CrossoverTypeLR2HP  CrossoverFilterType = 33 // Linkwitz-Riley 2nd-order high-pass
	CrossoverTypeLR4LP  CrossoverFilterType = 34 // Linkwitz-Riley 4th-order low-pass
	CrossoverTypeLR4HP  CrossoverFilterType = 35 // Linkwitz-Riley 4th-order high-pass
	CrossoverTypeLR6LP  CrossoverFilterType = 36 // Linkwitz-Riley 6th-order low-pass
	CrossoverTypeLR6HP  CrossoverFilterType = 37 // Linkwitz-Riley 6th-order high-pass
	CrossoverTypeLR8LP  CrossoverFilterType = 38 // Linkwitz-Riley 8th-order low-pass
	CrossoverTypeLR8HP  CrossoverFilterType = 39 // Linkwitz-Riley 8th-order high-pass
	CrossoverTypeBW1LP  CrossoverFilterType = 40 // Butterworth 1st-order low-pass
	CrossoverTypeBW1HP  CrossoverFilterType = 41 // Butterworth 1st-order high-pass
	CrossoverTypeBW2LP  CrossoverFilterType = 42 // Butterworth 2nd-order low-pass
	CrossoverTypeBW2HP  CrossoverFilterType = 43 // Butterworth 2nd-order high-pass
	CrossoverTypeBW3LP  CrossoverFilterType = 44 // Butterworth 3rd-order low-pass
	CrossoverTypeBW3HP  CrossoverFilterType = 45 // Butterworth 3rd-order high-pass
	CrossoverTypeBW4LP  CrossoverFilterType = 46 // Butterworth 4th-order low-pass
	CrossoverTypeBW4HP  CrossoverFilterType = 47 // Butterworth 4th-order high-pass
	CrossoverTypeBW5LP  CrossoverFilterType = 48 // Butterworth 5th-order low-pass
	CrossoverTypeBW5HP  CrossoverFilterType = 49 // Butterworth 5th-order high-pass
	CrossoverTypeBW6LP  CrossoverFilterType = 50 // Butterworth 6th-order low-pass
	CrossoverTypeBW6HP  CrossoverFilterType = 51 // Butterworth 6th-order high-pass
	CrossoverTypeBW7LP  CrossoverFilterType = 52 // Butterworth 7th-order low-pass
	CrossoverTypeBW7HP  CrossoverFilterType = 53 // Butterworth 7th-order high-pass
	CrossoverTypeBW8LP  CrossoverFilterType = 54 // Butterworth 8th-order low-pass
	CrossoverTypeBW8HP  CrossoverFilterType = 55 // Butterworth 8th-order high-pass
	CrossoverTypeBES2LP CrossoverFilterType = 56 // Bessel 2nd-order low-pass
	CrossoverTypeBES2HP CrossoverFilterType = 57 // Bessel 2nd-order high-pass
	CrossoverTypeBES4LP CrossoverFilterType = 58 // Bessel 4th-order low-pass
	CrossoverTypeBES4HP CrossoverFilterType = 59 // Bessel 4th-order high-pass
	CrossoverTypeBES6LP CrossoverFilterType = 60 // Bessel 6th-order low-pass
	CrossoverTypeBES6HP CrossoverFilterType = 61 // Bessel 6th-order high-pass
	CrossoverTypeBES8LP CrossoverFilterType = 62 // Bessel 8th-order low-pass
	CrossoverTypeBES8HP CrossoverFilterType = 63 // Bessel 8th-order high-pass
)

// String returns the human-readable name of the crossover filter type.
func (t CrossoverFilterType) String() string {
	switch t {
	case CrossoverTypeLR2LP:
		return "lr2-lp"
	case CrossoverTypeLR2HP:
		return "lr2-hp"
	case CrossoverTypeLR4LP:
		return "lr4-lp"
	case CrossoverTypeLR4HP:
		return "lr4-hp"
	case CrossoverTypeLR6LP:
		return "lr6-lp"
	case CrossoverTypeLR6HP:
		return "lr6-hp"
	case CrossoverTypeLR8LP:
		return "lr8-lp"
	case CrossoverTypeLR8HP:
		return "lr8-hp"
	case CrossoverTypeBW1LP:
		return "bw1-lp"
	case CrossoverTypeBW1HP:
		return "bw1-hp"
	case CrossoverTypeBW2LP:
		return "bw2-lp"
	case CrossoverTypeBW2HP:
		return "bw2-hp"
	case CrossoverTypeBW3LP:
		return "bw3-lp"
	case CrossoverTypeBW3HP:
		return "bw3-hp"
	case CrossoverTypeBW4LP:
		return "bw4-lp"
	case CrossoverTypeBW4HP:
		return "bw4-hp"
	case CrossoverTypeBW5LP:
		return "bw5-lp"
	case CrossoverTypeBW5HP:
		return "bw5-hp"
	case CrossoverTypeBW6LP:
		return "bw6-lp"
	case CrossoverTypeBW6HP:
		return "bw6-hp"
	case CrossoverTypeBW7LP:
		return "bw7-lp"
	case CrossoverTypeBW7HP:
		return "bw7-hp"
	case CrossoverTypeBW8LP:
		return "bw8-lp"
	case CrossoverTypeBW8HP:
		return "bw8-hp"
	case CrossoverTypeBES2LP:
		return "bes2-lp"
	case CrossoverTypeBES2HP:
		return "bes2-hp"
	case CrossoverTypeBES4LP:
		return "bes4-lp"
	case CrossoverTypeBES4HP:
		return "bes4-hp"
	case CrossoverTypeBES6LP:
		return "bes6-lp"
	case CrossoverTypeBES6HP:
		return "bes6-hp"
	case CrossoverTypeBES8LP:
		return "bes8-lp"
	case CrossoverTypeBES8HP:
		return "bes8-hp"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ParseCrossoverFilterType converts a string to a CrossoverFilterType.
func ParseCrossoverFilterType(s string) (CrossoverFilterType, error) {
	switch strings.ToLower(s) {
	case "lr2-lp":
		return CrossoverTypeLR2LP, nil
	case "lr2-hp":
		return CrossoverTypeLR2HP, nil
	case "lr4-lp":
		return CrossoverTypeLR4LP, nil
	case "lr4-hp":
		return CrossoverTypeLR4HP, nil
	case "lr6-lp":
		return CrossoverTypeLR6LP, nil
	case "lr6-hp":
		return CrossoverTypeLR6HP, nil
	case "lr8-lp":
		return CrossoverTypeLR8LP, nil
	case "lr8-hp":
		return CrossoverTypeLR8HP, nil
	case "bw1-lp":
		return CrossoverTypeBW1LP, nil
	case "bw1-hp":
		return CrossoverTypeBW1HP, nil
	case "bw2-lp":
		return CrossoverTypeBW2LP, nil
	case "bw2-hp":
		return CrossoverTypeBW2HP, nil
	case "bw3-lp":
		return CrossoverTypeBW3LP, nil
	case "bw3-hp":
		return CrossoverTypeBW3HP, nil
	case "bw4-lp":
		return CrossoverTypeBW4LP, nil
	case "bw4-hp":
		return CrossoverTypeBW4HP, nil
	case "bw5-lp":
		return CrossoverTypeBW5LP, nil
	case "bw5-hp":
		return CrossoverTypeBW5HP, nil
	case "bw6-lp":
		return CrossoverTypeBW6LP, nil
	case "bw6-hp":
		return CrossoverTypeBW6HP, nil
	case "bw7-lp":
		return CrossoverTypeBW7LP, nil
	case "bw7-hp":
		return CrossoverTypeBW7HP, nil
	case "bw8-lp":
		return CrossoverTypeBW8LP, nil
	case "bw8-hp":
		return CrossoverTypeBW8HP, nil
	case "bes2-lp":
		return CrossoverTypeBES2LP, nil
	case "bes2-hp":
		return CrossoverTypeBES2HP, nil
	case "bes4-lp":
		return CrossoverTypeBES4LP, nil
	case "bes4-hp":
		return CrossoverTypeBES4HP, nil
	case "bes6-lp":
		return CrossoverTypeBES6LP, nil
	case "bes6-hp":
		return CrossoverTypeBES6HP, nil
	case "bes8-lp":
		return CrossoverTypeBES8LP, nil
	case "bes8-hp":
		return CrossoverTypeBES8HP, nil
	default:
		return 0, fmt.Errorf("unknown crossover filter type: %s", s)
	}
}

// CrossoverBand describes a single crossover filter band.
type CrossoverBand struct {
	Channel int
	Band    int // 20-23 (local band 0-3)
	Type    CrossoverFilterType
	Freq    float64 // Hz
	Bypass  bool
}

// Validate checks that the crossover band parameters are within acceptable ranges.
func (b *CrossoverBand) Validate(maxChannel int) error {
	if b.Channel < 0 || b.Channel > maxChannel {
		return fmt.Errorf("channel %d out of range (0-%d)", b.Channel, maxChannel)
	}

	// Crossover is only allowed on output channels (channel >= CH_OUT_1 = 2)
	if b.Channel < 2 {
		return fmt.Errorf("crossover is not supported on master channels (channel %d)", b.Channel)
	}

	if err := validateCrossoverBand(b.Band); err != nil {
		return err
	}

	if b.Type < CrossoverTypeLR2LP || b.Type > CrossoverTypeBES8HP {
		return fmt.Errorf("invalid crossover type: %d", b.Type)
	}

	if b.Freq <= 0 {
		return fmt.Errorf("frequency must be > 0")
	}

	return nil
}

// SetCrossoverBand uploads a single crossover filter band to the device.
func (d *Device) SetCrossoverBand(band *CrossoverBand) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	maxChannel := 6
	if d.platform == PlatformRP2350 {
		maxChannel = 10
	}

	err := band.Validate(maxChannel)
	if err != nil {
		return fmt.Errorf("validating crossover band: %w", err)
	}

	buf := make([]byte, 16)
	buf[0] = byte(band.Channel)
	buf[1] = byte(band.Band)
	buf[2] = byte(band.Type)
	if band.Bypass {
		buf[3] = 1
	}
	binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(float32(band.Freq)))
	// Q and gain are ignored for crossover types; send safe defaults
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(0.707))
	binary.LittleEndian.PutUint32(buf[12:16], math.Float32bits(0))

	_, err = d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetEQParam, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_EQ_PARAM: %w", err)
	}

	return nil
}

// GetCrossoverBand reads a single crossover filter band from the device.
func (d *Device) GetCrossoverBand(channel, band int) (*CrossoverBand, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	maxChannel := 6
	if d.platform == PlatformRP2350 {
		maxChannel = 10
	}

	if channel < 0 || channel > maxChannel {
		return nil, fmt.Errorf("channel %d out of range (0-%d)", channel, maxChannel)
	}

	if channel < 2 {
		return nil, fmt.Errorf("crossover is not supported on master channels (channel %d)", channel)
	}

	if err := validateCrossoverBand(band); err != nil {
		return nil, err
	}

	// Firmware packs wValue as: bits[15:8]=channel, bits[7:3]=band (5 bits),
	// bits[2:0]=param (0=type, 1=freq, 2=Q, 3=gain, 4=bypass).
	base := uint16(channel)<<8 | uint16(band)<<3

	// param 0: type (uint32)
	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM type: %w", err)
	}
	filterType := CrossoverFilterType(binary.LittleEndian.Uint32(buf))

	// param 1: freq (float32)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base+1, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM freq: %w", err)
	}
	freq := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))

	// param 4: bypass (uint32)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base+4, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_EQ_PARAM bypass: %w", err)
	}
	bypass := binary.LittleEndian.Uint32(buf) != 0

	return &CrossoverBand{
		Channel: channel,
		Band:    band,
		Type:    filterType,
		Freq:    freq,
		Bypass:  bypass,
	}, nil
}

// MaxCrossoverBands returns the number of crossover bands per channel.
func (d *Device) MaxCrossoverBands() int {
	return maxXoverBands
}

// validateCrossoverBand checks that a band index is in the crossover range (20-23).
func validateCrossoverBand(band int) error {
	if band < xoverBandBase || band >= xoverBandBase+maxXoverBands {
		return fmt.Errorf("band %d is not a crossover band (%d-%d)", band, xoverBandBase, xoverBandBase+maxXoverBands-1)
	}
	return nil
}
