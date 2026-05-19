package dspi

import "fmt"

// SetMasterVolumeMode sets the persistence mode for master volume.
// mode: 0=independent, 1=with preset.
func (d *Device) SetMasterVolumeMode(mode int) error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	_, err := d.device.Control(vendorInterfaceOutRequest, reqSetMasterVolumeMode, 0, vendorInterface, []byte{byte(mode)})

	if err != nil {
		return fmt.Errorf("REQ_SET_MASTER_VOLUME_MODE: %w", err)
	}

	return nil
}

// GetMasterVolumeMode returns the current persistence mode.
// 0=independent, 1=with preset.
func (d *Device) GetMasterVolumeMode() (int, error) {
	if d.device == nil {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.device.Control(vendorInterfaceInRequest, reqGetMasterVolumeMode, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_MASTER_VOLUME_MODE: %w", err)
	}

	return int(buf[0]), nil
}

// SaveMasterVolume saves the current live master volume as the boot default
// (used in mode 0 persistence).
func (d *Device) SaveMasterVolume() error {
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.device.Control(vendorInterfaceInRequest, reqSaveMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SAVE_MASTER_VOLUME: %w", err)
	}

	return nil
}
