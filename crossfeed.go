package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// CrossfeedStatus holds the current crossfeed settings.
type CrossfeedStatus struct {
	Enabled bool
	Preset  int     // 0–4: default, diffuse, careful, medium, aggressive
	Freq    float64 // crossover frequency in Hz
	Feed    float64 // feed level in dB
	ITD     bool    // interaural time delay enabled
}

// SetCrossfeed enables or disables crossfeed.
func (d *Device) SetCrossfeed(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCrossfeed, 0, vendorInterface, []byte{val})
	if err != nil {
		return fmt.Errorf("REQ_SET_CROSSFEED: %w", err)
	}

	return nil
}

// GetCrossfeed returns whether crossfeed is currently enabled.
func (d *Device) GetCrossfeed() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCrossfeed, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_CROSSFEED: %w", err)
	}

	return buf[0] != 0, nil
}

// SetCrossfeedPreset sets the crossfeed preset (0–4).
// 0=default, 1=diffuse, 2=careful, 3=medium, 4=aggressive.
func (d *Device) SetCrossfeedPreset(preset int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCrossfeedPreset, 0, vendorInterface, []byte{byte(preset)})
	if err != nil {
		return fmt.Errorf("REQ_SET_CROSSFEED_PRESET: %w", err)
	}

	return nil
}

// GetCrossfeedPreset returns the current crossfeed preset (0–4).
func (d *Device) GetCrossfeedPreset() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCrossfeedPreset, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_CROSSFEED_PRESET: %w", err)
	}

	return int(buf[0]), nil
}

// SetCrossfeedFreq sets the crossfeed crossover frequency in Hz.
func (d *Device) SetCrossfeedFreq(freq float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(freq)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCrossfeedFreq, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_CROSSFEED_FREQ: %w", err)
	}

	return nil
}

// GetCrossfeedFreq returns the crossfeed crossover frequency in Hz.
func (d *Device) GetCrossfeedFreq() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCrossfeedFreq, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_CROSSFEED_FREQ: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}

// SetCrossfeedFeed sets the crossfeed feed level in dB.
func (d *Device) SetCrossfeedFeed(feed float64) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(feed)))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCrossfeedFeed, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SET_CROSSFEED_FEED: %w", err)
	}

	return nil
}

// GetCrossfeedFeed returns the crossfeed feed level in dB.
func (d *Device) GetCrossfeedFeed() (float64, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCrossfeedFeed, 0, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_GET_CROSSFEED_FEED: %w", err)
	}

	return float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))), nil
}

// SetCrossfeedITD enables or disables interaural time delay for crossfeed.
func (d *Device) SetCrossfeedITD(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCrossfeedITD, 0, vendorInterface, []byte{val})
	if err != nil {
		return fmt.Errorf("REQ_SET_CROSSFEED_ITD: %w", err)
	}

	return nil
}

// GetCrossfeedITD returns whether interaural time delay is enabled.
func (d *Device) GetCrossfeedITD() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCrossfeedITD, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_CROSSFEED_ITD: %w", err)
	}

	return buf[0] != 0, nil
}
