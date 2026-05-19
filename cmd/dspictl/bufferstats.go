package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newBufferStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "buffer-stats",
		Short: "Read buffer fill statistics",
		RunE:  runBufferStats,
	}
}

func runBufferStats(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		stats, err := d.GetBufferStats()

		if err != nil {
			slog.Error("getting buffer stats failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:", d.Serial())

		for i, b := range stats.Data {
			if i%16 == 0 {
				fmt.Printf("\n  ")
			} else if i%4 == 0 {
				fmt.Printf(" ")
			}

			fmt.Printf("%02x", b)
		}

		fmt.Println()
	}

	return nil
}
