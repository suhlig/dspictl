package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Hardware configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:               "output-type <slot> [spdif|i2s]",
		Short:             "Get or set slot output type",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runConfigOutputType,
		ValidArgsFunction: completeChoices(nil, []string{"spdif", "i2s"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "output-pin <output> [<gpio>]",
		Short: "Get or set output GPIO pin",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runConfigOutputPin,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "bck-pin [<gpio>]",
		Short: "Get or set shared I2S BCK pin",
		Long: `Get or set the shared I2S BCK (bit clock) GPIO pin.

LRCLK (word select) is always BCK + 1 — this is a PIO hardware constraint.
When you set the BCK pin, LRCLK is automatically placed on the next GPIO.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runConfigBckPin,
	})

	mckCmd := &cobra.Command{
		Use:   "mck",
		Short: "I2S master clock configuration",
	}

	mckCmd.AddCommand(&cobra.Command{
		Use:               "enable [true|false]",
		Short:             "Get or set MCK output state",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runConfigMCKEnable,
		ValidArgsFunction: completeChoices([]string{"true", "false"}),
	})

	mckCmd.AddCommand(&cobra.Command{
		Use:   "pin [<gpio>]",
		Short: "Get or set MCK GPIO pin",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runConfigMCKPin,
	})

	mckCmd.AddCommand(&cobra.Command{
		Use:               "multiplier [128|256]",
		Short:             "Get or set MCK multiplier",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runConfigMCKMultiplier,
		ValidArgsFunction: completeChoices([]string{"128", "256"}),
	})

	cmd.AddCommand(mckCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "i2s-rx-pin [<gpio>]",
		Short: "Get or set the I2S RX data GPIO pin",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runConfigI2SRxPin,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "output-config-mode [independent|preset]",
		Short:             "Get or set output configuration persistence mode",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runConfigOutputConfigMode,
		ValidArgsFunction: completeChoices([]string{"independent", "preset"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "export",
		Short: "Export complete DSP state to stdout",
		RunE:  runConfigExport,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "import",
		Short: "Import complete DSP state from stdin",
		RunE:  runConfigImport,
	})

	return cmd
}

func runConfigExport(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(devices) > 1 && targetSerial == "" {
		return fmt.Errorf("multiple devices found; use --target to select one")
	}

	for _, d := range devices {
		bp, err := d.GetAllParams()
		if err != nil {
			return fmt.Errorf("%s: get all params: %w", d.Serial(), err)
		}

		_, err = os.Stdout.Write(bp.Raw)
		if err != nil {
			return fmt.Errorf("writing export: %w", err)
		}
	}

	return nil
}

func runConfigImport(cmd *cobra.Command, args []string) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("no data on stdin")
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(devices) > 1 && targetSerial == "" {
		return fmt.Errorf("multiple devices found; use --target to select one")
	}

	bp := &dspi.BulkParams{
		Header: dspi.ParseBulkHeader(data),
		Raw:    data,
	}

	for _, d := range devices {
		if err := d.SetAllParams(bp); err != nil {
			return fmt.Errorf("%s: set all params: %w", d.Serial(), err)
		}

		slog.Info("imported", "serial", d.Serial(), "bytes", len(data))
	}

	return nil
}

func runConfigOutputType(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 1 {
		for _, d := range devices {
			t, err := d.GetOutputType(slot)

			if err != nil {
				slog.Error("getting output type failed", "serial", d.Serial(), "slot", slot, "error", err)

				continue
			}

			name := "spdif"

			if t == 1 {
				name = "i2s"
			}

			fmt.Printf("%s: slot %d: %s\n", d.Serial(), slot, name)
		}

		return nil
	}

	var t int

	switch args[1] {
	case "spdif":
		t = 0
	case "i2s":
		t = 1
	default:
		return fmt.Errorf("invalid type: %s (expected spdif or i2s)", args[1])
	}

	for _, d := range devices {
		err := d.SetOutputType(slot, t)

		if err != nil {
			slog.Error("setting output type failed", "serial", d.Serial(), "slot", slot, "error", err)

			continue
		}

		fmt.Printf("%s: slot %d: %s\n", d.Serial(), slot, args[1])
	}

	return nil
}

func runConfigOutputPin(cmd *cobra.Command, args []string) error {
	output, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid output: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 1 {
		for _, d := range devices {
			pin, err := d.GetOutputPin(output)

			if err != nil {
				slog.Error("getting output pin failed", "serial", d.Serial(), "output", output, "error", err)

				continue
			}

			fmt.Printf("%s: output %d: GPIO %d\n", d.Serial(), output, pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[1])

	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		err := d.SetOutputPin(output, pin)

		if err != nil {
			slog.Error("setting output pin failed", "serial", d.Serial(), "output", output, "error", err)

			continue
		}

		fmt.Printf("%s: output %d -> GPIO %d\n", d.Serial(), output, pin)
	}

	return nil
}

func runConfigBckPin(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			pin, err := d.GetI2SBckPin()

			if err != nil {
				slog.Error("getting BCK pin failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: BCK=GPIO %d\n", d.Serial(), pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		err := d.SetI2SBckPin(pin)

		if err != nil {
			slog.Error("setting BCK pin failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: BCK=GPIO %d\n", d.Serial(), pin)
	}

	return nil
}

func runConfigMCKEnable(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			enabled, err := d.GetMCKEnable()

			if err != nil {
				slog.Error("getting MCK enable failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: MCK=%v\n", d.Serial(), enabled)
		}

		return nil
	}

	var enabled bool

	switch args[0] {
	case "true":
		enabled = true
	case "false":
		enabled = false
	default:
		return fmt.Errorf("invalid value: %s (expected true or false)", args[0])
	}

	for _, d := range devices {
		err := d.SetMCKEnable(enabled)

		if err != nil {
			slog.Error("setting MCK enable failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: MCK=%v\n", d.Serial(), enabled)
	}

	return nil
}

func runConfigMCKPin(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			pin, err := d.GetMCKPin()

			if err != nil {
				slog.Error("getting MCK pin failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: MCK=GPIO %d\n", d.Serial(), pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		err := d.SetMCKPin(pin)

		if err != nil {
			slog.Error("setting MCK pin failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: MCK=GPIO %d\n", d.Serial(), pin)
	}

	return nil
}

func runConfigMCKMultiplier(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			m, err := d.GetMCKMultiplier()

			if err != nil {
				slog.Error("getting MCK multiplier failed", "serial", d.Serial(), "error", err)

				continue
			}

			label := "128"

			if m == 1 {
				label = "256"
			}

			fmt.Printf("%s: MCK multiplier=%s\n", d.Serial(), label)
		}

		return nil
	}

	var m int

	switch args[0] {
	case "128":
		m = 0
	case "256":
		m = 1
	default:
		return fmt.Errorf("invalid multiplier: %s (expected 128 or 256)", args[0])
	}

	for _, d := range devices {
		err := d.SetMCKMultiplier(m)

		if err != nil {
			slog.Error("setting MCK multiplier failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: MCK multiplier=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runConfigOutputConfigMode(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mode, err := d.GetOutputConfigMode()

			if err != nil {
				slog.Error("getting output config mode failed", "serial", d.Serial(), "error", err)

				continue
			}

			name := "independent"

			if mode == 1 {
				name = "preset"
			}

			fmt.Printf("%s: %s\n", d.Serial(), name)
		}

		return nil
	}

	var mode int

	switch args[0] {
	case "independent":
		mode = 0
	case "preset":
		mode = 1
	default:
		return fmt.Errorf("invalid mode: %s (expected independent or preset)", args[0])
	}

	for _, d := range devices {
		err := d.SetOutputConfigMode(mode)

		if err != nil {
			slog.Error("setting output config mode failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: output-config-mode=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runConfigI2SRxPin(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			pin, err := d.GetI2SRxPin()

			if err != nil {
				slog.Error("getting I2S RX pin failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: I2S RX pin=GPIO %d\n", d.Serial(), pin)
		}

		return nil
	}

	pin, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid GPIO: %w", err)
	}

	for _, d := range devices {
		if err := d.SetI2SRxPin(pin); err != nil {
			slog.Error("setting I2S RX pin failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: I2S RX pin=GPIO %d\n", d.Serial(), pin)
	}

	return nil
}
