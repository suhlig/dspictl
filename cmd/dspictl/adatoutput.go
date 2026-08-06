package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

func newAdatOutputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adat",
		Short: "ADAT bulk output (RP2350 only)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "enable [on|off]",
		Short:             "Get or set the ADAT bulk output enable",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runAdatOutputEnable,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pin [<gpio>]",
		Short: "Get or set the ADAT output GPIO pin (default 12; 0xFF resets)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAdatOutputPin,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show ADAT output stream status",
		RunE:  runAdatOutputStatus,
	})

	return cmd
}

func runAdatOutputEnable(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			enabled, err := d.GetAdatOutputEnable()
			if err != nil {
				slog.Error("getting ADAT output enable failed", "serial", d.Serial(), "error", err)
				continue
			}

			state := "off"
			if enabled {
				state = "on"
			}
			fmt.Printf("%s: ADAT output %s\n", d.Serial(), state)
		}

		return nil
	}

	enabled, err := parseBoolArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid enable value: %w", err)
	}

	for _, d := range devices {
		if err := d.SetAdatOutputEnable(enabled); err != nil {
			slog.Error("setting ADAT output enable failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if enabled {
			state = "on"
		}
		fmt.Printf("%s: ADAT output %s\n", d.Serial(), state)
	}

	return nil
}

func runAdatOutputPin(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			pin, err := d.GetAdatOutputPin()
			if err != nil {
				slog.Error("getting ADAT output pin failed", "serial", d.Serial(), "error", err)
				continue
			}

			fmt.Printf("%s: ADAT output pin=GPIO %d\n", d.Serial(), pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		if err := d.SetAdatOutputPin(pin); err != nil {
			slog.Error("setting ADAT output pin failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: ADAT output pin=GPIO %d\n", d.Serial(), pin)
	}

	return nil
}

func runAdatOutputStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		st, err := d.GetAdatOutputStatus()
		if err != nil {
			slog.Error("getting ADAT output status failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: ADAT output\n", d.Serial())
		fmt.Printf("  enabled:  %v\n", st.Enabled)
		fmt.Printf("  active:   %v\n", st.Active)
		fmt.Printf("  pin:      GPIO %d\n", st.Pin)
		fmt.Printf("  rate ok:  %v\n", st.RateOK)
		fmt.Printf("  resyncs:  %d\n", st.ResyncCount)
		fmt.Printf("  slips:    %d\n", st.SlipCount)
	}

	return nil
}
