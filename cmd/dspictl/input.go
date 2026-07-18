package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Input source selection and I2S configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "source [<usb|spdif|i2s|adat|spdif2|spdif3>]",
		Short:             "Get or set the active input source",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputSource,
		ValidArgsFunction: completeChoices([]string{"usb", "spdif", "i2s", "adat", "spdif2", "spdif3"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "rate [<44100|48000|96000>]",
		Short:             "Get or set the I2S input sample rate",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputRate,
		ValidArgsFunction: completeChoices([]string{"44100", "48000", "96000"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "channels [<2|4|6|8>]",
		Short:             "Get or set the number of I2S input channels",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputChannels,
		ValidArgsFunction: completeChoices([]string{"2", "4", "6", "8"}),
	})

	adatCmd := &cobra.Command{
		Use:   "adat",
		Short: "ADAT input configuration (RP2350 only)",
	}

	adatCmd.AddCommand(&cobra.Command{
		Use:               "enable [on|off]",
		Short:             "Get or set the ADAT input enable state",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runAdatInputEnable,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	})

	adatCmd.AddCommand(&cobra.Command{
		Use:   "pin [<gpio>]",
		Short: "Get or set the ADAT input RX GPIO pin",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAdatInputPin,
	})

	adatCmd.AddCommand(&cobra.Command{
		Use:               "clock-mode [master|slave]",
		Short:             "Get or set the ADAT input clock mode",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runAdatInputClockMode,
		ValidArgsFunction: completeChoices([]string{"master", "slave"}),
	})

	adatCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show the ADAT input receiver status",
		Args:  cobra.NoArgs,
		RunE:  runAdatInputStatus,
	})

	cmd.AddCommand(adatCmd)

	return cmd
}

func runInputSource(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			src, err := d.GetInputSource()

			if err != nil {
				slog.Error("getting input source failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: %s\n", d.Serial(), dspi.InputSourceName(src))
		}

		return nil
	}

	var source int

	switch args[0] {
	case "usb":
		source = dspi.InputSourceUSB
	case "spdif":
		source = dspi.InputSourceSPDIF
	case "i2s":
		source = dspi.InputSourceI2S
	case "adat":
		source = dspi.InputSourceADAT
	case "spdif2":
		source = dspi.InputSourceSPDIF2
	case "spdif3":
		source = dspi.InputSourceSPDIF3
	default:
		return fmt.Errorf("invalid source: %s (expected usb, spdif, i2s, adat, spdif2, or spdif3)", args[0])
	}

	for _, d := range devices {
		if err := d.SetInputSource(source); err != nil {
			slog.Error("setting input source failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), dspi.InputSourceName(source))
	}

	return nil
}

func runInputRate(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			rate, err := d.GetInputRate()

			if err != nil {
				slog.Error("getting input rate failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: pipeline=%d Hz  selected_i2s=%d Hz\n", d.Serial(), rate.PipelineHz, rate.SelectedHz)
		}

		return nil
	}

	var hz uint32

	switch args[0] {
	case "44100":
		hz = 44100
	case "48000":
		hz = 48000
	case "96000":
		hz = 96000
	default:
		return fmt.Errorf("invalid rate: %s (expected 44100, 48000, or 96000)", args[0])
	}

	for _, d := range devices {
		if err := d.SetInputRate(hz); err != nil {
			slog.Error("setting input rate failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: I2S rate set to %d Hz\n", d.Serial(), hz)
	}

	return nil
}

func runInputChannels(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			ch, err := d.GetI2SInputChannels()

			if err != nil {
				slog.Error("getting I2S input channels failed", "serial", d.Serial(), "error", err)

				continue
			}

			if ch == 0 {
				fmt.Printf("%s: I2S input channels not configured\n", d.Serial())
			} else {
				fmt.Printf("%s: %d I2S input channels\n", d.Serial(), ch)
			}
		}

		return nil
	}

	n, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid channel count: %w", err)
	}

	switch n {
	case 2, 4, 6, 8:
	default:
		return fmt.Errorf("invalid channel count: %d (expected 2, 4, 6, or 8)", n)
	}

	for _, d := range devices {
		if err := d.SetI2SInputChannels(n); err != nil {
			slog.Error("setting I2S input channels failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: I2S input channels set to %d\n", d.Serial(), n)
	}

	return nil
}

func runAdatInputEnable(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			enabled, err := d.GetAdatInputEnable()
			if err != nil {
				slog.Error("getting ADAT input enable failed", "serial", d.Serial(), "error", err)
				continue
			}

			fmt.Printf("%s: ADAT input %s\n", d.Serial(), onOff(enabled))
		}
		return nil
	}

	enabled, err := parseBoolArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid enable value: %w", err)
	}

	for _, d := range devices {
		if err := d.SetAdatInputEnable(enabled); err != nil {
			slog.Error("setting ADAT input enable failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: ADAT input %s\n", d.Serial(), onOff(enabled))
	}

	return nil
}

func runAdatInputPin(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			pin, err := d.GetAdatInputPin()
			if err != nil {
				slog.Error("getting ADAT input pin failed", "serial", d.Serial(), "error", err)
				continue
			}

			if pin == 0xFF {
				fmt.Printf("%s: ADAT input pin unset\n", d.Serial())
			} else {
				fmt.Printf("%s: ADAT input pin=GPIO %d\n", d.Serial(), pin)
			}
		}
		return nil
	}

	pin, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		if err := d.SetAdatInputPin(pin); err != nil {
			slog.Error("setting ADAT input pin failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: ADAT input pin set to GPIO %d\n", d.Serial(), pin)
	}

	return nil
}

func runAdatInputClockMode(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mode, err := d.GetAdatInputClockMode()
			if err != nil {
				slog.Error("getting ADAT input clock mode failed", "serial", d.Serial(), "error", err)
				continue
			}

			fmt.Printf("%s: ADAT input clock mode=%s\n", d.Serial(), dspi.AdatClockModeName(mode))
		}
		return nil
	}

	var mode int
	switch args[0] {
	case "master":
		mode = dspi.AdatClockModeMaster
	case "slave":
		mode = dspi.AdatClockModeSlave
	default:
		return fmt.Errorf("invalid clock mode: %s (expected master or slave)", args[0])
	}

	for _, d := range devices {
		if err := d.SetAdatInputClockMode(mode); err != nil {
			slog.Error("setting ADAT input clock mode failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: ADAT input clock mode=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runAdatInputStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		status, err := d.GetAdatInputStatus()
		if err != nil {
			slog.Error("getting ADAT input status failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  state:       %s\n", status.State)
		fmt.Printf("  clock mode:  %s\n", dspi.AdatClockModeName(status.ClockMode))
		fmt.Printf("  enabled:     %v\n", status.Enabled)
		fmt.Printf("  pin:         ")
		if status.Pin == 0xFF {
			fmt.Println("unset")
		} else {
			fmt.Printf("GPIO %d\n", status.Pin)
		}
		fmt.Printf("  rate ok:     %v\n", status.RateOK)
		fmt.Printf("  lock count:  %d\n", status.LockCount)
		fmt.Printf("  loss count:  %d\n", status.LossCount)
		fmt.Printf("  slip count:  %d\n", status.SlipCount)
		fmt.Printf("  header err:  %d\n", status.HeaderErrors)
		if status.DetectedRate > 0 {
			fmt.Printf("  detected:    %d Hz\n", status.DetectedRate)
		}
		if status.MeasuredHz > 0 {
			fmt.Printf("  measured:    %d Hz\n", status.MeasuredHz)
		}
	}

	return nil
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
