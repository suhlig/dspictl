package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newPresetEQCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eq",
		Short: "Modify filters in a preset slot",
	}

	cmd.AddCommand(newPresetEQMasterCmd())
	cmd.AddCommand(newPresetEQOutputCmd())
	cmd.AddCommand(newPresetEQCrossoverCmd())

	return cmd
}

func newPresetEQMasterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master",
		Short: "Master EQ (channels 0-1) in a preset slot",
	}

	setCmd := &cobra.Command{
		Use:               "set <slot> <channel> <band>",
		Short:             "Configure a master EQ band in a preset slot",
		Args:              cobra.ExactArgs(3),
		RunE:              runPresetEQMasterSet,
		ValidArgsFunction: completePresetEQChannelBand(func(ch dspi.ChannelInfo) bool { return ch.Index <= 1 }, true),
	}
	setCmd.Flags().String("type", "", "Filter type: flat, peak, lowshelf, highshelf, lowpass, highpass")
	setCmd.Flags().Float64("freq", 0, "Frequency in Hz")
	setCmd.Flags().Float64("q", 0, "Q factor")
	setCmd.Flags().Float64("gain", 0, "Gain in dB")
	_ = setCmd.MarkFlagRequired("type")
	cmd.AddCommand(setCmd)

	cmd.AddCommand(&cobra.Command{
		Use:               "clear <slot> <channel>",
		Short:             "Reset all master bands to flat in a preset slot",
		Args:              cobra.ExactArgs(2),
		RunE:              runPresetEQMasterClear,
		ValidArgsFunction: completePresetEQChannelBand(func(ch dspi.ChannelInfo) bool { return ch.Index <= 1 }, false),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "bypass <slot> [true|false]",
		Short:             "Get or set master EQ bypass in a preset slot",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runPresetEQMasterBypass,
		ValidArgsFunction: completeChoices(nil, []string{"true", "false"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "band-bypass <slot> <channel> <band> [true|false]",
		Short:             "Get or set bypass for a single master band in a preset slot",
		Args:              cobra.RangeArgs(3, 4),
		RunE:              runPresetEQMasterBandBypass,
		ValidArgsFunction: completePresetEQChannelBand(func(ch dspi.ChannelInfo) bool { return ch.Index <= 1 }, true),
	})

	return cmd
}

func newPresetEQOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Per-output EQ in a preset slot",
	}

	setCmd := &cobra.Command{
		Use:               "set <slot> <channel> <band>",
		Short:             "Configure an output EQ band in a preset slot",
		Args:              cobra.ExactArgs(3),
		RunE:              runPresetEQOutputSet,
		ValidArgsFunction: completePresetEQOutputChannelBand(true),
	}
	setCmd.Flags().String("type", "", "Filter type: flat, peak, lowshelf, highshelf, lowpass, highpass")
	setCmd.Flags().Float64("freq", 0, "Frequency in Hz")
	setCmd.Flags().Float64("q", 0, "Q factor")
	setCmd.Flags().Float64("gain", 0, "Gain in dB")
	_ = setCmd.MarkFlagRequired("type")
	cmd.AddCommand(setCmd)

	cmd.AddCommand(&cobra.Command{
		Use:               "clear <slot> <channel>",
		Short:             "Reset all output bands to flat in a preset slot",
		Args:              cobra.ExactArgs(2),
		RunE:              runPresetEQOutputClear,
		ValidArgsFunction: completePresetEQOutputChannelBand(false),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "band-bypass <slot> <channel> <band> [true|false]",
		Short:             "Get or set bypass for a single output band in a preset slot",
		Args:              cobra.RangeArgs(3, 4),
		RunE:              runPresetEQOutputBandBypass,
		ValidArgsFunction: completePresetEQOutputChannelBand(true),
	})

	return cmd
}

func newPresetEQCrossoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crossover",
		Short: "Crossover filters in a preset slot (output channels only, bands 20-23)",
	}

	setCmd := &cobra.Command{
		Use:               "set <slot> <channel> <band>",
		Short:             "Configure a crossover band in a preset slot",
		Args:              cobra.ExactArgs(3),
		RunE:              runPresetEQCrossoverSet,
		ValidArgsFunction: completePresetEQOutputChannelBand(true),
	}
	setCmd.Flags().String("type", "", "Crossover filter type (e.g. lr4-lp, bw2-hp, bes6-lp)")
	setCmd.Flags().Float64("freq", 0, "Frequency in Hz")
	setCmd.Flags().Bool("bypass", false, "Bypass this band")
	_ = setCmd.MarkFlagRequired("type")
	cmd.AddCommand(setCmd)

	cmd.AddCommand(&cobra.Command{
		Use:               "clear <slot> <channel>",
		Short:             "Reset all crossover bands to flat in a preset slot",
		Args:              cobra.ExactArgs(2),
		RunE:              runPresetEQCrossoverClear,
		ValidArgsFunction: completePresetEQOutputChannelBand(false),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "bypass <slot> <channel> <band> [true|false]",
		Short:             "Get or set bypass for a crossover band in a preset slot",
		Args:              cobra.RangeArgs(3, 4),
		RunE:              runPresetEQCrossoverBypass,
		ValidArgsFunction: completePresetEQOutputChannelBand(true),
	})

	return cmd
}

// -- Completion helpers for preset eq commands --

// completePresetEQChannelBand returns a ValidArgsFunction for preset eq commands
// that take <slot> <channel> [<band>]. The channelFilter selects which channels
// are valid completions. hasBand controls whether a band argument is expected.
func completePresetEQChannelBand(channelFilter func(dspi.ChannelInfo) bool, hasBand bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return presetEQSlotChoices, cobra.ShellCompDirectiveNoFileComp
		case 1:
			return completeChannelIndices(channelFilter)
		case 2:
			if hasBand {
				return completeEQBands()
			}
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// completePresetEQOutputChannelBand returns a ValidArgsFunction for preset eq
// commands operating on output channels (0-based output indices).
func completePresetEQOutputChannelBand(hasBand bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return presetEQSlotChoices, cobra.ShellCompDirectiveNoFileComp
		case 1:
			return completeOutputChannels(cmd, nil, toComplete)
		case 2:
			if hasBand {
				return completeEQBands()
			}
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

var presetEQSlotChoices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// -- Run functions for preset eq master --

func runPresetEQMasterSet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := parseEQBandArgs(args[1:], cmd.Flags(), ch)
	if err != nil {
		return err
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetEQBand(band)
		})
	})
}

func runPresetEQMasterClear(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			maxBand, err := d.MaxBands()
			if err != nil {
				return fmt.Errorf("getting max bands: %w", err)
			}

			for band := range maxBand {
				if err := d.SetEQBand(&dspi.EQBand{
					Channel: ch,
					Band:    band,
					Type:    dspi.FilterTypeFlat,
				}); err != nil {
					return fmt.Errorf("clearing band %d: %w", band, err)
				}
			}

			return nil
		})
	})
}

func runPresetEQMasterBypass(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	if len(args) == 1 {
		// Get current bypass state
		return withEachDevice(func(d *dspi.Device) error {
			var bypass bool
			err := d.WithPresetSlot(slot, func() error {
				var err error
				bypass, err = d.GetMasterEQBypass()
				return err
			})
			if err != nil {
				return err
			}

			state := "off"
			if bypass {
				state = "on"
			}
			fmt.Printf("%s: slot %d master-bypass=%s\n", d.Serial(), slot, state)
			return nil
		})
	}

	var bypass bool
	switch args[1] {
	case "true":
		bypass = true
	case "false":
		bypass = false
	default:
		return fmt.Errorf("invalid value: %s (expected true or false)", args[1])
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetMasterEQBypass(bypass)
		})
	})
}

func runPresetEQMasterBandBypass(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid band: %w", err)
	}

	if len(args) == 3 {
		// Get current bypass state
		return withEachDevice(func(d *dspi.Device) error {
			var bypass bool
			err := d.WithPresetSlot(slot, func() error {
				var err error
				bypass, err = d.GetBandBypass(ch, band)
				return err
			})
			if err != nil {
				return err
			}

			state := "off"
			if bypass {
				state = "on"
			}
			fmt.Printf("%s: slot %d ch %d band %d: bypass=%s\n", d.Serial(), slot, ch, band, state)
			return nil
		})
	}

	var bypass bool
	switch args[3] {
	case "true":
		bypass = true
	case "false":
		bypass = false
	default:
		return fmt.Errorf("invalid value: %s (expected true or false)", args[3])
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetBandBypass(ch, band, bypass)
		})
	})
}

