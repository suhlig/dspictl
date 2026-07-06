package dspi

import (
	"encoding/binary"
	"fmt"
)

// DACHwMuteConfig holds the DAC hardware mute configuration.
type DACHwMuteConfig struct {
	Enabled   bool   // mute enabled
	ActiveLow bool   // active-low polarity
	Pin       int    // GPIO pin (255 = none / disabled)
	HoldMs    uint16 // hold time in ms before muting
	ReleaseMs uint16 // release time in ms after unmuting
}

// SetDACHwMute sets the DAC hardware mute configuration.
// The firmware expects a 16-byte payload:
//
//	byte 0: enabled (0/1)
//	byte 1: active_low (0/1)
//	byte 2: pin (0-255, 255 = none)
//	bytes 4-5: hold_ms (uint16 LE)
//	bytes 6-7: release_ms (uint16 LE)
func (d *Device) SetDACHwMute(cfg *DACHwMuteConfig) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	data := make([]byte, 16)

	if cfg.Enabled {
		data[0] = 1
	}
	if cfg.ActiveLow {
		data[1] = 1
	}
	data[2] = byte(cfg.Pin)
	binary.LittleEndian.PutUint16(data[4:6], cfg.HoldMs)
	binary.LittleEndian.PutUint16(data[6:8], cfg.ReleaseMs)

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetDACHwMuteConfig, 0, vendorInterface, data)
	if err != nil {
		return fmt.Errorf("REQ_SET_DAC_HW_MUTE_CONFIG: %w", err)
	}

	return nil
}

// GetDACHwMute reads the current DAC hardware mute configuration.
func (d *Device) GetDACHwMute() (*DACHwMuteConfig, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetDACHwMuteConfig, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_DAC_HW_MUTE_CONFIG: %w", err)
	}

	cfg := &DACHwMuteConfig{
		Enabled:   buf[0] != 0,
		ActiveLow: buf[1] != 0,
		Pin:       int(buf[2]),
		HoldMs:    binary.LittleEndian.Uint16(buf[4:6]),
		ReleaseMs: binary.LittleEndian.Uint16(buf[6:8]),
	}

	return cfg, nil
}

// TestDACHwMute triggers a test of the DAC hardware mute.
// The device acknowledges and then performs a brief mute cycle.
func (d *Device) TestDACHwMute() error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqTestDACHwMute, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_TEST_DAC_HW_MUTE: %w", err)
	}

	return nil
}
