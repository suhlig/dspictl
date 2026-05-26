package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
	"github.com/suhlig/dspi/cmd/dspictl/mixer"
)

var targetSerial string

var version = "unknown"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})))

	err := mainE(context.Background())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func mainE(ctx context.Context) error {
	cmd := newRootCmd()
	cmd.SetArgs(protectNegativeArgs(os.Args[1:]))

	return cmd.ExecuteContext(ctx)
}

var knownValueTakingFlags = map[string]bool{
	"--target": true,
}

var negativeNumberRe = regexp.MustCompile(`^-\d+(\.\d+)?$`)

// protectNegativeArgs inserts -- before negative numbers that look like dB
// values, so cobra/pflag doesn't interpret them as shorthand flags.
//
// For example, ["volume", "set", "-20"] becomes ["volume", "set", "--", "-20"].
// Negative numbers preceded by a known value-taking flag (like --target) are
// left alone.
func protectNegativeArgs(args []string) []string {
	var result []string
	afterDoubleDash := false

	for i, arg := range args {
		if arg == "--" {
			afterDoubleDash = true

			result = append(result, arg)

			continue
		}

		if !afterDoubleDash && negativeNumberRe.MatchString(arg) {
			if i > 0 && knownValueTakingFlags[args[i-1]] {
				result = append(result, arg)

				continue
			}

			result = append(result, "--")
		}

		result = append(result, arg)
	}

	return result
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "dspictl",
		Short:         "Control DSPi audio devices",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	rootCmd.PersistentFlags().StringVar(&targetSerial, "target", "", "Operate on a specific device by serial number")
	_ = rootCmd.RegisterFlagCompletionFunc("target", completeSerialNumbers)

	rootCmd.AddCommand(newMuteCmd())
	rootCmd.AddCommand(newUnmuteCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newClearClipsCmd())
	rootCmd.AddCommand(newBufferStatsCmd())
	rootCmd.AddCommand(newUSBErrorsCmd())
	rootCmd.AddCommand(newCore1Cmd())
	rootCmd.AddCommand(newBootloaderCmd())
	rootCmd.AddCommand(newFactoryResetCmd())
	rootCmd.AddCommand(newVolumeCmd())
	rootCmd.AddCommand(newPreampCmd())
	rootCmd.AddCommand(newOutputCmd())
	rootCmd.AddCommand(newPresetCmd())
	rootCmd.AddCommand(newMatrixCmd())
	rootCmd.AddCommand(newChannelNameCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newEQCmd())
	rootCmd.AddCommand(newTargetsCmd())
	rootCmd.AddCommand(mixer.NewCmd())

	return rootCmd
}

func closeDevices(devices []*dspi.Device) {
	for _, d := range devices {
		d.Close()
	}
}

// completeSerialNumbers returns connected DSPi serial numbers for shell completion.
func completeSerialNumbers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	devices, err := dspi.List()

	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	serials := make([]string, 0, len(devices))

	for _, d := range devices {
		serials = append(serials, d.Serial)
	}

	return serials, cobra.ShellCompDirectiveNoFileComp
}

// completeChoices returns a ValidArgsFunction that suggests values per argument position.
func completeChoices(choices ...[]string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) < len(choices) {
			return choices[len(args)], cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeInputChannels returns input channel indices (0-1) with names for shell completion.
func completeInputChannels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return completeChannelIndices(func(ch dspi.ChannelInfo) bool { return ch.Index <= 1 })
}

// completeEQMasterChannels returns master EQ channel indices (0-1) with names.
func completeEQMasterChannels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return completeChannelIndices(func(ch dspi.ChannelInfo) bool { return ch.Index <= 1 })
}

// completeEQMasterChannelsAndBands returns master channels for arg 0, bands 0-9 for arg 1.
func completeEQMasterChannelsAndBands(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return completeChannelIndices(func(ch dspi.ChannelInfo) bool { return ch.Index <= 1 })
	case 1:
		var bands []string
		for i := range 10 {
			bands = append(bands, fmt.Sprintf("%d", i))
		}
		return bands, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeEQOutputChannelsAndBands returns output channels for arg 0, bands 0-9 for arg 1.
func completeEQOutputChannelsAndBands(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return completeOutputChannels(cmd, args, toComplete)
	case 1:
		var bands []string
		for i := range 10 {
			bands = append(bands, fmt.Sprintf("%d", i))
		}
		return bands, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeOutputChannels returns output channel indices with names for shell completion.
func completeOutputChannels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
			if ch.Index < 2 {
				continue
			}

			outputIndex := ch.Index - 2
			item := fmt.Sprintf("%d\t%s", outputIndex, ch.Name)

			if !seen[item] {
				seen[item] = true
				completions = append(completions, item)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeMatrixRoutes returns input (0-1) or output indices with names for matrix commands.
func completeMatrixRoutes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	devices, err := openDevices()

	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	defer closeDevices(devices)

	seen := make(map[string]bool)
	var completions []string

	switch len(args) {
	case 0:
		for _, d := range devices {
			channels, err := d.Channels()

			if err != nil {
				continue
			}

			for _, ch := range channels {
				if ch.Index > 1 {
					continue
				}

				item := fmt.Sprintf("%d\t%s", ch.Index, ch.Name)

				if !seen[item] {
					seen[item] = true
					completions = append(completions, item)
				}
			}
		}
	case 1:
		for _, d := range devices {
			channels, err := d.Channels()

			if err != nil {
				continue
			}

			for _, ch := range channels {
				if ch.Index < 2 {
					continue
				}

				outputIndex := ch.Index - 2
				item := fmt.Sprintf("%d\t%s", outputIndex, ch.Name)

				if !seen[item] {
					seen[item] = true
					completions = append(completions, item)
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeChannelIndices returns channel completions matching the filter.
func completeChannelIndices(filter func(dspi.ChannelInfo) bool) ([]string, cobra.ShellCompDirective) {
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
			if !filter(ch) {
				continue
			}

			item := fmt.Sprintf("%d\t%s", ch.Index, ch.Name)

			if !seen[item] {
				seen[item] = true
				completions = append(completions, item)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func openDevices() ([]*dspi.Device, error) {
	if targetSerial != "" {
		dev, err := dspi.Open(dspi.DeviceInfo{Serial: targetSerial})

		if err != nil {
			return nil, err
		}

		return []*dspi.Device{dev}, nil
	}

	return dspi.OpenAll()
}
