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
		Use:   "preamp",
		Short: "Per-channel input preamp get/set",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get [channel]",
		Short: "Show preamp for all channels, or one channel",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPreampGet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <channel> <db>",
		Short: "Set preamp gain for a channel",
		Args:  cobra.ExactArgs(2),
		RunE:  runPreampSet,
	})

	return cmd
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
