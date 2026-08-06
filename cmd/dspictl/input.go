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
		Use:               "source [<usb|spdif|i2s|adat|spdif2|spdif3|spdif4>]",
		Short:             "Get or set the active input source",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runInputSource,
		ValidArgsFunction: completeChoices([]string{"usb", "spdif", "i2s", "adat", "spdif2", "spdif3", "spdif4"}),
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

	cmd.AddCommand(&cobra.Command{
		Use:               "clock-mode [master|slave]",
		Short:             "Get or set the I2S input clock mode (deferred apply)",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runI2SClockMode,
		ValidArgsFunction: completeChoices([]string{"master", "slave"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "slave-status",
		Short: "Show the I2S external-clock slave lock status",
		Args:  cobra.NoArgs,
		RunE:  runI2SSlaveStatus,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "spdif-enable <2|3|4> [on|off]",
		Short:             "Get or set the enable state of an optional S/PDIF input",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runSpdifInputEnable,
		ValidArgsFunction: completeChoices([]string{"2", "3", "4"}, []string{"on", "off"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "spdif-pin <1|2|3|4> [<gpio>]",
		Short: "Get or set the S/PDIF RX GPIO pin of a specific input (0xFF resets to default)",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runSpdifInputPin,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "spdif-config",
		Short: "List the S/PDIF input inventory (count, enable mask, pins)",
		Args:  cobra.NoArgs,
		RunE:  runSpdifInputConfig,
	})

	return cmd
}

func runI2SClockMode(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mode, err := d.GetI2SClockMode()
			if err != nil {
				slog.Error("getting I2S clock mode failed", "serial", d.Serial(), "error", err)
				continue
			}

			name := "master"
			if mode == dspi.I2SClockModeSlave {
				name = "slave"
			}
			fmt.Printf("%s: I2S clock mode %s\n", d.Serial(), name)
		}

		return nil
	}

	var mode int
	switch args[0] {
	case "master":
		mode = dspi.I2SClockModeMaster
	case "slave":
		mode = dspi.I2SClockModeSlave
	default:
		return fmt.Errorf("invalid clock mode: %s (expected master or slave)", args[0])
	}

	for _, d := range devices {
		if err := d.SetI2SClockMode(mode); err != nil {
			slog.Error("setting I2S clock mode failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: I2S clock mode %s (deferred)\n", d.Serial(), args[0])
	}

	return nil
}

func runI2SSlaveStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		st, err := d.GetI2sSlaveStatus()
		if err != nil {
			slog.Error("getting I2S slave status failed", "serial", d.Serial(), "error", err)
			continue
		}

		clockMode := "master"
		if st.ClockMode == dspi.I2SClockModeSlave {
			clockMode = "slave"
		}
		fmt.Printf("%s: I2S slave %s (clock mode %s)\n", d.Serial(), st.State, clockMode)
		fmt.Printf("  detected rate: %d Hz  measured: %d Hz\n", st.DetectedRate, st.MeasuredHz)
		fmt.Printf("  locks: %d  losses: %d  slips: %d\n", st.LockCount, st.LossCount, st.SlipCount)
	}

	return nil
}

func runSpdifInputEnable(cmd *cobra.Command, args []string) error {
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid input index: %w", err)
	}
	if index < 2 || index > 4 {
		return fmt.Errorf("input index %d out of range (2-4)", index)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 1 {
		for _, d := range devices {
			cfg, err := d.GetSpdifInputConfig()
			if err != nil {
				slog.Error("getting S/PDIF input config failed", "serial", d.Serial(), "error", err)
				continue
			}
			enabled := (cfg.EnableMask>>(index-1))&1 != 0
			state := "off"
			if enabled {
				state = "on"
			}
			fmt.Printf("%s: S/PDIF input %d %s\n", d.Serial(), index, state)
		}
		return nil
	}

	enabled, err := parseBoolArg(args[1])
	if err != nil {
		return fmt.Errorf("invalid enable value: %w", err)
	}

	for _, d := range devices {
		if err := d.SetSpdifInputEnable(index, enabled); err != nil {
			slog.Error("setting S/PDIF input enable failed", "serial", d.Serial(), "input", index, "error", err)
			continue
		}

		state := "off"
		if enabled {
			state = "on"
		}
		fmt.Printf("%s: S/PDIF input %d %s\n", d.Serial(), index, state)
	}

	return nil
}

func runSpdifInputPin(cmd *cobra.Command, args []string) error {
	index, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid input index: %w", err)
	}
	if index < 1 || index > 4 {
		return fmt.Errorf("input index %d out of range (1-4)", index)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 1 {
		for _, d := range devices {
			pin, err := d.GetSpdifRxPinForIndex(index - 1)
			if err != nil {
				slog.Error("getting S/PDIF RX pin failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: S/PDIF input %d pin=GPIO %d\n", d.Serial(), index, pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		if err := d.SetSpdifRxPinForIndex(index-1, pin); err != nil {
			slog.Error("setting S/PDIF RX pin failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: S/PDIF input %d pin=GPIO %d\n", d.Serial(), index, pin)
	}

	return nil
}

func runSpdifInputConfig(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		cfg, err := d.GetSpdifInputConfig()
		if err != nil {
			slog.Error("getting S/PDIF input config failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: S/PDIF inputs (%d)\n", d.Serial(), cfg.Count)
		for i := range cfg.Pins {
			state := "-"
			if (cfg.EnableMask>>i)&1 != 0 {
				state = "on"
			}
			fmt.Printf("  input %d: GPIO %d (%s)\n", i+1, cfg.Pins[i], state)
		}
	}

	return nil
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
	case "spdif4":
		source = dspi.InputSourceSPDIF4
	default:
		return fmt.Errorf("invalid source: %s (expected usb, spdif, i2s, adat, spdif2, spdif3, or spdif4)", args[0])
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
