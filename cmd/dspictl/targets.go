package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newTargetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List discovered DSPi devices",
		RunE:  runTargets,
	}
}

func runTargets(cmd *cobra.Command, args []string) error {
	devices, err := dspi.List()

	if err != nil {
		return fmt.Errorf("listing DSPi devices: %w", err)
	}

	if len(devices) == 0 {
		fmt.Println("No DSPi devices found")

		return nil
	}

	for _, d := range devices {
		fmt.Println(d.Serial)
	}

	return nil
}
