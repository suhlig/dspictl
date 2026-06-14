package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show a summary of connected DSPi devices",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		volume, err := d.GetMasterVolume()

		if err != nil {
			slog.Error("getting master volume failed", "serial", d.Serial(), "error", err)
			continue
		}

		preset, err := d.GetActivePreset()

		if err != nil {
			slog.Error("getting active preset failed", "serial", d.Serial(), "error", err)
			continue
		}

		meter := d.ReadMeter()

		if meter.Err() != nil {
			slog.Error("reading meter failed", "serial", d.Serial(), "error", meter.Err())
			continue
		}

		inputSource, err := d.GetInputSource()

		if err != nil {
			slog.Error("getting input source failed", "serial", d.Serial(), "error", err)
			continue
		}

		inputRate, err := d.GetInputRate()

		if err != nil {
			slog.Error("getting input rate failed", "serial", d.Serial(), "error", err)
			continue
		}

		fwVersion := d.FirmwareVersion().String()

		fmt.Printf("Serial: %s\n", d.Serial())
		fmt.Printf("  USB Bus Number: %d\n", d.Bus())
		fmt.Printf("  USB Device Address: %d\n", d.Address())
		fmt.Printf("  Type: %s\n", d.Platform())
		fmt.Printf("  Firmware: %s\n", fwVersion)
		fmt.Printf("  Volume: %s\n", volume)
		fmt.Printf("  Preset: %d\n", preset)
		fmt.Printf("  Input: %s\n", dspi.InputSourceName(inputSource))
		fmt.Printf("  Rate: %d Hz", inputRate.PipelineHz)

		if inputSource == dspi.InputSourceI2S {
			fmt.Printf(" (I2S %d Hz)", inputRate.SelectedHz)
		}

		fmt.Printf("\n")

		printLoudnessCompact(d)

		fmt.Printf("  CPU: %d%% / %d%%\n", meter.CPU0, meter.CPU1)
	}

	return nil
}
