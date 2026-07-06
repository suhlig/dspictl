package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newDACMuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dac-mute",
		Short: "DAC hardware mute control",
		RunE:  runDACMuteStatus,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "on",
		Short: "Enable DAC hardware mute",
		RunE:  runDACMuteOn,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Disable DAC hardware mute",
		RunE:  runDACMuteOff,
	})

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure DAC hardware mute parameters",
		Long: `Configure the DAC hardware mute GPIO parameters.

All parameters are set at once:

  --enabled            Enable DAC hardware mute (on/off, true/false, or 1/0)
  --active-low         Polarity (on/off, true/false, or 1/0)
  --pin                GPIO pin number (255 = none/disabled)
  --hold-ms            Hold time in milliseconds before muting
  --release-ms         Release time in milliseconds after unmuting`,
		Args: cobra.NoArgs,
		RunE: runDACMuteConfig,
	}
	configCmd.Flags().Bool("enabled", false, "Enable DAC hardware mute")
	configCmd.Flags().Bool("active-low", false, "Polarity (active low)")
	configCmd.Flags().Int("pin", 255, "GPIO pin number (255 = none/disabled)")
	configCmd.Flags().Int("hold-ms", 0, "Hold time in milliseconds")
	configCmd.Flags().Int("release-ms", 0, "Release time in milliseconds")
	_ = configCmd.MarkFlagRequired("enabled")
	_ = configCmd.MarkFlagRequired("active-low")
	_ = configCmd.MarkFlagRequired("pin")
	_ = configCmd.MarkFlagRequired("hold-ms")
	_ = configCmd.MarkFlagRequired("release-ms")
	cmd.AddCommand(configCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Test DAC hardware mute (brief mute cycle)",
		RunE:  runDACMuteTest,
	})

	return cmd
}

func runDACMuteStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		cfg, err := d.GetDACHwMute()
		if err != nil {
			slog.Error("getting DAC mute config failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if cfg.Enabled {
			state = "on"
		}
		activeLow := "off"
		if cfg.ActiveLow {
			activeLow = "on"
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  Enabled: %s\n", state)
		fmt.Printf("  Active-low: %s\n", activeLow)
		fmt.Printf("  Pin: %d\n", cfg.Pin)
		fmt.Printf("  Hold: %d ms\n", cfg.HoldMs)
		fmt.Printf("  Release: %d ms\n", cfg.ReleaseMs)
	}

	return nil
}

func runDACMuteOn(cmd *cobra.Command, args []string) error {
	return setDACMuteEnabled(true)
}

func runDACMuteOff(cmd *cobra.Command, args []string) error {
	return setDACMuteEnabled(false)
}

func setDACMuteEnabled(enabled bool) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		// Read current config, toggle enabled, write back.
		cfg, err := d.GetDACHwMute()
		if err != nil {
			slog.Error("getting DAC mute config failed", "serial", d.Serial(), "error", err)
			continue
		}

		cfg.Enabled = enabled

		if err := d.SetDACHwMute(cfg); err != nil {
			slog.Error("setting DAC mute failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if enabled {
			state = "on"
		}
		fmt.Printf("%s: DAC hardware mute=%s\n", d.Serial(), state)
	}

	return nil
}

func runDACMuteConfig(cmd *cobra.Command, args []string) error {
	enabled, err := cmd.Flags().GetBool("enabled")
	if err != nil {
		return fmt.Errorf("getting enabled flag: %w", err)
	}

	activeLow, err := cmd.Flags().GetBool("active-low")
	if err != nil {
		return fmt.Errorf("getting active-low flag: %w", err)
	}

	pin, err := cmd.Flags().GetInt("pin")
	if err != nil {
		return fmt.Errorf("getting pin flag: %w", err)
	}

	holdMs, err := cmd.Flags().GetInt("hold-ms")
	if err != nil {
		return fmt.Errorf("getting hold-ms flag: %w", err)
	}

	releaseMs, err := cmd.Flags().GetInt("release-ms")
	if err != nil {
		return fmt.Errorf("getting release-ms flag: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	cfg := &dspi.DACHwMuteConfig{
		Enabled:   enabled,
		ActiveLow: activeLow,
		Pin:       pin,
		HoldMs:    uint16(holdMs),
		ReleaseMs: uint16(releaseMs),
	}

	for _, d := range devices {
		if err := d.SetDACHwMute(cfg); err != nil {
			slog.Error("setting DAC mute config failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: DAC hardware mute configured\n", d.Serial())
	}

	return nil
}

func runDACMuteTest(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.TestDACHwMute(); err != nil {
			slog.Error("testing DAC mute failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: DAC hardware mute test triggered\n", d.Serial())
	}

	return nil
}

// parseBoolArg parses on/off, true/false, 1/0.
func parseBoolArg(s string) (bool, error) {
	switch s {
	case "on", "true", "1":
		return true, nil
	case "off", "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("expected on/off, true/false, or 1/0, got %q", s)
	}
}
