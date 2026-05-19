package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newUSBErrorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "usb-errors",
		Short: "Read USB PHY error counters",
		RunE:  runUSBErrors,
	}
}

func runUSBErrors(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		stats, err := d.GetUSBErrorStats()

		if err != nil {
			slog.Error("getting USB error stats failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  CRC errors:     %d\n", stats.CRC)
		fmt.Printf("  Bit-stuff:      %d\n", stats.BitStuff)
		fmt.Printf("  Timeout:        %d\n", stats.Timeout)
		fmt.Printf("  Overflow:       %d\n", stats.Overflow)
		fmt.Printf("  Sequence:       %d\n", stats.Sequence)
		fmt.Printf("  Unknown:        %d\n", stats.Unknown)
	}

	return nil
}
