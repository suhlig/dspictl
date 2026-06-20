package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Input source selection and I2S configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "source [<usb|spdif|i2s>]",
		Short:             "Get or set the active input source",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputSource,
		ValidArgsFunction: completeChoices([]string{"usb", "spdif", "i2s"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "rate [<44100|48000|96000>]",
		Short:             "Get or set the I2S input sample rate",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputRate,
		ValidArgsFunction: completeChoices([]string{"44100", "48000", "96000"}),
	})

	return cmd
}

func runInputSource(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			src, err := d.GetInputSource()

			if err != nil {
				slog.Error("getting input source failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: %s\n", d.Serial(), dspi.InputSourceName(src))
		}

		return nil
	}

	var source int

	switch args[0] {
	case "usb":
		source = dspi.InputSourceUSB
	case "spdif":
		source = dspi.InputSourceSPDIF
	case "i2s":
		source = dspi.InputSourceI2S
	default:
		return fmt.Errorf("invalid source: %s (expected usb, spdif, or i2s)", args[0])
	}

	for _, d := range devices {
		if err := d.SetInputSource(source); err != nil {
			slog.Error("setting input source failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), dspi.InputSourceName(source))
	}

	return nil
}

func runInputRate(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			rate, err := d.GetInputRate()

			if err != nil {
				slog.Error("getting input rate failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: pipeline=%d Hz  selected_i2s=%d Hz\n", d.Serial(), rate.PipelineHz, rate.SelectedHz)
		}

		return nil
	}

	var hz uint32

	switch args[0] {
	case "44100":
		hz = 44100
	case "48000":
		hz = 48000
	case "96000":
		hz = 96000
	default:
		return fmt.Errorf("invalid rate: %s (expected 44100, 48000, or 96000)", args[0])
	}

	for _, d := range devices {
		if err := d.SetInputRate(hz); err != nil {
			slog.Error("setting input rate failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: I2S rate set to %d Hz\n", d.Serial(), hz)
	}

	return nil
}
