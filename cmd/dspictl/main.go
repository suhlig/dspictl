package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

var targetSerial string

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
	}

	rootCmd.PersistentFlags().StringVar(&targetSerial, "target", "", "Operate on a specific device by serial number")

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

	return rootCmd
}

func closeDevices(devices []*dspi.Device) {
	for _, d := range devices {
		d.Close()
	}
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
