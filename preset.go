package dspi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PresetDirectory contains the metadata for all preset slots.
type PresetDirectory struct {
	SlotOccupied  uint16
	StartupMode   int
	DefaultSlot   int
	LastActive    int
	IncludePins   bool
	MasterVolMode int
}

func (d *Device) PresetSave(slot int) error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.device.Control(vendorInterfaceInRequest, reqPresetSave, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_SAVE: %w", err)
	}

	if buf[0] != 0 {
		return fmt.Errorf("REQ_PRESET_SAVE: status 0x%02X", buf[0])
	}

	return nil
}

func (d *Device) PresetLoad(slot int) error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.device.Control(vendorInterfaceInRequest, reqPresetLoad, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_LOAD: %w", err)
	}

	if buf[0] != 0 {
		return fmt.Errorf("REQ_PRESET_LOAD: status 0x%02X", buf[0])
	}

	return nil
}

func (d *Device) PresetDelete(slot int) error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.device.Control(vendorInterfaceInRequest, reqPresetDelete, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_DELETE: %w", err)
	}

	if buf[0] != 0 {
		return fmt.Errorf("REQ_PRESET_DELETE: status 0x%02X", buf[0])
	}

	return nil
}

func (d *Device) GetPresetName(slot int) (string, error) {
	if d.device == nil {
		return "", fmt.Errorf("device is closed")
	}

	buf := make([]byte, 32)
	_, err := d.device.Control(vendorInterfaceInRequest, reqPresetGetName, uint16(slot), vendorInterface, buf)

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
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 32)
	copy(buf, []byte(name))
	_, err := d.device.Control(vendorInterfaceOutRequest, reqPresetSetName, uint16(slot), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_PRESET_SET_NAME: %w", err)
	}

	return nil
}

func (d *Device) GetPresetDirectory() (*PresetDirectory, error) {
	if d.device == nil {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 7)
	_, err := d.device.Control(vendorInterfaceInRequest, reqPresetGetDir, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_PRESET_GET_DIR: %w", err)
	}

	dir := &PresetDirectory{
		SlotOccupied:  binary.LittleEndian.Uint16(buf[0:2]),
		StartupMode:   int(buf[2]),
		DefaultSlot:   int(buf[3]),
		LastActive:    int(buf[4]),
		IncludePins:   buf[5] != 0,
		MasterVolMode: int(buf[6]),
	}

	return dir, nil
}

func (d *Device) GetActivePreset() (int, error) {
	if d.device == nil {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.device.Control(vendorInterfaceInRequest, reqPresetGetActive, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_PRESET_GET_ACTIVE: %w", err)
	}

	return int(buf[0]), nil
}
