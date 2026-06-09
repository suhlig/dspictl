package dspi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PresetDirectory contains the metadata for all preset slots.
type PresetDirectory struct {
	SlotOccupied     uint16
	StartupMode      int
	DefaultSlot      int
	LastActive       int
	OutputConfigMode int // 0=independent, 1=per-preset
	MasterVolMode    int
}

func (d *Device) PresetSave(slot int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetSave, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_SAVE: %w", err)
	}

	if buf[0] != 0 {
		return fmt.Errorf("REQ_PRESET_SAVE: status 0x%02X", buf[0])
	}

	return nil
}

func (d *Device) PresetLoad(slot int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetLoad, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_LOAD: %w", err)
	}

	if buf[0] != 0 {
		return fmt.Errorf("REQ_PRESET_LOAD: status 0x%02X", buf[0])
	}

	return nil
}

func (d *Device) PresetDelete(slot int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetDelete, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_DELETE: %w", err)
	}

	if buf[0] != 0 {
		return fmt.Errorf("REQ_PRESET_DELETE: status 0x%02X", buf[0])
	}

	return nil
}

func (d *Device) GetPresetName(slot int) (string, error) {
	if d.closed {
		return "", fmt.Errorf("device is closed")
	}

	buf := make([]byte, 32)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetGetName, uint16(slot), vendorInterface, buf)

	if err != nil {
		return "", fmt.Errorf("REQ_PRESET_GET_NAME: %w", err)
	}

	end := bytes.IndexByte(buf, 0)

	if end == -1 {
		end = len(buf)
	}

	return string(bytes.TrimSpace(buf[:end])), nil
}

func (d *Device) SetPresetName(slot int, name string) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 32)
	copy(buf, []byte(name))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqPresetSetName, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_SET_NAME: %w", err)
	}

	return nil
}

func (d *Device) GetPresetDirectory() (*PresetDirectory, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 7)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetGetDir, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_PRESET_GET_DIR: %w", err)
	}

	dir := &PresetDirectory{
		SlotOccupied:     binary.LittleEndian.Uint16(buf[0:2]),
		StartupMode:      int(buf[2]),
		DefaultSlot:      int(buf[3]),
		LastActive:       int(buf[4]),
		OutputConfigMode: int(buf[5]),
		MasterVolMode:    int(buf[6]),
	}

	return dir, nil
}

func (d *Device) GetActivePreset() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetGetActive, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_PRESET_GET_ACTIVE: %w", err)
	}

	return int(buf[0]), nil
}

// SetPresetStartup sets the startup mode and default slot.
// mode: 0=specified (use defaultSlot), 1=last active.
func (d *Device) SetPresetStartup(mode, defaultSlot int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := []byte{byte(mode), byte(defaultSlot)}
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqPresetSetStartup, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_SET_STARTUP: %w", err)
	}

	return nil
}

// GetPresetStartup returns the startup mode and default slot.
// mode: 0=specified, 1=last active.
func (d *Device) GetPresetStartup() (mode int, defaultSlot int, err error) {
	if d.closed {
		return 0, 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 2)
	_, err = d.usb.ControlTransfer(vendorInterfaceInRequest, ReqPresetGetStartup, 0, vendorInterface, buf)

	if err != nil {
		return 0, 0, fmt.Errorf("REQ_PRESET_GET_STARTUP: %w", err)
	}

	return int(buf[0]), int(buf[1]), nil
}

// SetOutputConfigMode sets whether output configuration is stored per-preset or independently.
// mode: 0=independent, 1=per-preset.
func (d *Device) SetOutputConfigMode(mode int) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetOutputConfigMode, 0, vendorInterface, []byte{byte(mode)})

	if err != nil {
		return fmt.Errorf("REQ_SET_OUTPUT_CONFIG_MODE: %w", err)
	}

	return nil
}

// GetOutputConfigMode returns the output configuration persistence mode.
// 0=independent, 1=per-preset.
func (d *Device) GetOutputConfigMode() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetOutputConfigMode, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_OUTPUT_CONFIG_MODE: %w", err)
	}

	return int(buf[0]), nil
}
