package dspi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/google/gousb"
)

// USBErrorStats holds USB PHY error counters.
type USBErrorStats struct {
	CRC      uint32
	BitStuff uint32
	Timeout  uint32
	Overflow uint32
	Sequence uint32
	Unknown  uint32
}

// BufferStats holds raw buffer fill statistics.
type BufferStats struct {
	Data []byte
}

// FactoryReset resets the live DSP state to factory defaults.
// Does NOT erase any preset slots.
func (d *Device) FactoryReset() error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqFactoryReset, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_FACTORY_RESET: %w", err)
	}

	return nil
}

// EnterBootloader sends the enter-bootloader command, causing the device to
// reboot into UF2 mode for firmware updates.
//
// The device acknowledges the command and then disconnects to reboot into the
// ROM bootloader. A 5-second control transfer timeout prevents hanging on
// macOS when the device disconnects mid-transfer. Any libusb error (including
// timeout) is treated as success — the device rebooting is expected.
func (d *Device) EnterBootloader() error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	d.setControlTimeout(5 * time.Second)

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqEnterBootloader, 0, vendorInterface, buf)

	if err != nil {
		var usbErr gousb.Error

		if errors.As(err, &usbErr) {
			return nil
		}

		return fmt.Errorf("REQ_ENTER_BOOTLOADER: %w", err)
	}

	return nil
}

// GetCore1Mode queries the Core 1 operating mode.
// Returns: 0=Idle, 1=PDM, 2=EQ Worker.
func (d *Device) GetCore1Mode() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCore1Mode, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_CORE1_MODE: %w", err)
	}

	return int(buf[0]), nil
}

// GetCore1Conflict checks if a PDM vs EQ Worker conflict exists.
func (d *Device) GetCore1Conflict() (bool, error) {
	if d.closed {
		return false, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCore1Conflict, 0, vendorInterface, buf)

	if err != nil {
		return false, fmt.Errorf("REQ_GET_CORE1_CONFLICT: %w", err)
	}

	return buf[0] != 0, nil
}

// GetBufferStats reads buffer fill statistics from the device.
func (d *Device) GetBufferStats() (*BufferStats, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 64)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetBufferStats, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_BUFFER_STATS: %w", err)
	}

	return &BufferStats{Data: buf[:n]}, nil
}

// GetUSBErrorStats reads USB PHY error counters.
func (d *Device) GetUSBErrorStats() (*USBErrorStats, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 24)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetUSBErrorStats, 0, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_USB_ERROR_STATS: %w", err)
	}

	stats := &USBErrorStats{
		CRC:      binary.LittleEndian.Uint32(buf[0:4]),
		BitStuff: binary.LittleEndian.Uint32(buf[4:8]),
		Timeout:  binary.LittleEndian.Uint32(buf[8:12]),
		Overflow: binary.LittleEndian.Uint32(buf[12:16]),
		Sequence: binary.LittleEndian.Uint32(buf[16:20]),
		Unknown:  binary.LittleEndian.Uint32(buf[20:24]),
	}

	return stats, nil
}