// -- Run functions for preset eq output --

func runPresetEQOutputSet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	// Output channels are firmware channels (ch + 2)
	firmwareCh := ch + 2
	band, err := parseEQBandArgs(args[1:], cmd.Flags(), firmwareCh)
	if err != nil {
		return err
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetEQBand(band)
		})
	})
}

func runPresetEQOutputClear(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	firmwareCh := ch + 2

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			maxBand, err := d.MaxBands()
			if err != nil {
				return fmt.Errorf("getting max bands: %w", err)
			}

			for band := range maxBand {
				if err := d.SetEQBand(&dspi.EQBand{
					Channel: firmwareCh,
					Band:    band,
					Type:    dspi.FilterTypeFlat,
				}); err != nil {
					return fmt.Errorf("clearing band %d: %w", band, err)
				}
			}

			return nil
		})
	})
}

func runPresetEQOutputBandBypass(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid band: %w", err)
	}

	firmwareCh := ch + 2

	if len(args) == 3 {
		return withEachDevice(func(d *dspi.Device) error {
			var bypass bool
			err := d.WithPresetSlot(slot, func() error {
				var err error
				bypass, err = d.GetBandBypass(firmwareCh, band)
				return err
			})
			if err != nil {
				return err
			}

			state := "off"
			if bypass {
				state = "on"
			}
			fmt.Printf("%s: slot %d ch %d band %d: bypass=%s\n", d.Serial(), slot, firmwareCh, band, state)
			return nil
		})
	}

	var bypass bool
	switch args[3] {
	case "true":
		bypass = true
	case "false":
		bypass = false
	default:
		return fmt.Errorf("invalid value: %s (expected true or false)", args[3])
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetBandBypass(firmwareCh, band, bypass)
		})
	})
}

// -- Run functions for preset eq crossover --

func runPresetEQCrossoverSet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid band: %w", err)
	}

	filterType, err := cmd.Flags().GetString("type")
	if err != nil {
		return fmt.Errorf("getting type flag: %w", err)
	}

	t, err := dspi.ParseCrossoverFilterType(filterType)
	if err != nil {
		return err
	}

	freq, _ := cmd.Flags().GetFloat64("freq")
	bypass, _ := cmd.Flags().GetBool("bypass")

	cb := &dspi.CrossoverBand{
		Channel: ch + 2,
		Band:    band,
		Type:    t,
		Freq:    freq,
		Bypass:  bypass,
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetCrossoverBand(cb)
		})
	})
}

func runPresetEQCrossoverClear(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	firmwareCh := ch + 2

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			for band := 20; band < 20+d.MaxCrossoverBands(); band++ {
				if err := d.SetCrossoverBand(&dspi.CrossoverBand{
					Channel: firmwareCh,
					Band:    band,
					Type:    dspi.CrossoverTypeLR2LP,
					Freq:    1000,
				}); err != nil {
					return fmt.Errorf("clearing crossover band %d: %w", band, err)
				}
			}

			return nil
		})
	})
}

func runPresetEQCrossoverBypass(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	ch, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid band: %w", err)
	}

	firmwareCh := ch + 2

	if len(args) == 3 {
		return withEachDevice(func(d *dspi.Device) error {
			var bypass bool
			err := d.WithPresetSlot(slot, func() error {
				var err error
				bypass, err = d.GetBandBypass(firmwareCh, band)
				return err
			})
			if err != nil {
				return err
			}

			state := "off"
			if bypass {
				state = "on"
			}
			fmt.Printf("%s: slot %d ch %d band %d: bypass=%s\n", d.Serial(), slot, firmwareCh, band, state)
			return nil
		})
	}

	var bypass bool
	switch args[3] {
	case "true":
		bypass = true
	case "false":
		bypass = false
	default:
		return fmt.Errorf("invalid value: %s (expected true or false)", args[3])
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlot(slot, func() error {
			return d.SetBandBypass(firmwareCh, band, bypass)
		})
	})
}

// withEachDevice opens all target devices and calls fn for each one.
// It logs errors per-device and continues to the next device.
func withEachDevice(fn func(*dspi.Device) error) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	var lastErr error

	for _, d := range devices {
		if err := fn(d); err != nil {
			slog.Error("operation failed", "serial", d.Serial(), "error", err)
			lastErr = err
		}
	}

	return lastErr
}
