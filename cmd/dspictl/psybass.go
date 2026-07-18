package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newPsybassCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "psybass",
		Short: "Psychoacoustic bass enhancement (missing-fundamental harmonics)",
		RunE:  runPsybassStatus,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "on",
		Short: "Enable psychoacoustic bass",
		RunE:  runPsybassOn,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Disable psychoacoustic bass",
		RunE:  runPsybassOff,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "cutoff [hz]",
		Short: "Get or set speaker LF limit in Hz",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPsybassFloat,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "harmonics [db]",
		Short: "Get or set harmonic mix level in dB",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPsybassFloat,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "drive [db]",
		Short: "Get or set odd-path clipper drive in dB",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPsybassFloat,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "character [pct]",
		Short: "Get or set even/odd harmonic blend percentage",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPsybassFloat,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "original [db]",
		Short: "Get or set original low-band level in dB",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPsybassFloat,
	})

	outCmd := &cobra.Command{
		Use:   "outputs [on|off] [<channels...>]",
		Short: "Get or set per-output psybass mask",
		Long: `Get or set which output channels are processed.

With no arguments, shows the current active outputs.
With "on" or "off" followed by channel numbers, toggles specific outputs.
With a preset name, sets the mask to a predefined value.

Presets:
  all   – all outputs (default)
  none  – disable all outputs`,
		Args: cobra.ArbitraryArgs,
		RunE: runPsybassOutputs,
	}
	cmd.AddCommand(outCmd)

	return cmd
}

func runPsybassStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		printPsybassStatus(d)
	}

	return nil
}

func printPsybassStatus(d *dspi.Device) {
	enabled, mask, err := d.GetPsybass()
	if err != nil {
		slog.Error("getting psybass status failed", "serial", d.Serial(), "error", err)
		return
	}

	state := "disabled"
	if enabled {
		state = "enabled"
	}

	fmt.Printf("%s:\n", d.Serial())
	fmt.Printf("  Psybass: %s\n", state)
	fmt.Printf("  Active outputs: %s\n", formatMaskU16(mask, 16))

	printPsybassFloatIfOk(d, "cutoff", d.GetPsybassCutoff, "Hz")
	printPsybassFloatIfOk(d, "harmonics", d.GetPsybassHarmonics, "dB")
	printPsybassFloatIfOk(d, "drive", d.GetPsybassDrive, "dB")
	printPsybassFloatIfOk(d, "character", d.GetPsybassCharacter, "%")
	printPsybassFloatIfOk(d, "original", d.GetPsybassOriginal, "dB")
}

func printPsybassFloatIfOk(d *dspi.Device, label string, getter func() (float32, error), unit string) {
	v, err := getter()
	if err != nil {
		slog.Error("getting psybass parameter failed", "serial", d.Serial(), "param", label, "error", err)
		return
	}
	fmt.Printf("  %s: %.2f %s\n", label, v, unit)
}

func runPsybassOn(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetPsybass(true, 0xFFFF); err != nil {
			slog.Error("enabling psybass failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: psybass enabled\n", d.Serial())
	}

	return nil
}

func runPsybassOff(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetPsybass(false, 0); err != nil {
			slog.Error("disabling psybass failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: psybass disabled\n", d.Serial())
	}

	return nil
}

func runPsybassFloat(cmd *cobra.Command, args []string) error {
	label := cmd.Name()

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			v, err := getPsybassValue(d, label)
			if err != nil {
				slog.Error("getting psybass parameter failed", "serial", d.Serial(), "param", label, "error", err)
				continue
			}
			fmt.Printf("%s: psybass %s = %.2f\n", d.Serial(), label, v)
		}
		return nil
	}

	v, err := strconv.ParseFloat(args[0], 32)
	if err != nil {
		return fmt.Errorf("invalid %s value: %w", label, err)
	}

	for _, d := range devices {
		if err := setPsybassValue(d, label, float32(v)); err != nil {
			slog.Error("setting psybass parameter failed", "serial", d.Serial(), "param", label, "error", err)
			continue
		}
		fmt.Printf("%s: psybass %s = %.2f\n", d.Serial(), label, v)
	}

	return nil
}

func getPsybassValue(d *dspi.Device, param string) (float32, error) {
	switch param {
	case "cutoff":
		return d.GetPsybassCutoff()
	case "harmonics":
		return d.GetPsybassHarmonics()
	case "drive":
		return d.GetPsybassDrive()
	case "character":
		return d.GetPsybassCharacter()
	case "original":
		return d.GetPsybassOriginal()
	}
	return 0, fmt.Errorf("unknown psybass parameter: %s", param)
}

func setPsybassValue(d *dspi.Device, param string, v float32) error {
	switch param {
	case "cutoff":
		return d.SetPsybassCutoff(v)
	case "harmonics":
		return d.SetPsybassHarmonics(v)
	case "drive":
		return d.SetPsybassDrive(v)
	case "character":
		return d.SetPsybassCharacter(v)
	case "original":
		return d.SetPsybassOriginal(v)
	}
	return fmt.Errorf("unknown psybass parameter: %s", param)
}

func runPsybassOutputs(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mask, err := d.GetPsybassMask()
			if err != nil {
				slog.Error("getting psybass mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: psybass active outputs: %s\n", d.Serial(), formatMaskU16(mask, 16))
		}
		return nil
	}

	switch args[0] {
	case "all":
		for _, d := range devices {
			if err := d.SetPsybassMask(0xFFFF); err != nil {
				slog.Error("setting psybass mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: psybass active outputs: all (16)\n", d.Serial())
		}
		return nil
	case "none":
		for _, d := range devices {
			if err := d.SetPsybassMask(0x0000); err != nil {
				slog.Error("setting psybass mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: psybass active outputs: none\n", d.Serial())
		}
		return nil
	}

	if args[0] != "on" && args[0] != "off" {
		return fmt.Errorf("expected \"on\", \"off\", \"all\", or \"none\", got %q", args[0])
	}
	enable := args[0] == "on"
	channelArgs := args[1:]
	if len(channelArgs) == 0 {
		return fmt.Errorf("%s requires at least one channel number", args[0])
	}

	for _, d := range devices {
		current, err := d.GetPsybassMask()
		if err != nil {
			slog.Error("getting psybass mask failed", "serial", d.Serial(), "error", err)
			continue
		}

		var newMask uint16
		if enable {
			newMask, err = maskSetBits(current, channelArgs, 16)
		} else {
			newMask, err = maskClearBits(current, channelArgs, 16)
		}
		if err != nil {
			return err
		}

		if err := d.SetPsybassMask(newMask); err != nil {
			slog.Error("setting psybass mask failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: psybass active outputs: %s\n", d.Serial(), formatMaskU16(newMask, 16))
	}

	return nil
}
