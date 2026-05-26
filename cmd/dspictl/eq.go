package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/suhlig/dspi"
)

func newEQCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eq",
		Short: "Parametric equalizer control",
	}

	cmd.AddCommand(newEQMasterCmd())
	cmd.AddCommand(newEQOutputCmd())

	return cmd
}

func newEQMasterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master",
		Short: "Master EQ (channels 0-1)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Show all active master EQ bands",
		RunE:  runEQMasterList,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "get <channel> <band>",
		Short:             "Show a single EQ band",
		Args:              cobra.ExactArgs(2),
		RunE:              runEQMasterGet,
		ValidArgsFunction: completeEQMasterChannelsAndBands,
	})

	setCmd := &cobra.Command{
		Use:               "set <channel> <band>",
		Short:             "Configure an EQ band",
		Args:              cobra.ExactArgs(2),
		RunE:              runEQMasterSet,
		ValidArgsFunction: completeEQMasterChannelsAndBands,
	}
	setCmd.Flags().String("type", "", "Filter type: flat, peak, lowshelf, highshelf, lowpass, highpass")
	setCmd.Flags().Float64("freq", 0, "Frequency in Hz")
	setCmd.Flags().Float64("q", 0, "Q factor")
	setCmd.Flags().Float64("gain", 0, "Gain in dB")
	_ = setCmd.MarkFlagRequired("type")
	cmd.AddCommand(setCmd)

	cmd.AddCommand(&cobra.Command{
		Use:               "clear <channel>",
		Short:             "Reset all 10 bands to flat",
		Args:              cobra.ExactArgs(1),
		RunE:              runEQMasterClear,
		ValidArgsFunction: completeEQMasterChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "bypass [true|false]",
		Short:             "Get or set master EQ bypass",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runEQMasterBypass,
		ValidArgsFunction: completeChoices([]string{"true", "false"}),
	})

	return cmd
}

func newEQOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Per-output EQ",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "list <channel>",
		Short:             "Show all active EQ bands for an output channel",
		Args:              cobra.ExactArgs(1),
		RunE:              runEQOutputList,
		ValidArgsFunction: completeOutputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "get <channel> <band>",
		Short:             "Show a single EQ band",
		Args:              cobra.ExactArgs(2),
		RunE:              runEQOutputGet,
		ValidArgsFunction: completeEQOutputChannelsAndBands,
	})

	setCmd := &cobra.Command{
		Use:               "set <channel> <band>",
		Short:             "Configure an EQ band",
		Args:              cobra.ExactArgs(2),
		RunE:              runEQOutputSet,
		ValidArgsFunction: completeEQOutputChannelsAndBands,
	}
	setCmd.Flags().String("type", "", "Filter type: flat, peak, lowshelf, highshelf, lowpass, highpass")
	setCmd.Flags().Float64("freq", 0, "Frequency in Hz")
	setCmd.Flags().Float64("q", 0, "Q factor")
	setCmd.Flags().Float64("gain", 0, "Gain in dB")
	_ = setCmd.MarkFlagRequired("type")
	cmd.AddCommand(setCmd)

	cmd.AddCommand(&cobra.Command{
		Use:               "clear <channel>",
		Short:             "Reset all 10 bands to flat",
		Args:              cobra.ExactArgs(1),
		RunE:              runEQOutputClear,
		ValidArgsFunction: completeOutputChannels,
	})

	return cmd
}

func parseEQBandArgs(args []string, flags *pflag.FlagSet, firmwareChannel int) (*dspi.EQBand, error) {
	band, err := strconv.Atoi(args[1])

	if err != nil {
		return nil, fmt.Errorf("invalid band: %w", err)
	}

	filterType, err := flags.GetString("type")

	if err != nil {
		return nil, fmt.Errorf("getting type flag: %w", err)
	}

	t, err := dspi.ParseFilterType(filterType)

	if err != nil {
		return nil, err
	}

	freq, _ := flags.GetFloat64("freq")
	q, _ := flags.GetFloat64("q")
	gain, _ := flags.GetFloat64("gain")

	return &dspi.EQBand{
		Channel:       firmwareChannel,
		Band:          band,
		Type:          t,
		Freq:          freq,
		QualityFactor: q,
		Gain:          gain,
	}, nil
}

func printEQBand(band *dspi.EQBand) {
	if band.Type == dspi.FilterTypeFlat {
		return
	}

	switch band.Type {
	case dspi.FilterTypeLowPass, dspi.FilterTypeHighPass:
		fmt.Printf("    band %d: %s  %.1f Hz  Q %.2f\n", band.Band, band.Type, band.Freq, band.QualityFactor)
	default:
		fmt.Printf("    band %d: %s  %.1f Hz  Q %.2f  %+.1f dB\n", band.Band, band.Type, band.Freq, band.QualityFactor, band.Gain)
	}
}

func runEQMasterList(cmd *cobra.Command, args []string) error {
	return listEQForChannels(0, 1)
}

func runEQOutputList(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return listEQForChannels(ch+2, ch+2)
}

