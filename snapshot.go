package dspi

import (
	"fmt"
	"math"
)

// Level represents an audio signal level in linear scale (0.0–1.0).
type Level float64

// NewLevel constructs a Level from a linear amplitude value.
func NewLevel(v float64) Level {
	return Level(v)
}

// Linear returns the linear amplitude value.
func (l Level) Linear() float64 {
	return float64(l)
}

// DBFS converts the level to decibels relative to full scale.
// Returns -Inf for zero or negative values.
func (l Level) DBFS() float64 {
	if l <= 0 {
		return math.Inf(-1)
	}

	return 20 * math.Log10(float64(l))
}

// String formats the level as a dBFS string (e.g. "-6.0 dBFS", "-∞ dBFS").
func (l Level) String() string {
	dbfs := l.DBFS()

	if math.IsInf(dbfs, -1) {
		return "-∞ dBFS"
	}

	return fmt.Sprintf("%.1f dBFS", dbfs)
}

// MeterSnapshot contains a single poll of all telemetry data.
type MeterSnapshot struct {
	Peaks     [maxChannels]Level
	CPU0      int
	CPU1      int
	ClipFlags uint16
	Channels  int
	err       error
}

func (m MeterSnapshot) Err() error { return m.err }
