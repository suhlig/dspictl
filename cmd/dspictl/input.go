package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Input source selection and I2S configuration",
	}

	selectCmd := &cobra.Command{
		Use:   "select-i2s",
		Short: "Switch to I2S input (orchestrates pin, rate, source)",
		RunE:  runInputSelectI2S,
	}

	selectCmd.Flags().String("rate", "48000", "I2S sample rate (44100, 48000, 96000)")
	selectCmd.Flags().String("pin", "", "I2S RX data GPIO pin")
	selectCmd.Flags().String("bck-pin", "", "I2S BCK GPIO pin")

	cmd.AddCommand(selectCmd)

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
		Use:   "i2s-rx-pin [<gpio>]",
		Short: "Get or set the I2S RX data GPIO pin",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInputI2SRxPin,
	})

	return cmd
}

func runInputSelectI2S(cmd *cobra.Command, args []string) error {
	rateStr, err := cmd.Flags().GetString("rate")

	if err != nil {
		return err
	}

	pinStr, err := cmd.Flags().GetString("pin")

	if err != nil {
		return err
	}

	bckPinStr, err := cmd.Flags().GetString("bck-pin")

	if err != nil {
		return err
	}

	var hz uint32

	switch rateStr {
	case "44100":
		hz = 44100
	case "48000":
		hz = 48000
	case "96000":
		hz = 96000
	default:
		return fmt.Errorf("invalid rate: %s (expected 44100, 48000, or 96000)", rateStr)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		if pinStr != "" {
			pin, err := strconv.Atoi(pinStr)

			if err != nil {
				return fmt.Errorf("invalid pin: %w", err)
			}

			if err := d.SetI2SRxPin(pin); err != nil {
				slog.Error("setting I2S RX pin failed", "serial", d.Serial(), "error", err)

				continue
			}
		}

		if bckPinStr != "" {
			pin, err := strconv.Atoi(bckPinStr)

			if err != nil {
				return fmt.Errorf("invalid BCK pin: %w", err)
			}

			if err := d.SetI2SBckPin(pin); err != nil {
				slog.Error("setting I2S BCK pin failed", "serial", d.Serial(), "error", err)

				continue
			}
		}

		if err := d.SetInputRate(hz); err != nil {
			slog.Error("setting input rate failed", "serial", d.Serial(), "error", err)

			continue
		}

		if err := d.SetInputSource(dspi.InputSourceI2S); err != nil {
			slog.Error("setting input source failed", "serial", d.Serial(), "error", err)

			continue
		}

		// Poll until the switch applies.
		settled := false

		for range 50 {
			src, err := d.GetInputSource()

			if err == nil && src == dspi.InputSourceI2S {
				settled = true

				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		if !settled {
			slog.Error("I2S input did not settle in time", "serial", d.Serial())

			continue
		}

		fmt.Printf("%s: I2S input selected at %d Hz\n", d.Serial(), hz)
	}

	return nil
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

func runInputI2SRxPin(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			pin, err := d.GetI2SRxPin()

			if err != nil {
				slog.Error("getting I2S RX pin failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: I2S RX pin=GPIO %d\n", d.Serial(), pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		if err := d.SetI2SRxPin(pin); err != nil {
			slog.Error("setting I2S RX pin failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: I2S RX pin=GPIO %d\n", d.Serial(), pin)
	}

	return nil
}
