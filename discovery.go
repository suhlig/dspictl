package dspi

import (
	"fmt"

	"github.com/google/gousb"
)

// DeviceInfo describes a discovered DSPi device without an open connection.
type DeviceInfo struct {
	Serial  string
	Bus     int
	Address int
}

// List enumerates all connected DSPi devices.
func List() ([]DeviceInfo, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(dspiVID) && desc.Product == gousb.ID(dspiPID)
	})

	if err != nil {
		return nil, fmt.Errorf("enumerating DSPi devices: %w", err)
	}

	infos := make([]DeviceInfo, 0, len(devs))

	for _, dev := range devs {
		serial, err := dev.SerialNumber()

		if err != nil {
			dev.Close()

			continue
		}

		desc := dev.Desc
		infos = append(infos, DeviceInfo{
			Serial:  serial,
			Bus:     desc.Bus,
			Address: desc.Address,
		})
		dev.Close()
	}

	return infos, nil
}
