package dspi

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

