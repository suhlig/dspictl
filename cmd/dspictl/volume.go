package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Master volume get/set",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Print current master volume",
		RunE:  runVolumeGet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <db>",
		Short: "Set master volume (-128 to 0 dB)",
		Args:  cobra.ExactArgs(1),
		RunE:  runVolumeSet,
	})

	return cmd
}

func runVolumeGet(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		gain, err := d.GetMasterVolume()

		if err != nil {
			slog.Error("getting master volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), gain)
	}

	return nil
}

func runVolumeSet(cmd *cobra.Command, args []string) error {
	db, err := strconv.ParseFloat(args[0], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMasterVolume(dspi.NewGain(db))

		if err != nil {
			slog.Error("setting master volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), dspi.NewGain(db))
	}

	return nil
}
