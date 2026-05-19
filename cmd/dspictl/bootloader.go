package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newBootloaderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootloader",
		Short: "Reboot into UF2 bootloader for firmware updates",
		RunE:  runBootloader,
	}
}

func runBootloader(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.EnterBootloader()

		if err != nil {
			slog.Error("entering bootloader failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: rebooting into bootloader\n", d.Serial())
	}

	return nil
}
