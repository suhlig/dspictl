package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

func newChannelNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel-name",
		Short: "Read or write user-configurable channel names",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get [<channel>]",
		Short: "Show all channel names, or one channel",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runChannelNameGet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <channel> <name>",
		Short: "Set a channel name (max 31 chars)",
		Args:  cobra.ExactArgs(2),
		RunE:  runChannelNameSet,
	})

	return cmd
}

func runChannelNameGet(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			channels, err := d.Channels()

			if err != nil {
				slog.Error("getting channels failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s:\n", d.Serial())

			for _, ch := range channels {
				fmt.Printf("  ch %d: %s\n", ch.Index, ch.Name)
			}
		}

		return nil
	}

	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	for _, d := range devices {
		name, err := d.ChannelName(ch)

		if err != nil {
			slog.Error("getting channel name failed", "serial", d.Serial(), "channel", ch, "error", err)

			continue
		}

		fmt.Printf("%s: ch %d: %s\n", d.Serial(), ch, name)
	}

	return nil
}

func runChannelNameSet(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	name := args[1]

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetChannelName(ch, name)

		if err != nil {
			slog.Error("setting channel name failed", "serial", d.Serial(), "channel", ch, "error", err)

			continue
		}

		fmt.Printf("%s: ch %d: %s\n", d.Serial(), ch, name)
	}

	return nil
}
