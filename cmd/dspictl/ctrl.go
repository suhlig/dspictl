package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newCtrlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ctrl",
		Short: "External control interfaces (UART / I2C target)",
	}

	uartCmd := &cobra.Command{
		Use:               "uart [on|off]",
		Short:             "Get or set the UART control interface (SET is USB-only)",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runCtrlUart,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	}
	uartCmd.Flags().Int("tx", 0, "TX GPIO pin (pin%%4 == 0)")
	uartCmd.Flags().Int("rx", 0, "RX GPIO pin (pin%%4 == 1), same UART instance")
	uartCmd.Flags().Uint32("baud", 0, "Baud rate (9600..1000000)")
	uartCmd.Flags().Bool("notify", false, "Push async notification frames")
	cmd.AddCommand(uartCmd)

	i2cCmd := &cobra.Command{
		Use:               "i2c [on|off]",
		Short:             "Get or set the I2C target control interface (SET is USB-only)",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runCtrlI2C,
		ValidArgsFunction: completeChoices([]string{"on", "off"}),
	}
	i2cCmd.Flags().Int("sda", 0, "SDA GPIO pin (pin%%2 == 0)")
	i2cCmd.Flags().Int("scl", 0, "SCL GPIO pin (pin%%2 == 1), same I2C instance")
	i2cCmd.Flags().Int("address", 0, "7-bit target address (0x08..0x77)")
	cmd.AddCommand(i2cCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show the live UART/I2C interface status",
		Args:  cobra.NoArgs,
		RunE:  runCtrlStatus,
	})

	return cmd
}

func runCtrlUart(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			cfg, err := d.GetUartConfig()
			if err != nil {
				slog.Error("getting UART config failed", "serial", d.Serial(), "error", err)
				continue
			}

			state := "off"
			if cfg.Enabled {
				state = "on"
			}
			fmt.Printf("%s: UART control %s (TX=GPIO %d, RX=GPIO %d, %d baud, notify %v)\n",
				d.Serial(), state, cfg.TxPin, cfg.RxPin, cfg.Baud, cfg.NotifyEnable)
		}

		return nil
	}

	enabled, err := parseBoolArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	for _, d := range devices {
		cfg, err := d.GetUartConfig()
		if err != nil {
			slog.Error("getting UART config failed", "serial", d.Serial(), "error", err)
			continue
		}

		cfg.Enabled = enabled
		if cmd.Flags().Changed("tx") {
			tx, _ := cmd.Flags().GetInt("tx")
			cfg.TxPin = uint8(tx)
		}
		if cmd.Flags().Changed("rx") {
			rx, _ := cmd.Flags().GetInt("rx")
			cfg.RxPin = uint8(rx)
		}
		if cmd.Flags().Changed("baud") {
			baud, _ := cmd.Flags().GetUint32("baud")
			cfg.Baud = baud
		}
		if cmd.Flags().Changed("notify") {
			notify, _ := cmd.Flags().GetBool("notify")
			cfg.NotifyEnable = notify
		}

		if err := d.SetUartConfig(cfg); err != nil {
			slog.Error("setting UART config failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if enabled {
			state = "on"
		}
		fmt.Printf("%s: UART control %s (TX=GPIO %d, RX=GPIO %d, %d baud, notify %v)\n",
			d.Serial(), state, cfg.TxPin, cfg.RxPin, cfg.Baud, cfg.NotifyEnable)
	}

	return nil
}

func runCtrlI2C(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			cfg, err := d.GetI2CConfig()
			if err != nil {
				slog.Error("getting I2C config failed", "serial", d.Serial(), "error", err)
				continue
			}

			state := "off"
			if cfg.Enabled {
				state = "on"
			}
			fmt.Printf("%s: I2C control %s (SDA=GPIO %d, SCL=GPIO %d, address 0x%02X)\n",
				d.Serial(), state, cfg.SdaPin, cfg.SclPin, cfg.Address)
		}

		return nil
	}

	enabled, err := parseBoolArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	for _, d := range devices {
		cfg, err := d.GetI2CConfig()
		if err != nil {
			slog.Error("getting I2C config failed", "serial", d.Serial(), "error", err)
			continue
		}

		cfg.Enabled = enabled
		if cmd.Flags().Changed("sda") {
			sda, _ := cmd.Flags().GetInt("sda")
			cfg.SdaPin = uint8(sda)
		}
		if cmd.Flags().Changed("scl") {
			scl, _ := cmd.Flags().GetInt("scl")
			cfg.SclPin = uint8(scl)
		}
		if cmd.Flags().Changed("address") {
			address, _ := cmd.Flags().GetInt("address")
			cfg.Address = uint8(address)
		}

		if err := d.SetI2CConfig(cfg); err != nil {
			slog.Error("setting I2C config failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if enabled {
			state = "on"
		}
		fmt.Printf("%s: I2C control %s (SDA=GPIO %d, SCL=GPIO %d, address 0x%02X)\n",
			d.Serial(), state, cfg.SdaPin, cfg.SclPin, cfg.Address)
	}

	return nil
}

func runCtrlStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		st, err := d.GetCtrlIfaceStatus()
		if err != nil {
			slog.Error("getting control interface status failed", "serial", d.Serial(), "error", err)
			continue
		}

		uart := "down"
		if st.UartLive {
			uart = "up"
		}
		i2c := "down"
		if st.I2cLive {
			i2c = "up"
		}
		fmt.Printf("%s: UART %s (last status 0x%02X), I2C %s (last status 0x%02X), proto v%d\n",
			d.Serial(), uart, st.UartLastStatus, i2c, st.I2cLastStatus, st.ProtoVersion)
	}

	return nil
}
