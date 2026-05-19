package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newClearClipsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-clips",
		Short: "Clear clip detection latches",
		RunE:  runClearClips,
	}
}

func runClearClips(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.ClearClips()

		if err != nil {
			slog.Error("clearing clips failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: clips cleared\n", d.Serial())
	}

	return nil
}
