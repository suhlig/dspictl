package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newCrossoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crossover",
		Short: "Crossover filter control (output channels only, bands 20-23)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "list <channel>",
		Short:             "Show all crossover bands for an output channel",
		Args:              cobra.ExactArgs(1),
		RunE:              runCrossoverList,
		ValidArgsFunction: completeOutputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "get <channel> <band>",
		Short:             "Show a single crossover band",
		Args:              cobra.ExactArgs(2),
		RunE:              runCrossoverGet,
		ValidArgsFunction: completeCrossoverChannelsAndBands,
	})

	setCmd := &cobra.Command{
		Use:               "set <channel> <band>",
		Short:             "Configure a crossover band",
		Args:              cobra.ExactArgs(2),
		RunE:              runCrossoverSet,
		ValidArgsFunction: completeCrossoverChannelsAndBands,
	}
	setCmd.Flags().String("type", "", "Filter type (e.g. lr4-lp, bw2-hp, bes6-lp)")
	setCmd.Flags().Float64("freq", 0, "Frequency in Hz")
	_ = setCmd.MarkFlagRequired("type")
	cmd.AddCommand(setCmd)

	cmd.AddCommand(&cobra.Command{
		Use:               "clear <channel>",
		Short:             "Reset all crossover bands to flat",
		Args:              cobra.ExactArgs(1),
		RunE:              runCrossoverClear,
		ValidArgsFunction: completeOutputChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "bypass <channel> <band> [on|off]",
		Short:             "Get or set bypass for a crossover band",
		Args:              cobra.RangeArgs(2, 3),
		RunE:              runCrossoverBypass,
		ValidArgsFunction: completeCrossoverChannelsAndBands,
	})

	return cmd
}

func runCrossoverList(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	firmwareChannel := ch + 2

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
		name := channels[firmwareChannel].Name
		fmt.Printf("  ch %d (%s):\n", firmwareChannel, name)

		hasActive := false
		for band := 20; band < 20+d.MaxCrossoverBands(); band++ {
			b, err := d.GetCrossoverBand(firmwareChannel, band)
			if err != nil {
				slog.Error("getting crossover band failed", "serial", d.Serial(), "channel", firmwareChannel, "band", band, "error", err)
				continue
			}

			if b.Type >= dspi.CrossoverTypeLR2LP {
				hasActive = true
				state := ""
				if b.Bypass {
					state = " (bypassed)"
				}
				fmt.Printf("    band %d: %s  %.1f Hz%s\n", b.Band, b.Type, b.Freq, state)
			}
		}

		if !hasActive {
			fmt.Println("    (no active crossover bands)")
		}
	}

	return nil
}

func runCrossoverGet(cmd *cobra.Command, args []string) error {
	ch, band, err := parseEQChannelAndBand(args)
	if err != nil {
		return err
	}

	firmwareChannel := ch + 2

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		b, err := d.GetCrossoverBand(firmwareChannel, band)
		if err != nil {
			slog.Error("getting crossover band failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := ""
		if b.Bypass {
			state = " (bypassed)"
		}
		fmt.Printf("%s: ch %d band %d: %s  %.1f Hz%s\n", d.Serial(), firmwareChannel, b.Band, b.Type, b.Freq, state)
	}

	return nil
}

func runCrossoverSet(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	band, err := strconv.Atoi(args[1])
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

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetCrossoverBand(cb)
		if err != nil {
			slog.Error("setting crossover band failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: ch %d band %d: %s  %.1f Hz\n", d.Serial(), cb.Channel, cb.Band, cb.Type, cb.Freq)
	}

	return nil
}

func runCrossoverClear(cmd *cobra.Command, args []string) error {
	ch, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %w", err)
	}

	firmwareChannel := ch + 2

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		for band := 20; band < 20+d.MaxCrossoverBands(); band++ {
			err := d.SetCrossoverBand(&dspi.CrossoverBand{
				Channel: firmwareChannel,
				Band:    band,
				Type:    dspi.CrossoverTypeLR2LP,
				Freq:    1000,
			})
			if err != nil {
				slog.Error("clearing crossover band failed", "serial", d.Serial(), "channel", firmwareChannel, "band", band, "error", err)
			}
		}

		fmt.Printf("%s: ch %d crossover cleared\n", d.Serial(), firmwareChannel)
	}

	return nil
}

func runCrossoverBypass(cmd *cobra.Command, args []string) error {
	ch, band, err := parseEQChannelAndBand(args)
	if err != nil {
		return err
	}

	firmwareChannel := ch + 2

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 2 {
		for _, d := range devices {
			bypass, err := d.GetBandBypass(firmwareChannel, band)
			if err != nil {
				slog.Error("getting band bypass failed", "serial", d.Serial(), "channel", firmwareChannel, "band", band, "error", err)
				continue
			}

			state := "off"
			if bypass {
				state = "on"
			}
			fmt.Printf("%s: ch %d band %d: bypass=%s\n", d.Serial(), firmwareChannel, band, state)
		}

		return nil
	}

	bypass, err := parseBoolArg(args[2])
	if err != nil {
		return fmt.Errorf("invalid bypass value: %w", err)
	}

	for _, d := range devices {
		err := d.SetBandBypass(firmwareChannel, band, bypass)
		if err != nil {
			slog.Error("setting band bypass failed", "serial", d.Serial(), "channel", firmwareChannel, "band", band, "error", err)
			continue
		}

		state := "off"
		if bypass {
			state = "on"
		}
		fmt.Printf("%s: ch %d band %d: bypass=%s\n", d.Serial(), firmwareChannel, band, state)
	}

	return nil
}
