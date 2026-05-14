package dspi

import (
	"fmt"
	"math"
)

// MeterSnapshot contains a single poll of all telemetry data.
type MeterSnapshot struct {
	Peaks     [maxChannels]float64
	CPU0      int
	CPU1      int
	ClipFlags uint16
	Channels  int
	err       error
}

func (m MeterSnapshot) Err() error { return m.err }

// DBFS converts a linear peak (0.0-1.0) to dBFS. Returns -inf for zero.
func DBFS(linear float64) string {
	if linear <= 0 {
		return "-inf"
	}

	return fmt.Sprintf("%.1f", 20*math.Log10(linear))
}
