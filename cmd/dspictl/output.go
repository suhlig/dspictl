package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Per-output gain, mute, delay, enable/disable",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Show all outputs with gain, mute, delay, enable",
		RunE:  runOutputList,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "gain <channel> <db>",
		Short: "Set output gain",
		Args:  cobra.ExactArgs(2),
		RunE:  runOutputGain,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "mute <channel>",
		Short: "Mute output",
		Args:  cobra.ExactArgs(1),
		RunE:  runOutputMute,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "unmute <channel>",
		Short: "Unmute output",
		Args:  cobra.ExactArgs(1),
		RunE:  runOutputUnmute,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delay <channel> <ms>",
		Short: "Set time alignment delay (0-85 ms)",
		Args:  cobra.ExactArgs(2),
		RunE:  runOutputDelay,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "enable <channel>",
		Short: "Enable output channel",
		Args:  cobra.ExactArgs(1),
		RunE:  runOutputEnable,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "disable <channel>",
		Short: "Disable output channel",
		Args:  cobra.ExactArgs(1),
		RunE:  runOutputDisable,
	})

	return cmd
}

func printOutputState(d *dspi.Device, ch int) {
	gain, err := d.GetOutputGain(ch)

	if err != nil {
		slog.Error("getting output gain failed", "serial", d.Serial(), "channel", ch, "error", err)

		return
	}

	muted, err := d.GetOutputMute(ch)

	if err != nil {
		slog.Error("getting output mute failed", "serial", d.Serial(), "channel", ch, "error", err)

		return
	}

	delay, err := d.GetOutputDelay(ch)

	if err != nil {
		slog.Error("getting output delay failed", "serial", d.Serial(), "channel", ch, "error", err)

		return
	}

	enabled, err := d.GetOutputEnable(ch)

	if err != nil {
		slog.Error("getting output enable failed", "serial", d.Serial(), "channel", ch, "error", err)

		return
	}

	muteStr := ""

	if muted {
		muteStr = " (muted)"
	}

	enableStr := "enabled"

	if !enabled {
		enableStr = "disabled"
	}

	fmt.Printf("  ch %d: gain %s  delay %5.1f ms  %s%s\n", ch, gain, delay, enableStr, muteStr)
}

func runOutputList(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		fmt.Printf("%s:\n", d.Serial())

		numOutputs := 5

		if d.Platform() == dspi.PlatformRP2350 {
			numOutputs = 9
		}

		for ch := 0; ch < numOutputs; ch++ {
			printOutputState(d, ch)
		}
	}

	return nil
}

func runOutputGain(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	db, err := strconv.ParseFloat(args[1], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	return setOutputForDevices(func(d *dspi.Device) error {
		return d.SetOutputGain(ch, dspi.NewGain(db))
	}, fmt.Sprintf("ch %d gain: %s", ch, dspi.NewGain(db)))
}

func runOutputMute(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return setOutputForDevices(func(d *dspi.Device) error {
		return d.SetOutputMute(ch, true)
	}, fmt.Sprintf("ch %d: muted", ch))
}

func runOutputUnmute(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return setOutputForDevices(func(d *dspi.Device) error {
		return d.SetOutputMute(ch, false)
	}, fmt.Sprintf("ch %d: unmuted", ch))
}

func runOutputDelay(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	ms, err := strconv.ParseFloat(args[1], 64)

	if err != nil {
		return fmt.Errorf("invalid ms value: %w", err)
	}

	return setOutputForDevices(func(d *dspi.Device) error {
		return d.SetOutputDelay(ch, ms)
	}, fmt.Sprintf("ch %d delay: %.1f ms", ch, ms))
}

func runOutputEnable(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return setOutputForDevices(func(d *dspi.Device) error {
		return d.SetOutputEnable(ch, true)
	}, fmt.Sprintf("ch %d: enabled", ch))
}

func runOutputDisable(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return setOutputForDevices(func(d *dspi.Device) error {
		return d.SetOutputEnable(ch, false)
	}, fmt.Sprintf("ch %d: disabled", ch))
}

func setOutputForDevices(fn func(*dspi.Device) error, msg string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := fn(d)

		if err != nil {
			slog.Error("output command failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), msg)
	}

	return nil
}
