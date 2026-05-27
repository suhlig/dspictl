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

	cmd.AddCommand(&cobra.Command{
		Use:   "reference [spl]",
		Short: "Get or set reference SPL (40–100 dB)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLoudnessReference,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "intensity [pct]",
		Short: "Get or set intensity (0–200%)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLoudnessIntensity,
	})

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
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
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

	spl, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid SPL value: %w", err)
	}

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
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
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

	pct, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid intensity value: %w", err)
	}

	for _, d := range devices {
		if err := d.SetLoudnessIntensity(pct); err != nil {
			slog.Error("setting loudness intensity failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: intensity = %.0f%%\n", d.Serial(), pct)
	}

	return nil
}