func listEQForChannels(minCh, maxCh int) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		channels, err := d.Channels()

		if err != nil {
			slog.Error("getting channels failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:\n", d.Serial())

		for ch := minCh; ch <= maxCh; ch++ {
			name := channels[ch].Name
			fmt.Printf("  ch %d (%s):\n", ch, name)

			hasActive := false

			for band := range 10 {
				b, err := d.GetEQBand(ch, band)

				if err != nil {
					slog.Error("getting EQ band failed", "serial", d.Serial(), "channel", ch, "band", band, "error", err)

					continue
				}

				if b.Type != dspi.FilterTypeFlat {
					hasActive = true
					printEQBand(b)
				}
			}

			if !hasActive {
				fmt.Println("    (no active bands)")
			}
		}
	}

	return nil
}

func runEQMasterGet(cmd *cobra.Command, args []string) error {
	ch, band, err := parseEQChannelAndBand(args)

	if err != nil {
		return err
	}

	return getEQBand(ch, band)
}

func runEQOutputGet(cmd *cobra.Command, args []string) error {
	ch, band, err := parseEQChannelAndBand(args)

	if err != nil {
		return err
	}

	return getEQBand(ch+2, band)
}

func getEQBand(channel, band int) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		b, err := d.GetEQBand(channel, band)

		if err != nil {
			slog.Error("getting EQ band failed", "serial", d.Serial(), "error", err)

			continue
		}

		if b.Type == dspi.FilterTypeFlat {
			fmt.Printf("%s: ch %d band %d: flat\n", d.Serial(), channel, band)
		} else {
			printEQBandForDevice(d.Serial(), b)
		}
	}

	return nil
}

func printEQBandForDevice(serial string, band *dspi.EQBand) {
	switch band.Type {
	case dspi.FilterTypeLowPass, dspi.FilterTypeHighPass:
		fmt.Printf("%s: ch %d band %d: %s  %.1f Hz  Q %.2f\n", serial, band.Channel, band.Band, band.Type, band.Freq, band.QualityFactor)
	default:
		fmt.Printf("%s: ch %d band %d: %s  %.1f Hz  Q %.2f  %+.1f dB\n", serial, band.Channel, band.Band, band.Type, band.Freq, band.QualityFactor, band.Gain)
	}
}

func runEQMasterSet(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := parseEQBandArgs(args, cmd.Flags(), ch)

	if err != nil {
		return err
	}

	return setEQBandForDevices(band)
}

func runEQOutputSet(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := parseEQBandArgs(args, cmd.Flags(), ch+2)

	if err != nil {
		return err
	}

	return setEQBandForDevices(band)
}

func setEQBandForDevices(band *dspi.EQBand) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetEQBand(band)

		if err != nil {
			slog.Error("setting EQ band failed", "serial", d.Serial(), "error", err)

			continue
		}

		printEQBandForDevice(d.Serial(), band)
	}

	return nil
}

func runEQMasterClear(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return clearEQChannel(ch)
}

func runEQOutputClear(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	return clearEQChannel(ch + 2)
}

func clearEQChannel(channel int) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		for band := range 10 {
			err := d.SetEQBand(&dspi.EQBand{
				Channel: channel,
				Band:    band,
				Type:    dspi.FilterTypeFlat,
			})

			if err != nil {
				slog.Error("clearing EQ band failed", "serial", d.Serial(), "channel", channel, "band", band, "error", err)
			}
		}

		fmt.Printf("%s: ch %d cleared\n", d.Serial(), channel)
	}

	return nil
}

func runEQMasterBypass(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			bypass, err := d.GetMasterEQBypass()

			if err != nil {
				slog.Error("getting bypass failed", "serial", d.Serial(), "error", err)

				continue
			}

			state := "off"

			if bypass {
				state = "on"
			}

			fmt.Printf("%s: bypass=%s\n", d.Serial(), state)
		}

		return nil
	}

	var bypass bool

	switch args[0] {
	case "true":
		bypass = true
	case "false":
		bypass = false
	default:
		return fmt.Errorf("invalid value: %s (expected true or false)", args[0])
	}

	for _, d := range devices {
		err := d.SetMasterEQBypass(bypass)

		if err != nil {
			slog.Error("setting bypass failed", "serial", d.Serial(), "error", err)

			continue
		}

		state := "off"

		if bypass {
			state = "on"
		}

		fmt.Printf("%s: bypass=%s\n", d.Serial(), state)
	}

	return nil
}

func parseEQChannelAndBand(args []string) (int, int, error) {
	ch, err := strconv.Atoi(args[0])

	if err != nil {
		return 0, 0, fmt.Errorf("invalid channel: %w", err)
	}

	band, err := strconv.Atoi(args[1])

	if err != nil {
		return 0, 0, fmt.Errorf("invalid band: %w", err)
	}

	if band < 0 || band > 9 {
		return 0, 0, fmt.Errorf("band %d out of range (0-9)", band)
	}

	return ch, band, nil
}
