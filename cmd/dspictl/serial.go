package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newSerialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serial",
		Short: "Read the firmware serial number from each device",
		RunE:  runSerial,
	}
}

func runSerial(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		serial, err := d.GetSerial()

		if err != nil {
			slog.Error("getting serial failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Println(serial)
	}

	return nil
}
