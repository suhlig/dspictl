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

		inputSource, inputSourceErr := d.GetInputSource()
		inputRate, inputRateErr := d.GetInputRate()
		sampleRate, sampleRateErr := d.GetSampleRate()

		fwVersion := d.FirmwareVersion().String()

		fmt.Printf("Serial: %s\n", d.Serial())
		fmt.Printf("  USB Bus Number: %d\n", d.Bus())
		fmt.Printf("  USB Device Address: %d\n", d.Address())
		fmt.Printf("  Type: %s\n", d.Platform())
		fmt.Printf("  Firmware: %s\n", fwVersion)
		fmt.Printf("  Volume: %s\n", volume)
		fmt.Printf("  Preset: %d\n", preset)

		if inputSourceErr != nil {
			fmt.Printf("  Input: (unknown)\n")
			fmt.Printf("  Rate: (unknown)\n")
		} else {
			fmt.Printf("  Input: %s\n", dspi.InputSourceName(inputSource))
			if inputRateErr != nil {
				fmt.Printf("  Rate: (unknown)\n")
			} else {
				fmt.Printf("  Rate: %d Hz", inputRate.PipelineHz)
				if inputSource == dspi.InputSourceI2S {
					fmt.Printf(" (I2S %d Hz)", inputRate.SelectedHz)
				}
				fmt.Printf("\n")

				if sampleRateErr == nil && sampleRate > 0 {
					fmt.Printf("  Sample Rate (raw): %d Hz\n", sampleRate)
				}
			}

			if inputSource == dspi.InputSourceADAT {
				printAdatInputStatusCompact(d)
			}
		}

		printMCKStatus(d)

		printLoudnessCompact(d)

		printSiggenCompact(d)

		fmt.Printf("  CPU: %d%% / %d%%\n", meter.CPU0, meter.CPU1)
	}

	return nil
}

// printSiggenCompact prints a one-line signal-generator summary for use in status output.
// It silently skips the line if the firmware does not support the generator.
func printSiggenCompact(d *dspi.Device) {
	status, err := d.GetSiggenStatus()
	if err != nil {
		return
	}

	if status.State == dspi.SiggenStateIdle {
		fmt.Printf("  Siggen: idle\n")
		return
	}

	fmt.Printf("  Siggen: %s %s", status.State, status.SignalType)

	if cfg, err := d.GetSiggenConfig(); err == nil && cfg.ChannelMask != 0 {
		fmt.Printf(" on %s", formatMaskU16(cfg.ChannelMask, 16))
	}

	if status.ActiveChannel >= 0 {
		fmt.Printf(" (walk ch %d)", status.ActiveChannel)
	}

	if status.CurrentFreq > 0 {
		fmt.Printf(" @ %.1f Hz", status.CurrentFreq)
	}

	fmt.Println()
}

// printAdatInputStatusCompact prints a one-line ADAT input summary for use in status output.
func printAdatInputStatusCompact(d *dspi.Device) {
	status, err := d.GetAdatInputStatus()
	if err != nil {
		return // silently skip if firmware doesn't support it
	}

	fmt.Printf("  ADAT Input: %s", status.State)
	if status.DetectedRate > 0 {
		fmt.Printf(" @ %d Hz", status.DetectedRate)
	}
	fmt.Printf(" (%s, rate_ok=%v)\n", dspi.AdatClockModeName(status.ClockMode), status.RateOK)
}

// printMCKStatus prints a one-line MCK summary for use in status output.
func printMCKStatus(d *dspi.Device) {
	enabled, err := d.GetMCKEnable()

	if err != nil {
		return // silently skip if firmware doesn't support it
	}

	pin, err := d.GetMCKPin()

	if err != nil {
		return
	}

	multiplier, err := d.GetMCKMultiplier()

	if err != nil {
		return
	}

	label := "128"

	if multiplier == 1 {
		label = "256"
	}

	fmt.Printf("  MCK: %v (GPIO %d, %s×)\n", enabled, pin, label)
}
