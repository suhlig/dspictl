package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

func newChannelNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "channel-name [<channel> [<name>]]",
		Short:             "Read or write user-configurable channel names",
		Args:              cobra.RangeArgs(0, 2),
		RunE:              runChannelName,
		ValidArgsFunction: completeChannels,
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "get [<channel>]",
		Short:             "Show all channel names, or one channel (alias for `channel-name [channel]`)",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runChannelNameGet,
		ValidArgsFunction: completeChannels,
	})

	channelNameSetCmd := &cobra.Command{
		Use:               "set <channel> [<name>]",
		Short:             "Set a channel name (max 31 chars) (alias for `channel-name <channel> <name>`)",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runChannelNameSet,
		ValidArgsFunction: completeChannels,
	}
	channelNameSetCmd.Flags().String("name", "", "Channel name (max 31 chars)")
	cmd.AddCommand(channelNameSetCmd)

	return cmd
}

// completeChannels returns channel indices with their current names for shell completion.
func completeChannels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	devices, err := openDevices()

	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	defer closeDevices(devices)

	seen := make(map[string]bool)
	var completions []string

	for _, d := range devices {
		channels, err := d.Channels()

		if err != nil {
			continue
		}

		for _, ch := range channels {
			item := fmt.Sprintf("%d\t%s", ch.Index, ch.Name)

			if !seen[item] {
				seen[item] = true
				completions = append(completions, item)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func runChannelName(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 0, 1:
		return runChannelNameGet(cmd, args)
	case 2:
		return runChannelNameSet(cmd, args)
	default:
		return fmt.Errorf("expected 0–2 arguments, got %d", len(args))
	}
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

	// Check --name flag first (canonical form), fall back to positional arg
	name, err := cmd.Flags().GetString("name")

	if err != nil {
		return fmt.Errorf("getting name flag: %w", err)
	}

	if name == "" && len(args) > 1 {
		name = args[1]
	}

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
