package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newUnmuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unmute",
		Short: "Unmute master volume to firmware default (-20 dB)",
		RunE:  runUnmute,
	}
}

func runUnmute(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMasterVolume(dspi.NewGain(-20))

		if err != nil {
			return fmt.Errorf("%s: %w", d.Serial(), err)
		}

		fmt.Printf("%s: unmuted (-20 dB)\n", d.Serial())
	}

	return nil
}
