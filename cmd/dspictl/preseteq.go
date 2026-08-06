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
		Short: "Modify or list filters in a preset slot",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "list <slot>",
		Short:             "Show all filter bands in a preset slot",
		Args:              cobra.ExactArgs(1),
		RunE:              runPresetEQList,
		ValidArgsFunction: completeChoices(presetEQSlotChoices),
	})

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
	setCmd.Flags().String("type", "", "Filter type: flat, peak, lowshelf, highshelf, lowpass, highpass, notch, allpass, allpass1, lowshelf1, highshelf1, lowpass1, highpass1, linkwitz")
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
		Use:               "bypass <slot> [on|off]",
		Short:             "Get or set master EQ bypass in a preset slot",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runPresetEQMasterBypass,
		ValidArgsFunction: completeChoices(nil, []string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "band-bypass <slot> <channel> <band> [on|off]",
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
	setCmd.Flags().String("type", "", "Filter type: flat, peak, lowshelf, highshelf, lowpass, highpass, notch, allpass, allpass1, lowshelf1, highshelf1, lowpass1, highpass1, linkwitz")
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
		Use:               "band-bypass <slot> <channel> <band> [on|off]",
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
		Use:               "bypass <slot> <channel> <band> [on|off]",
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

	bypass, err := parseBoolArg(args[1])
	if err != nil {
		return fmt.Errorf("invalid bypass value: %w", err)
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

	bypass, err := parseBoolArg(args[3])
	if err != nil {
		return fmt.Errorf("invalid bypass value: %w", err)
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

	bypass, err := parseBoolArg(args[3])
	if err != nil {
		return fmt.Errorf("invalid bypass value: %w", err)
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

	cb := &dspi.CrossoverBand{
		Channel: ch + 2,
		Band:    band,
		Type:    t,
		Freq:    freq,
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

	bypass, err := parseBoolArg(args[3])
	if err != nil {
		return fmt.Errorf("invalid bypass value: %w", err)
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

// -- Run function for preset eq list --

func runPresetEQList(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	return withEachDevice(func(d *dspi.Device) error {
		return d.WithPresetSlotReadOnly(slot, func() error {
			fmt.Printf("%s: slot %d\n", d.Serial(), slot)
			return listDeviceFilters(d)
		})
	})
}

func listDeviceFilters(d *dspi.Device) error {
	maxBand, err := d.MaxBands()
	if err != nil {
		return fmt.Errorf("getting max bands: %w", err)
	}

	maxChannel := d.MaxEQChannel()

	channels, err := d.Channels()
	if err != nil {
		return fmt.Errorf("getting channels: %w", err)
	}

	channelName := func(idx int) string {
		if idx < len(channels) {
			return channels[idx].Name
		}
		return fmt.Sprintf("ch %d", idx)
	}

	// Master EQ (channels 0-1)
	fmt.Println("  Master EQ:")

	for ch := 0; ch <= 1; ch++ {
		hasActive := false

		for band := range maxBand {
			b, err := d.GetEQBand(ch, band)
			if err != nil {
				slog.Debug("skipping master EQ band", "serial", d.Serial(), "channel", ch, "band", band, "error", err)
				continue
			}

			if b.Type != dspi.FilterTypeFlat {
				if !hasActive {
					fmt.Printf("    ch %d (%s):\n", ch, channelName(ch))
					hasActive = true
				}

				printEQBand(b)
			}
		}

		if !hasActive {
			fmt.Printf("    ch %d (%s): (no active bands)\n", ch, channelName(ch))
		}
	}

	// Master EQ bypass (best-effort)
	bypass, err := d.GetMasterEQBypass()
	if err == nil {
		state := "off"
		if bypass {
			state = "on"
		}
		fmt.Printf("    bypass=%s\n", state)
	}

	// Output EQ and crossover for each output channel
	for ch := 2; ch <= maxChannel; ch++ {
		name := channelName(ch)
		outputIdx := ch - 2

		// Output EQ (best-effort)
		fmt.Printf("  Output ch %d (%s) EQ:\n", outputIdx, name)
		hasActive := false

		for band := range maxBand {
			b, err := d.GetEQBand(ch, band)
			if err != nil {
				slog.Debug("skipping output EQ band", "serial", d.Serial(), "channel", ch, "band", band, "error", err)
				continue
			}

			if b.Type != dspi.FilterTypeFlat {
				if !hasActive {
					hasActive = true
				}

				printEQBand(b)
			}
		}

		if !hasActive {
			fmt.Println("    (no active bands)")
		}

		// Crossover (best-effort: skip channels/firmware that don't support it)
		fmt.Printf("  Output ch %d (%s) crossover:\n", outputIdx, name)
		hasXover := false

		for band := 20; band < 20+d.MaxCrossoverBands(); band++ {
			b, err := d.GetCrossoverBand(ch, band)
			if err != nil {
				slog.Debug("skipping crossover band", "serial", d.Serial(), "channel", ch, "band", band, "error", err)
				continue
			}

			if b.Type >= dspi.CrossoverTypeLR2LP {
				if !hasXover {
					hasXover = true
				}

				xstate := ""
				if b.Bypass {
					xstate = " (bypassed)"
				}

				fmt.Printf("    band %d: %s  %.1f Hz%s\n", b.Band, b.Type, b.Freq, xstate)
			}
		}

		if !hasXover {
			fmt.Println("    (no active crossover bands)")
		}
	}

	return nil
}
