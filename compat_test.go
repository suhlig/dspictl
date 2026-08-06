package dspi

import (
	"errors"
	"testing"
	"time"

	"github.com/google/gousb"
)

// internalMockTransfer is a minimal USBControlTransfer for the internal
// compatibility tests (the dspi_test mock lives in the external package).
type internalMockTransfer struct {
	probeErr   error // returned for every transfer when set
	stallFirst int   // STALL (pipe error) the first N transfers
	timeouts   []time.Duration
	calls      []uint8
}

func (m *internalMockTransfer) ControlTransfer(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
	m.calls = append(m.calls, bRequest)
	if m.probeErr != nil {
		return 0, m.probeErr
	}
	if m.stallFirst > 0 {
		m.stallFirst--
		return 0, gousb.ErrorPipe
	}
	return len(data), nil
}

func (m *internalMockTransfer) Close() error { return nil }

func (m *internalMockTransfer) setControlTimeout(dur time.Duration) {
	m.timeouts = append(m.timeouts, dur)
}

func newGatedTestDevice(inner *internalMockTransfer, platform Platform) (*Device, *compatGate) {
	dev := NewDevice(inner, platform, "test-serial", 1, 2)
	gate := &compatGate{inner: dev.usb}
	dev.usb = gate
	return dev, gate
}

func TestCompatGateBlocksIncompatibleTransfers(t *testing.T) {
	inner := &internalMockTransfer{}
	gateErr := errors.New("device predates the V16 wire protocol")
	gate := &compatGate{inner: inner, err: gateErr}

	if _, err := gate.ControlTransfer(0, ReqGetEQParam, 0, vendorInterface, make([]byte, 4)); err != gateErr {
		t.Fatalf("expected the gate error, got %v", err)
	}
	if len(inner.calls) != 0 {
		t.Fatalf("no transfer should have reached the device, got %d", len(inner.calls))
	}

	// The bootloader command must pass through so an old device can be
	// upgraded into compatibility.
	if _, err := gate.ControlTransfer(0, ReqEnterBootloader, 0, vendorInterface, make([]byte, 1)); err != nil {
		t.Fatalf("enter bootloader should bypass the gate: %v", err)
	}
	if len(inner.calls) != 1 || inner.calls[0] != ReqEnterBootloader {
		t.Fatalf("expected only the bootloader call to reach the device, got %v", inner.calls)
	}
}

func TestCompatGatePassesCompatibleTransfers(t *testing.T) {
	inner := &internalMockTransfer{}
	gate := &compatGate{inner: inner}

	if _, err := gate.ControlTransfer(0, ReqGetEQParam, 0, vendorInterface, make([]byte, 4)); err != nil {
		t.Fatalf("expected the transfer to pass, got %v", err)
	}
	if len(inner.calls) != 1 {
		t.Fatalf("expected one transfer, got %d", len(inner.calls))
	}
}

func TestCompatGateForwardsControlTimeout(t *testing.T) {
	inner := &internalMockTransfer{}
	gate := &compatGate{inner: inner}

	gate.setControlTimeout(5 * time.Second)
	if len(inner.timeouts) != 1 || inner.timeouts[0] != 5*time.Second {
		t.Fatalf("expected the timeout to be forwarded, got %v", inner.timeouts)
	}
}

func TestProbeArmsTheGateOnStall(t *testing.T) {
	inner := &internalMockTransfer{probeErr: errors.New("libusb: pipe error")}
	dev, gate := newGatedTestDevice(inner, PlatformRP2040)

	// Mirrors Open(): the probe result arms the gate.
	gate.err = dev.probeFirmwareCompatibility()
	if gate.err == nil {
		t.Fatal("expected a compatibility error when the chunk probe stalls")
	}

	if _, err := dev.usb.ControlTransfer(0, ReqGetAllParamsChunk, 0, vendorInterface, make([]byte, 16)); err != gate.err {
		t.Fatalf("expected the armed gate to block transfers, got %v", err)
	}
}

func TestProbeLeavesTheGateOpenOnSuccess(t *testing.T) {
	inner := &internalMockTransfer{}
	dev, gate := newGatedTestDevice(inner, PlatformRP2350)

	gate.err = dev.probeFirmwareCompatibility()
	if gate.err != nil {
		t.Fatalf("expected no error when the probe succeeds, got %v", gate.err)
	}

	if _, err := dev.usb.ControlTransfer(0, ReqGetAllParamsChunk, 0, vendorInterface, make([]byte, 16)); err != nil {
		t.Fatalf("expected the transfer to pass, got %v", err)
	}
}
