package dspi

import (
	"testing"
	"time"
)

// Bulk-transfer STALL retry tests.  The firmware STALLs bulk SET/GET while
// the main-loop apply of a previous SET is still running (bulk_params_pending)
// or another transport owns the shared bulk buffer; the host is expected to
// retry.  The backoff is shrunk to keep the tests fast and restored after.

func withFastBulkBackoff(t *testing.T) {
	t.Helper()
	original := bulkRetryBackoff
	bulkRetryBackoff = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	t.Cleanup(func() { bulkRetryBackoff = original })
}

func TestGetAllParamsRetriesOnStall(t *testing.T) {
	withFastBulkBackoff(t)

	inner := &internalMockTransfer{stallFirst: 1} // first chunk STALLs once
	dev := NewDevice(inner, PlatformRP2350, "test-serial", 1, 2)

	bp, err := dev.GetAllParams()
	if err != nil {
		t.Fatalf("expected the stalled read to be retried, got %v", err)
	}
	if len(bp.Raw) != wireBulkSize {
		t.Fatalf("expected a full %d-byte payload, got %d", wireBulkSize, len(bp.Raw))
	}
}

func TestGetAllParamsGivesUpAfterPersistentStall(t *testing.T) {
	withFastBulkBackoff(t)

	inner := &internalMockTransfer{stallFirst: 100}
	dev := NewDevice(inner, PlatformRP2350, "test-serial", 1, 2)

	_, err := dev.GetAllParams()
	if err == nil {
		t.Fatal("expected an error after the retries are exhausted")
	}
}

func TestSetAllParamsRetriesOnStall(t *testing.T) {
	withFastBulkBackoff(t)

	inner := &internalMockTransfer{stallFirst: 1} // first chunk STALLs once
	dev := NewDevice(inner, PlatformRP2350, "test-serial", 1, 2)

	err := dev.SetAllParams(&BulkParams{
		Header: BulkHeader{Platform: PlatformRP2350},
		Raw:    make([]byte, wireBulkSize),
	})
	if err != nil {
		t.Fatalf("expected the stalled upload to be retried, got %v", err)
	}
}

func TestSetAllParamsGivesUpAfterPersistentStall(t *testing.T) {
	withFastBulkBackoff(t)

	inner := &internalMockTransfer{stallFirst: 100}
	dev := NewDevice(inner, PlatformRP2350, "test-serial", 1, 2)

	err := dev.SetAllParams(&BulkParams{
		Header: BulkHeader{Platform: PlatformRP2350},
		Raw:    make([]byte, wireBulkSize),
	})
	if err == nil {
		t.Fatal("expected an error after the retries are exhausted")
	}
}
