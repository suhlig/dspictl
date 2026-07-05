package main

import (
	"fmt"
	"log/slog"
	"strconv"

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

	cmd.AddCommand(&cobra.Command{
		Use:               "channels [<2|4|6|8>]",
		Short:             "Get or set the number of I2S input channels",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputChannels,
		ValidArgsFunction: completeChoices([]string{"2", "4", "6", "8"}),
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

func runInputChannels(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			ch, err := d.GetI2SInputChannels()

			if err != nil {
				slog.Error("getting I2S input channels failed", "serial", d.Serial(), "error", err)

				continue
			}

			if ch == 0 {
				fmt.Printf("%s: I2S input channels not configured\n", d.Serial())
			} else {
				fmt.Printf("%s: %d I2S input channels\n", d.Serial(), ch)
			}
		}

		return nil
	}

	n, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel count: %w", err)
	}

	switch n {
	case 2, 4, 6, 8:
	default:
		return fmt.Errorf("invalid channel count: %d (expected 2, 4, 6, or 8)", n)
	}

	for _, d := range devices {
		if err := d.SetI2SInputChannels(n); err != nil {
			slog.Error("setting I2S input channels failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: I2S input channels set to %d\n", d.Serial(), n)
	}

	return nil
}
