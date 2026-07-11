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

	cmd.AddCommand(&cobra.Command{
		Use:   "detector-mask [on|off] [<channels...>]",
		Short: "Get or set detector input-channel mask",
		Long: `Get or set which input channels feed the shared RMS detector.

With no arguments, shows the current active inputs.
With "on" or "off" followed by channel numbers, toggles specific inputs.
With a preset name, sets the mask to a predefined value.

Presets:
  all      – all inputs (Night mode)
  center   – center channel only (Dialog boost)
  front-lr – front L/R only
  none     – disable all`,
		Args: cobra.ArbitraryArgs,
		RunE: runLevellerDetectorMask,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "apply-mask [on|off] [<channels...>]",
		Short: "Get or set apply input-channel mask",
		Long: `Get or set which input channels receive the computed gain.

With no arguments, shows the current active inputs.
With "on" or "off" followed by channel numbers, toggles specific inputs.
With a preset name, sets the mask to a predefined value.

Presets:
  all      – all inputs (Night mode)
  center   – center channel only (Dialog boost)
  front-lr – front L/R only
  none     – disable all`,
		Args: cobra.ArbitraryArgs,
		RunE: runLevellerApplyMask,
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

		detMask, appMask, _ := d.GetLevellerMasks()
		fmt.Printf("  Detector inputs: %s\n", formatMaskU8(detMask, 8))
		fmt.Printf("  Apply inputs: %s\n", formatMaskU8(appMask, 8))
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

func runLevellerDetectorMask(cmd *cobra.Command, args []string) error {
	return runLevellerMaskCmd(args, true)
}

func runLevellerApplyMask(cmd *cobra.Command, args []string) error {
	return runLevellerMaskCmd(args, false)
}

// levellerPresets maps preset names to detector/apply mask pairs.
var levellerPresets = map[string][2]uint8{
	"all":      {0xFF, 0xFF},
	"night":    {0xFF, 0xFF},
	"center":   {0x04, 0x04},
	"front-lr": {0x03, 0x03},
	"none":     {0x00, 0x00},
}

func runLevellerMaskCmd(args []string, isDetector bool) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			detMask, appMask, err := d.GetLevellerMasks()
			if err != nil {
				slog.Error("getting leveller masks failed", "serial", d.Serial(), "error", err)
				continue
			}
			if isDetector {
				fmt.Printf("%s: leveller detector inputs: %s\n", d.Serial(), formatMaskU8(detMask, 8))
			} else {
				fmt.Printf("%s: leveller apply inputs: %s\n", d.Serial(), formatMaskU8(appMask, 8))
			}
		}
		return nil
	}

	// Named presets
	if preset, ok := levellerPresets[args[0]]; ok {
		for _, d := range devices {
			if err := d.SetLevellerMasks(preset[0], preset[1]); err != nil {
				slog.Error("setting leveller masks failed", "serial", d.Serial(), "error", err)
				continue
			}
			if isDetector {
				fmt.Printf("%s: leveller detector inputs: %s\n", d.Serial(), formatMaskU8(preset[0], 8))
			} else {
				fmt.Printf("%s: leveller apply inputs: %s\n", d.Serial(), formatMaskU8(preset[1], 8))
			}
		}
		return nil
	}

	if args[0] != "on" && args[0] != "off" {
		return fmt.Errorf("expected \"on\", \"off\", \"all\", \"center\", \"front-lr\", \"night\", or \"none\", got %q", args[0])
	}
	enable := args[0] == "on"
	channelArgs := args[1:]
	if len(channelArgs) == 0 {
		return fmt.Errorf("%s requires at least one channel number", args[0])
	}

	for _, d := range devices {
		detMask, appMask, err := d.GetLevellerMasks()
		if err != nil {
			slog.Error("getting leveller masks failed", "serial", d.Serial(), "error", err)
			continue
		}

		var newMask uint8
		if enable {
			newMask, err = maskSetBitsU8(detMask, channelArgs, 8)
		} else {
			newMask, err = maskClearBitsU8(detMask, channelArgs, 8)
		}
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		var setErr error
		if isDetector {
			setErr = d.SetLevellerMasks(newMask, appMask)
		} else {
			setErr = d.SetLevellerMasks(detMask, newMask)
		}
		if setErr != nil {
			slog.Error("setting leveller masks failed", "serial", d.Serial(), "error", setErr)
			continue
		}

		if isDetector {
			fmt.Printf("%s: leveller detector inputs: %s\n", d.Serial(), formatMaskU8(newMask, 8))
		} else {
			fmt.Printf("%s: leveller apply inputs: %s\n", d.Serial(), formatMaskU8(newMask, 8))
		}
	}

	return nil
}
