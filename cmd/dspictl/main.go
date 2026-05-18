package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

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

	rootCmd.AddCommand(newGetMasterVolumeCmd())

	return rootCmd
}

func newGetMasterVolumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-master-volume",
		Short: "Print the current master volume of connected DSPi devices",
		RunE:  runGetMasterVolume,
	}
}

func runGetMasterVolume(cmd *cobra.Command, args []string) error {
	devices, err := dspi.OpenAll()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer func() {
		for _, d := range devices {
			d.Close()
		}
	}()

	for _, d := range devices {
		gain, err := d.GetMasterVolume()

		if err != nil {
			slog.Error("getting master volume failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), gain)
	}

	return nil
}
