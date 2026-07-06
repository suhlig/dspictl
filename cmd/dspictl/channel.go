package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Per-channel gain, mute, and delay control",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "gain <channel> [<db>]",
		Short:             "Get or set channel gain",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runChannelGain,
		ValidArgsFunction: completeInputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "mute <channel>",
		Short:             "Mute a channel",
		Args:              cobra.ExactArgs(1),
		RunE:              runChannelMute,
		ValidArgsFunction: completeInputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "unmute <channel>",
		Short:             "Unmute a channel",
		Args:              cobra.ExactArgs(1),
		RunE:              runChannelUnmute,
		ValidArgsFunction: completeInputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "delay <channel> [<ms>]",
		Short:             "Get or set channel delay in milliseconds",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runChannelDelay,
		ValidArgsFunction: completeInputChannels,
	})

	return cmd
}

func runChannelGain(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 1 {
		for _, d := range devices {
			gain, err := d.GetChannelGain(ch)
			if err != nil {
				slog.Error("getting channel gain failed", "serial", d.Serial(), "channel", ch, "error", err)
				continue
			}
			fmt.Printf("%s: channel %d gain: %s\n", d.Serial(), ch, gain)
		}
		return nil
	}

	db, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	for _, d := range devices {
		err := d.SetChannelGain(ch, dspi.NewGain(db))
		if err != nil {
			slog.Error("setting channel gain failed", "serial", d.Serial(), "channel", ch, "error", err)
			continue
		}
		fmt.Printf("%s: channel %d gain: %s\n", d.Serial(), ch, dspi.NewGain(db))
	}

	return nil
}

func runChannelMute(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetChannelMute(ch, true)
		if err != nil {
			slog.Error("muting channel failed", "serial", d.Serial(), "channel", ch, "error", err)
			continue
		}
		fmt.Printf("%s: channel %d muted\n", d.Serial(), ch)
	}

	return nil
}

func runChannelUnmute(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetChannelMute(ch, false)
		if err != nil {
			slog.Error("unmuting channel failed", "serial", d.Serial(), "channel", ch, "error", err)
			continue
		}
		fmt.Printf("%s: channel %d unmuted\n", d.Serial(), ch)
	}

	return nil
}

func runChannelDelay(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 1 {
		for _, d := range devices {
			delay, err := d.GetChannelDelay(ch)
			if err != nil {
				slog.Error("getting channel delay failed", "serial", d.Serial(), "channel", ch, "error", err)
				continue
			}
			fmt.Printf("%s: channel %d delay: %.2f ms\n", d.Serial(), ch, delay)
		}
		return nil
	}

	ms, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid delay value: %w", err)
	}

	for _, d := range devices {
		err := d.SetChannelDelay(ch, ms)
		if err != nil {
			slog.Error("setting channel delay failed", "serial", d.Serial(), "channel", ch, "error", err)
			continue
		}
		fmt.Printf("%s: channel %d delay: %.2f ms\n", d.Serial(), ch, ms)
	}

	return nil
}
