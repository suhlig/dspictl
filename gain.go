package dspi

import "fmt"

// Gain represents a gain level in decibels (dB).
type Gain float64

// NewGain creates a Gain from a dB value.
func NewGain(db float64) Gain {
	return Gain(db)
}

// DB returns the gain value in dB.
func (g Gain) DB() float64 {
	return float64(g)
}

// String formats the gain as a dB string.
// Returns "MUTE" for the mute sentinel value (-128 dB or below).
func (g Gain) String() string {
	if g <= -128 {
		return "MUTE"
	}

	return fmt.Sprintf("%.0f dB", float64(g))
}
