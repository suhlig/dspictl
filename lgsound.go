package dspi

import (
	"fmt"
)

// LGSoundSyncStatus holds the current LG Sound Sync status.
type LGSoundSyncStatus struct {
	Enabled bool
	Present bool // TV present / connected
	Volume  int  // volume value (0–100), -1 = unknown
	Muted   bool
}

// SetLGSoundSync enables or disables LG Sound Sync.
func (d *Device) SetLGSoundSync(enabled bool) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	val := byte(0)
	if enabled {
		val = 1
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetLGSoundSync, 0, vendorInterface, []byte{val})
	if err != nil {
		return fmt.Errorf("REQ_SET_LG_SOUND_SYNC: %w", err)
	}

	return nil
}

// GetLGSoundSync returns whether LG Sound Sync is enabled.
func (d *Device) GetLGSoundSync() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLGSoundSync, 0, vendorInterface, buf)
	if err != nil {
		return false, fmt.Errorf("REQ_GET_LG_SOUND_SYNC: %w", err)
	}

	return buf[0] != 0, nil
}

// GetLGSoundSyncStatus reads the current LG Sound Sync status.
// Returns status with enabled, present/connected, volume, and muted fields.
//
// Response bytes:
//
//	byte 0: enabled (0/1)
//	byte 1: TV present (0/1)
//	byte 2: volume value (0–100), 0xFF = unknown
//	byte 3: muted (0/1)
func (d *Device) GetLGSoundSyncStatus() (*LGSoundSyncStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetLGSoundSyncStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_LG_SOUND_SYNC_STATUS: %w", err)
	}

	status := &LGSoundSyncStatus{}

	if n >= 1 {
		status.Enabled = buf[0] != 0
	}
	if n >= 2 {
		status.Present = buf[1] != 0
	}
	if n >= 3 {
		if buf[2] == 0xFF {
			status.Volume = -1
		} else {
			status.Volume = int(buf[2])
		}
	}
	if n >= 4 {
		status.Muted = buf[3] != 0
	}

	return status, nil
}
