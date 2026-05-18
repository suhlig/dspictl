package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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
	return newRootCmd().ExecuteContext(ctx)
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
	rootCmd.AddCommand(newVolumeCmd())
	rootCmd.AddCommand(newPreampCmd())
	rootCmd.AddCommand(newOutputCmd())
	rootCmd.AddCommand(newPresetCmd())
	rootCmd.AddCommand(newMatrixCmd())

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
