package dspi

import "fmt"

// Platform identifies the DSPi hardware platform.
type Platform int

const (
	PlatformRP2040 Platform = 0
	PlatformRP2350 Platform = 1
)

func (p Platform) String() string {
	switch p {
	case PlatformRP2040:
		return "RP2040"
	case PlatformRP2350:
		return "RP2350"
	default:
		return "Unknown"
	}
}

// FirmwareVersion holds the decoded firmware version from REQ_GET_PLATFORM.
type FirmwareVersion struct {
	Major uint8
	Minor uint8
	Patch uint8
}

// String formats the version as "v1.2.3".
func (v FirmwareVersion) String() string {
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		return "unknown"
	}

	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}
