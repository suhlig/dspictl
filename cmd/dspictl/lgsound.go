package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newLGSoundSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lg-sound-sync",
		Short: "LG Sound Sync control",
		RunE:  runLGSoundSyncStatus,
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "enable [on|off]",
		Short:             "Get or set LG Sound Sync enable state",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runLGSoundSyncEnable,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show LG Sound Sync status (TV presence, volume, mute)",
		RunE:  runLGSoundSyncStatusDetail,
	})

	return cmd
}

func runLGSoundSyncStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		enabled, err := d.GetLGSoundSync()
		if err != nil {
			slog.Error("getting LG Sound Sync status failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if enabled {
			state = "on"
		}

		fmt.Printf("%s: LG Sound Sync=%s\n", d.Serial(), state)
	}

	return nil
}

func runLGSoundSyncEnable(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			enabled, err := d.GetLGSoundSync()
			if err != nil {
				slog.Error("getting LG Sound Sync failed", "serial", d.Serial(), "error", err)
				continue
			}
			state := "off"
			if enabled {
				state = "on"
			}
			fmt.Printf("%s: LG Sound Sync=%s\n", d.Serial(), state)
		}
		return nil
	}

	var enable bool
	switch args[0] {
	case "on":
		enable = true
	case "off":
		enable = false
	default:
		return fmt.Errorf("invalid value: %s (expected on or off)", args[0])
	}

	for _, d := range devices {
		err := d.SetLGSoundSync(enable)
		if err != nil {
			slog.Error("setting LG Sound Sync failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: LG Sound Sync=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runLGSoundSyncStatusDetail(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		status, err := d.GetLGSoundSyncStatus()
		if err != nil {
			slog.Error("getting LG Sound Sync status failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if status.Enabled {
			state = "on"
		}
		present := "no"
		if status.Present {
			present = "yes"
		}
		muted := "no"
		if status.Muted {
			muted = "yes"
		}

		volStr := "unknown"
		if status.Volume >= 0 {
			volStr = fmt.Sprintf("%d", status.Volume)
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  Enabled: %s\n", state)
		fmt.Printf("  TV Present: %s\n", present)
		fmt.Printf("  Volume: %s\n", volStr)
		fmt.Printf("  Muted: %s\n", muted)
	}

	return nil
}
