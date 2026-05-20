package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// MatrixRoute describes a single crosspoint in the matrix mixer.
type MatrixRoute struct {
	Input       int
	Output      int
	Enabled     bool
	PhaseInvert bool
	Gain        Gain
}

func (d *Device) SetMatrixRoute(route *MatrixRoute) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	enabled := byte(0)

	if route.Enabled {
		enabled = 1
	}

	phase := byte(0)

	if route.PhaseInvert {
		phase = 1
	}

	buf[0] = byte(route.Input)
	buf[1] = byte(route.Output)
	buf[2] = enabled
	buf[3] = phase
	binary.LittleEndian.PutUint32(buf[4:8], math.Float32bits(float32(route.Gain.DB())))

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetMatrixRoute, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MATRIX_ROUTE: %w", err)
	}

	return nil
}

func (d *Device) GetMatrixRoute(input, output int) (*MatrixRoute, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	wValue := uint16(input)<<8 | uint16(output)
	buf := make([]byte, 8)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetMatrixRoute, wValue, vendorInterface, buf)

	if err != nil {
		return nil, fmt.Errorf("REQ_GET_MATRIX_ROUTE: %w", err)
	}

	route := &MatrixRoute{
		Input:       int(buf[0]),
		Output:      int(buf[1]),
		Enabled:     buf[2] != 0,
		PhaseInvert: buf[3] != 0,
		Gain:        NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[4:8])))),
	}

	return route, nil
}
