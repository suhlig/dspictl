package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
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
