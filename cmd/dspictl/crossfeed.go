package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

func newCrossfeedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crossfeed",
		Short: "Crossfeed (headphone spatialization) control",
		RunE:  runCrossfeedStatus,
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "enable [on|off]",
		Short:             "Get or set crossfeed enable state",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runCrossfeedEnable,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "preset [<n>]",
		Short: "Get or set crossfeed preset (0-4)",
		Long: `Get or set the crossfeed preset.

Presets:
  0 = Default
  1 = Diffuse field
  2 = Careful
  3 = Medium
  4 = Aggressive`,
		Args: cobra.MaximumNArgs(1),
		RunE: runCrossfeedPreset,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "freq [<hz>]",
		Short: "Get or set crossfeed crossover frequency in Hz",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCrossfeedFreq,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "feed [<db>]",
		Short: "Get or set crossfeed feed level in dB",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCrossfeedFeed,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "itd [on|off]",
		Short:             "Get or set crossfeed interaural time delay",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runCrossfeedITD,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "outputs [mask]",
		Short: "Get or set output-pair mask (hex, e.g. 0x03)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCrossfeedOutputs,
	})

	return cmd
}

func runCrossfeedStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		enabled, err := d.GetCrossfeed()
		if err != nil {
			slog.Error("getting crossfeed status failed", "serial", d.Serial(), "error", err)
			continue
		}

		preset, _ := d.GetCrossfeedPreset()
		freq, _ := d.GetCrossfeedFreq()
		feed, _ := d.GetCrossfeedFeed()
		itd, _ := d.GetCrossfeedITD()

		state := "off"
		if enabled {
			state = "on"
		}

		itdState := "off"
		if itd {
			itdState = "on"
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  Enable: %s\n", state)
		fmt.Printf("  Preset: %d\n", preset)
		fmt.Printf("  Freq: %.0f Hz\n", freq)
		fmt.Printf("  Feed: %.1f dB\n", feed)
		fmt.Printf("  ITD: %s\n", itdState)

		outputMask, _ := d.GetCrossfeedOutputPairMask()
		outputState := fmt.Sprintf("0x%02X", outputMask)
		fmt.Printf("  Output pairs: %s\n", outputState)
	}

	return nil
}

func runCrossfeedEnable(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			enabled, err := d.GetCrossfeed()
			if err != nil {
				slog.Error("getting crossfeed enable failed", "serial", d.Serial(), "error", err)
				continue
			}
			state := "off"
			if enabled {
				state = "on"
			}
			fmt.Printf("%s: crossfeed=%s\n", d.Serial(), state)
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
		err := d.SetCrossfeed(enable)
		if err != nil {
			slog.Error("setting crossfeed enable failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: crossfeed=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runCrossfeedPreset(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			preset, err := d.GetCrossfeedPreset()
			if err != nil {
				slog.Error("getting crossfeed preset failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: crossfeed preset=%d\n", d.Serial(), preset)
		}
		return nil
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid preset: %w", err)
	}

	for _, d := range devices {
		err := d.SetCrossfeedPreset(n)
		if err != nil {
			slog.Error("setting crossfeed preset failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: crossfeed preset=%d\n", d.Serial(), n)
	}

	return nil
}

func runCrossfeedFreq(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			freq, err := d.GetCrossfeedFreq()
			if err != nil {
				slog.Error("getting crossfeed freq failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: crossfeed freq=%.0f Hz\n", d.Serial(), freq)
		}
		return nil
	}

	freq, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid frequency: %w", err)
	}

	for _, d := range devices {
		err := d.SetCrossfeedFreq(freq)
		if err != nil {
			slog.Error("setting crossfeed freq failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: crossfeed freq=%.0f Hz\n", d.Serial(), freq)
	}

	return nil
}

func runCrossfeedFeed(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			feed, err := d.GetCrossfeedFeed()
			if err != nil {
				slog.Error("getting crossfeed feed failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: crossfeed feed=%.1f dB\n", d.Serial(), feed)
		}
		return nil
	}

	feed, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid feed value: %w", err)
	}

	for _, d := range devices {
		err := d.SetCrossfeedFeed(feed)
		if err != nil {
			slog.Error("setting crossfeed feed failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: crossfeed feed=%.1f dB\n", d.Serial(), feed)
	}

	return nil
}

func runCrossfeedITD(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			itd, err := d.GetCrossfeedITD()
			if err != nil {
				slog.Error("getting crossfeed ITD failed", "serial", d.Serial(), "error", err)
				continue
			}
			state := "off"
			if itd {
				state = "on"
			}
			fmt.Printf("%s: crossfeed ITD=%s\n", d.Serial(), state)
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
		err := d.SetCrossfeedITD(enable)
		if err != nil {
			slog.Error("setting crossfeed ITD failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: crossfeed ITD=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runCrossfeedOutputs(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mask, err := d.GetCrossfeedOutputPairMask()
			if err != nil {
				slog.Error("getting crossfeed output mask failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: crossfeed output-pair mask = 0x%02X\n", d.Serial(), mask)
		}
		return nil
	}

	mask, err := strconv.ParseUint(args[0], 0, 16)
	if err != nil {
		return fmt.Errorf("invalid mask value: %w", err)
	}

	for _, d := range devices {
		if err := d.SetCrossfeedOutputPairMask(uint8(mask)); err != nil {
			slog.Error("setting crossfeed output mask failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: crossfeed output-pair mask = 0x%02X\n", d.Serial(), uint8(mask))
	}

	return nil
}
