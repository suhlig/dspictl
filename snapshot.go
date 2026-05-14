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

// DBFS converts a linear peak (0.0-1.0) to dBFS. Returns -Inf for zero.
func DBFS(linear float64) float64 {
	if linear <= 0 {
		return math.Inf(-1)
	}

	return 20 * math.Log10(linear)
}

// FormatDBFS formats a dBFS value for display.
func FormatDBFS(dbfs float64) string {
	if math.IsInf(dbfs, -1) {
		return "-inf"
	}

	return fmt.Sprintf("%.1f", dbfs)
}
