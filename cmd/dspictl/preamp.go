package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newPreampCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "preamp [<channel> [<db>]]",
		Short:             "Per-channel input preamp get/set",
		Args:              cobra.RangeArgs(0, 2),
		RunE:              runPreamp,
		ValidArgsFunction: completeInputChannels,
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "get [channel]",
		Short:             "Show all channels, or one channel's preamp (alias for `preamp [channel]`)",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runPreampGet,
		ValidArgsFunction: completeInputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "set <channel> <db>",
		Short:             "Set preamp gain for a channel (alias for `preamp <channel> <db>`)",
		Args:              cobra.ExactArgs(2),
		RunE:              runPreampSet,
		ValidArgsFunction: completeInputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "global [<db>]",
		Short: "Get or set the global preamp (applied before per-channel preamp)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPreampGlobal,
	})

	return cmd
}

func runPreamp(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 0, 1:
		return runPreampGet(cmd, args)
	case 2:
		return runPreampSet(cmd, args)
	default:
		return fmt.Errorf("expected 0–2 arguments, got %d", len(args))
	}
}

func runPreampGet(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	// Only preamp channels are input channels (0=USB L, 1=USB R)
	numChannels := 2

	for _, d := range devices {
		channels, err := d.Channels()

		if err != nil {
			slog.Error("getting channels failed", "serial", d.Serial(), "error", err)

			continue
		}

		if len(args) == 0 {
			// Show all input channels
			for _, ch := range channels {
				if ch.Index >= numChannels {
					break
				}

				gain, err := d.GetPreampChannel(ch.Index)

				if err != nil {
					slog.Error("getting preamp failed", "serial", d.Serial(), "channel", ch.Index, "error", err)

					continue
				}

				fmt.Printf("%s: ch %d %s: %s\n", d.Serial(), ch.Index, ch.Name, gain)
			}
		} else {
			ch, err := strconv.Atoi(args[0])

			if err != nil {
				return fmt.Errorf("invalid channel: %w", err)
			}

			gain, err := d.GetPreampChannel(ch)

			if err != nil {
				return fmt.Errorf("%s: %w", d.Serial(), err)
			}

			name := channels[ch].Name
			fmt.Printf("%s: ch %d %s: %s\n", d.Serial(), ch, name, gain)
		}
	}

	return nil
}

func runPreampSet(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	db, err := strconv.ParseFloat(args[1], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetPreampChannel(ch, dspi.NewGain(db))

		if err != nil {
			slog.Error("setting preamp failed", "serial", d.Serial(), "channel", ch, "error", err)

			continue
		}

		channels, err := d.Channels()

		if err != nil {
			slog.Error("getting channels failed", "serial", d.Serial(), "error", err)

			continue
		}

		name := channels[ch].Name
		fmt.Printf("%s: ch %d %s: %s\n", d.Serial(), ch, name, dspi.NewGain(db))
	}

	return nil
}

func runPreampGlobal(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			gain, err := d.GetPreamp()

			if err != nil {
				slog.Error("getting global preamp failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: global preamp: %s\n", d.Serial(), gain)
		}

		return nil
	}

	db, err := strconv.ParseFloat(args[0], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	for _, d := range devices {
		err := d.SetPreamp(db)

		if err != nil {
			slog.Error("setting global preamp failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: global preamp: %s\n", d.Serial(), dspi.NewGain(db))
	}

	return nil
}
