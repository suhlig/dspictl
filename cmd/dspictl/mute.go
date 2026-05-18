package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newMuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mute",
		Short: "Mute master volume",
		RunE:  runMute,
	}
}

func runMute(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMasterVolume(dspi.NewGain(-128))

		if err != nil {
			return fmt.Errorf("%s: %w", d.Serial(), err)
		}

		fmt.Printf("%s: muted\n", d.Serial())
	}

	return nil
}
