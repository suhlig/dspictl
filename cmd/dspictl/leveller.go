package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

func newLevellerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leveller",
		Short: "Dynamic range compression (leveller) control",
		RunE:  runLevellerStatus,
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "enable [on|off]",
		Short:             "Get or set leveller enable state",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runLevellerEnable,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "amount [<value>]",
		Short: "Get or set leveller compression amount",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLevellerAmount,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "speed [<n>]",
		Short: "Get or set leveller attack/release speed",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLevellerSpeed,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "maxgain [<db>]",
		Short: "Get or set leveller maximum gain reduction in dB",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLevellerMaxGain,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "lookahead [on|off]",
		Short:             "Get or set leveller lookahead",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runLevellerLookahead,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "gate [<db>]",
		Short: "Get or set leveller noise gate threshold in dB",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runLevellerGate,
	})

	return cmd
}

func runLevellerStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		enabled, err := d.GetLeveller()
		if err != nil {
			slog.Error("getting leveller status failed", "serial", d.Serial(), "error", err)
			continue
		}

		amount, _ := d.GetLevellerAmount()
		speed, _ := d.GetLevellerSpeed()
		maxGain, _ := d.GetLevellerMaxGain()
		lookahead, _ := d.GetLevellerLookahead()
		gate, _ := d.GetLevellerGate()

		state := "off"
		if enabled {
			state = "on"
		}

		laState := "off"
		if lookahead {
			laState = "on"
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  Enable: %s\n", state)
		fmt.Printf("  Amount: %.1f\n", amount)
		fmt.Printf("  Speed: %d\n", speed)
		fmt.Printf("  Max Gain: %.1f dB\n", maxGain)
		fmt.Printf("  Lookahead: %s\n", laState)
		fmt.Printf("  Gate: %.1f dB\n", gate)
	}

	return nil
}

func runLevellerEnable(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			enabled, err := d.GetLeveller()
			if err != nil {
				slog.Error("getting leveller enable failed", "serial", d.Serial(), "error", err)
				continue
			}
			state := "off"
			if enabled {
				state = "on"
			}
			fmt.Printf("%s: leveller=%s\n", d.Serial(), state)
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
		err := d.SetLeveller(enable)
		if err != nil {
			slog.Error("setting leveller enable failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: leveller=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runLevellerAmount(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			amount, err := d.GetLevellerAmount()
			if err != nil {
				slog.Error("getting leveller amount failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: leveller amount=%.1f\n", d.Serial(), amount)
		}
		return nil
	}

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	for _, d := range devices {
		err := d.SetLevellerAmount(amount)
		if err != nil {
			slog.Error("setting leveller amount failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: leveller amount=%.1f\n", d.Serial(), amount)
	}

	return nil
}

func runLevellerSpeed(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			speed, err := d.GetLevellerSpeed()
			if err != nil {
				slog.Error("getting leveller speed failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: leveller speed=%d\n", d.Serial(), speed)
		}
		return nil
	}

	speed, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid speed: %w", err)
	}

	for _, d := range devices {
		err := d.SetLevellerSpeed(speed)
		if err != nil {
			slog.Error("setting leveller speed failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: leveller speed=%d\n", d.Serial(), speed)
	}

	return nil
}

func runLevellerMaxGain(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			maxGain, err := d.GetLevellerMaxGain()
			if err != nil {
				slog.Error("getting leveller max gain failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: leveller maxgain=%.1f dB\n", d.Serial(), maxGain)
		}
		return nil
	}

	maxGain, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid max gain: %w", err)
	}

	for _, d := range devices {
		err := d.SetLevellerMaxGain(maxGain)
		if err != nil {
			slog.Error("setting leveller max gain failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: leveller maxgain=%.1f dB\n", d.Serial(), maxGain)
	}

	return nil
}

func runLevellerLookahead(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			lookahead, err := d.GetLevellerLookahead()
			if err != nil {
				slog.Error("getting leveller lookahead failed", "serial", d.Serial(), "error", err)
				continue
			}
			state := "off"
			if lookahead {
				state = "on"
			}
			fmt.Printf("%s: leveller lookahead=%s\n", d.Serial(), state)
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
		err := d.SetLevellerLookahead(enable)
		if err != nil {
			slog.Error("setting leveller lookahead failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: leveller lookahead=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runLevellerGate(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			gate, err := d.GetLevellerGate()
			if err != nil {
				slog.Error("getting leveller gate failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: leveller gate=%.1f dB\n", d.Serial(), gate)
		}
		return nil
	}

	gate, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid gate value: %w", err)
	}

	for _, d := range devices {
		err := d.SetLevellerGate(gate)
		if err != nil {
			slog.Error("setting leveller gate failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: leveller gate=%.1f dB\n", d.Serial(), gate)
	}

	return nil
}
