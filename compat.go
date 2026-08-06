package dspi

import (
	"fmt"
	"time"
)

// dspictl 2.x speaks the V16+ wire protocol (unified channel model, chunked
// bulk snapshots).  The chunked bulk commands (REQ_GET_ALL_PARAMS_CHUNK /
// REQ_SET_ALL_PARAMS_CHUNK) only exist since firmware v1.1.5-beta3; older
// firmware STALLs them.  A firmware version string alone cannot identify the
// protocol generation (the early v1.1.5 betas report the same 1.1.x shape as
// final releases), so compatibility is probed functionally at open.

// compatGate wraps a USBControlTransfer and refuses every transfer with a
// clear error once the device firmware has been identified as predating the
// V16 wire protocol.  The bootloader command (REQ_ENTER_BOOTLOADER) always
// passes through: entering the bootloader is how an old device is upgraded.
type compatGate struct {
	inner USBControlTransfer
	err   error // non-nil: firmware incompatible; returned for every transfer
}

func (g *compatGate) ControlTransfer(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
	if g.err != nil && bRequest != ReqEnterBootloader {
		return 0, g.err
	}
	return g.inner.ControlTransfer(bmRequestType, bRequest, wValue, wIndex, data)
}

func (g *compatGate) Close() error { return g.inner.Close() }

func (g *compatGate) setControlTimeout(dur time.Duration) {
	if s, ok := g.inner.(controlTimeoutSetter); ok {
		s.setControlTimeout(dur)
	}
}

// probeFirmwareCompatibility requests the first 16 bytes of the chunked bulk
// snapshot.  Firmware with the V16+ wire protocol answers; older firmware
// STALLs the request.  The result arms the compatGate installed by Open().
//
// The probe is retried on failure because the bulk buffer can transiently be
// owned by an external transport mid-stream; a genuine STALL from old
// firmware is deterministic and survives the retries.
func (d *Device) probeFirmwareCompatibility() error {
	const attempts = 3

	var lastErr error
	for range attempts {
		buf := make([]byte, 16)
		if _, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetAllParamsChunk, 0, vendorInterface, buf); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf(
		"device %s (firmware %s) predates the V16 wire protocol: the chunked bulk snapshot STALLed (%v); "+
			"dspictl requires firmware v1.1.5-beta3 or later — upgrade with `dspictl firmware upgrade`",
		d.serial, d.fwVersion, lastErr)
}
