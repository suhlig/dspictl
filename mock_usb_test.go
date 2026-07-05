package dspi_test

import (
	"fmt"

	"github.com/suhlig/dspi"
)

type mockControlTransfer struct {
	ReturnData       map[[3]uint16][]byte
	ReturnErrors     map[[3]uint16]error
	CapturedRequests []capturedControl
}

type capturedControl struct {
	BmRequestType uint8
	BRequest      uint8
	WValue        uint16
	WIndex        uint16
	Data          []byte
}

func (m *mockControlTransfer) ControlTransfer(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
	m.CapturedRequests = append(m.CapturedRequests, capturedControl{
		BmRequestType: bmRequestType,
		BRequest:      bRequest,
		WValue:        wValue,
		WIndex:        wIndex,
		Data:          append([]byte{}, data...),
	})

	key := [3]uint16{uint16(bRequest), wValue, wIndex}

	if err, ok := m.ReturnErrors[key]; ok {
		return 0, err
	}

	if ret, ok := m.ReturnData[key]; ok {
		n := copy(data, ret)

		return n, nil
	}

	return 0, fmt.Errorf("no mock data for request 0x%02x", bRequest)
}

func (m *mockControlTransfer) Close() error { return nil }

// newTestDevice creates a Device backed by a mockControlTransfer with the given platform.
func newTestDevice(mock *mockControlTransfer, platform dspi.Platform) *dspi.Device {
	return dspi.NewDevice(mock, platform, "test-serial", 1, 2)
}
