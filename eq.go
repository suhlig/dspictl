package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
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
func (b *EQBand) Validate(maxChannel int) error {
	if b.Channel < 0 || b.Channel > maxChannel {
		return fmt.Errorf("channel %d out of range (0-%d)", b.Channel, maxChannel)
	}

	if b.Band < 0 || b.Band > 9 {
		return fmt.Errorf("band %d out of range (0-9)", b.Band)
	}

	if b.Type != FilterTypeFlat && b.Freq <= 0 {
		return fmt.Errorf("frequency must be > 0")
	}

	if b.Type != FilterTypeFlat && b.QualityFactor <= 0 {
		return fmt.Errorf("quality factor must be > 0")
	}

	return nil
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

	err := band.Validate(maxChannel)

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

	base := uint16(channel)<<8 | uint16(band)<<4

	// param 0: type (uint32)
	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetEQParam, base, vendorInterface, buf)

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

// SetBandBypass enables or disables bypass for a single EQ band.
func (d *Device) SetBandBypass(channel, band int, bypass bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	var val byte
	if bypass {
		val = 1
	}

	wValue := uint16(channel)<<8 | uint16(band)
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetBandBypass, wValue, vendorInterface, []byte{val})

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

	wValue := uint16(channel)<<8 | uint16(band)
	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetBandBypass, wValue, vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_BAND_BYPASS: %w", err)
	}

	return buf[0] != 0, nil
}
