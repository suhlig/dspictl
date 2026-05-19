package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newFactoryResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "factory-reset",
		Short: "Reset live DSP state to factory defaults",
		RunE:  runFactoryReset,
	}
}

func runFactoryReset(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.FactoryReset()

		if err != nil {
			slog.Error("factory reset failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: factory reset\n", d.Serial())
	}

	return nil
}
