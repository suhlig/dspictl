package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newLoudnessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loudness",
		Short: "Loudness compensation (ISO 226:2003 equal-loudness contours)",
		RunE:  runLoudnessStatus,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "on",
		Short: "Enable loudness compensation",
		RunE:  runLoudnessOn,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Disable loudness compensation",
		RunE:  runLoudnessOff,
	})

	refCmd := &cobra.Command{
		Use:   "reference [spl]",
		Short: "Get or set reference SPL (40–100 dB)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLoudnessReference,
	}
	refCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Show current reference SPL (alias for `reference` with no args)",
		RunE:  runLoudnessReferenceGet,
	})
	refCmd.AddCommand(&cobra.Command{
		Use:   "set <spl>",
		Short: "Set reference SPL (alias for `reference <spl>`)",
		Args:  cobra.ExactArgs(1),
		RunE:  runLoudnessReferenceSet,
	})
	cmd.AddCommand(refCmd)

	intCmd := &cobra.Command{
		Use:   "intensity [pct]",
		Short: "Get or set intensity (0–200%)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLoudnessIntensity,
	}
	intCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Show current intensity (alias for `intensity` with no args)",
		RunE:  runLoudnessIntensityGet,
	})
	intCmd.AddCommand(&cobra.Command{
		Use:   "set <pct>",
		Short: "Set intensity (alias for `intensity <pct>`)",
		Args:  cobra.ExactArgs(1),
		RunE:  runLoudnessIntensitySet,
	})
	cmd.AddCommand(intCmd)

	outCmd := &cobra.Command{
		Use:   "outputs [on|off] [<channels...>]",
		Short: "Get or set per-output loudness mask",
		Long: `Get or set which output channels are compensated.

With no arguments, shows the current active outputs.
With "on" or "off" followed by channel numbers, toggles specific outputs.
With a preset name, sets the mask to a predefined value.

Presets:
  all   – all outputs (default)
  none  – disable all outputs`,
		Args: cobra.ArbitraryArgs,
		RunE: runLoudnessOutputs,
	}
	cmd.AddCommand(outCmd)

	return cmd
}

func runLoudnessStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		printLoudnessStatus(d)
	}

	return nil
}

// printLoudnessCompact prints a one-line loudness summary for use in status output.
func printLoudnessCompact(d *dspi.Device) {
	enabled, err := d.GetLoudness()
	if err != nil {
		return // silently skip if firmware doesn't support it
	}

	if !enabled {
		fmt.Printf("  Loudness: disabled\n")
		return
	}

	ref, err := d.GetLoudnessReference()
	if err != nil {
		return
	}

	intensity, err := d.GetLoudnessIntensity()
	if err != nil {
		return
	}

	fmt.Printf("  Loudness: enabled, ref=%.0f dB SPL, intensity=%.0f%%\n", ref, intensity)

	mask, err := d.GetLoudnessOutputMask()
	if err == nil {
		fmt.Printf("  Active outputs: %s\n", formatMaskU16(mask, 16))
	}
}

func printLoudnessStatus(d *dspi.Device) {
	enabled, err := d.GetLoudness()
	if err != nil {
		slog.Error("getting loudness failed", "serial", d.Serial(), "error", err)
		return
	}

	ref, err := d.GetLoudnessReference()
	if err != nil {
		slog.Error("getting loudness reference failed", "serial", d.Serial(), "error", err)
		return
	}

	intensity, err := d.GetLoudnessIntensity()
	if err != nil {
		slog.Error("getting loudness intensity failed", "serial", d.Serial(), "error", err)
		return
	}

	state := "disabled"
	if enabled {
		state = "enabled"
	}

	fmt.Printf("%s:\n", d.Serial())
	fmt.Printf("  Loudness: %s\n", state)
	fmt.Printf("  Reference SPL: %.0f dB\n", ref)
	fmt.Printf("  Intensity: %.0f%%\n", intensity)

	mask, err := d.GetLoudnessOutputMask()
	if err == nil {
		fmt.Printf("  Active outputs: %s\n", formatMaskU16(mask, 16))
	}
}

func runLoudnessOn(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetLoudness(true); err != nil {
			slog.Error("enabling loudness failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: loudness enabled\n", d.Serial())
	}

	return nil
}

func runLoudnessOff(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetLoudness(false); err != nil {
			slog.Error("disabling loudness failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: loudness disabled\n", d.Serial())
	}

	return nil
}

func runLoudnessReference(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runLoudnessReferenceGet(cmd, args)
	}

	return runLoudnessReferenceSet(cmd, args)
}

func runLoudnessReferenceGet(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		ref, err := d.GetLoudnessReference()
		if err != nil {
			slog.Error("getting loudness reference failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: reference SPL = %.0f dB\n", d.Serial(), ref)
	}

	return nil
}

func runLoudnessReferenceSet(cmd *cobra.Command, args []string) error {
	spl, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid SPL value: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetLoudnessReference(spl); err != nil {
			slog.Error("setting loudness reference failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: reference SPL = %.0f dB\n", d.Serial(), spl)
	}

	return nil
}

func runLoudnessIntensity(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runLoudnessIntensityGet(cmd, args)
	}

	return runLoudnessIntensitySet(cmd, args)
}

func runLoudnessIntensityGet(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		intensity, err := d.GetLoudnessIntensity()
		if err != nil {
			slog.Error("getting loudness intensity failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: intensity = %.0f%%\n", d.Serial(), intensity)
	}

	return nil
}

func runLoudnessIntensitySet(cmd *cobra.Command, args []string) error {
	pct, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid intensity value: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetLoudnessIntensity(pct); err != nil {
			slog.Error("setting loudness intensity failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: intensity = %.0f%%\n", d.Serial(), pct)
	}

	return nil
}

func runLoudnessOutputs(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		// Show status
		for _, d := range devices {
			mask, err := d.GetLoudnessOutputMask()
			if err != nil {
				slog.Error("getting loudness output mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: loudness active outputs: %s\n", d.Serial(), formatMaskU16(mask, 16))
		}
		return nil
	}

	// Named presets
	switch args[0] {
	case "all":
		for _, d := range devices {
			if err := d.SetLoudnessOutputMask(0xFFFF); err != nil {
				slog.Error("setting loudness output mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: loudness active outputs: all (16)\n", d.Serial())
		}
		return nil
	case "none":
		for _, d := range devices {
			if err := d.SetLoudnessOutputMask(0x0000); err != nil {
				slog.Error("setting loudness output mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: loudness active outputs: none\n", d.Serial())
		}
		return nil
	}

	// on/off with channel numbers
	if args[0] != "on" && args[0] != "off" {
		return fmt.Errorf("expected \"on\", \"off\", \"all\", or \"none\", got %q", args[0])
	}
	enable := args[0] == "on"
	channelArgs := args[1:]
	if len(channelArgs) == 0 {
		return fmt.Errorf("%s requires at least one channel number", args[0])
	}

	for _, d := range devices {
		current, err := d.GetLoudnessOutputMask()
		if err != nil {
			slog.Error("getting loudness output mask failed", "serial", d.Serial(), "error", err)
			continue
		}

		var newMask uint16
		if enable {
			newMask, err = maskSetBits(current, channelArgs, 16)
		} else {
			newMask, err = maskClearBits(current, channelArgs, 16)
		}
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if err := d.SetLoudnessOutputMask(newMask); err != nil {
			slog.Error("setting loudness output mask failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: loudness active outputs: %s\n", d.Serial(), formatMaskU16(newMask, 16))
	}

	return nil
}
