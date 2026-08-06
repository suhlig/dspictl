package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newUpmixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upmix",
		Short: "Stereo upmixer (RP2350 only)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show upmixer live telemetry",
		RunE:  runUpmixStatus,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "Show the full upmixer configuration",
		RunE:  runUpmixConfig,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "on",
		Short:             "Enable the upmixer",
		RunE:              runUpmixOn,
		ValidArgsFunction: cobra.NoFileCompletions,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Disable the upmixer",
		RunE:  runUpmixOff,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <param> <value>",
		Short: "Set a single upmixer parameter (enabled, center-mode, surround-mode, strength, center-width, threshold, attack, release, det-hpf, surround-delay, surround-hpf, surround-lpf, decorr, presence)",
		Args:  cobra.ExactArgs(2),
		RunE:  runUpmixSet,
	})

	return cmd
}

func runUpmixStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		st, err := d.GetUpmixStatus()
		if err != nil {
			slog.Error("getting upmix status failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "inactive"
		if st.Active {
			state = "active"
		}
		fmt.Printf("%s: %s (%s)\n", d.Serial(), state, dspi.ParkedReasonName(st.ParkedReason))
		fmt.Printf("  correlation: %d/16384  balance: %d/16384\n", st.CorrQ14, st.BalanceQ14)
		fmt.Printf("  centre gain: %d/32768  Ls: %d/32768  Rs: %d/32768\n", st.CenterGainQ15, st.LsGainQ15, st.RsGainQ15)
	}

	return nil
}

func runUpmixConfig(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		cfg, err := d.GetUpmixConfig()
		if err != nil {
			slog.Error("getting upmix config failed", "serial", d.Serial(), "error", err)
			continue
		}

		state := "off"
		if cfg.Enabled {
			state = "on"
		}
		fmt.Printf("%s: upmixer %s\n", d.Serial(), state)
		fmt.Printf("  centre mode:   %s\n", dspi.UpmixCenterModeName(cfg.CenterMode))
		fmt.Printf("  surround mode: %s\n", dspi.UpmixSurroundModeName(cfg.SurroundMode))
		fmt.Printf("  presence:      %+.1f dB\n", cfg.PresenceDB())
		fmt.Printf("  strength:      %.1f %%\n", cfg.StrengthPct)
		fmt.Printf("  centre width:  %.1f %%\n", cfg.CenterWidthPct)
		fmt.Printf("  threshold:     %.1f %%\n", cfg.CorrThresholdPct)
		fmt.Printf("  attack:        %.1f ms\n", cfg.AttackMs)
		fmt.Printf("  release:       %.1f ms\n", cfg.ReleaseMs)
		fmt.Printf("  det HPF:       %.1f Hz\n", cfg.DetectorHpfHz)
		fmt.Printf("  surround delay: %.1f ms\n", cfg.SurroundDelayMs)
		fmt.Printf("  surround HPF:  %.1f Hz\n", cfg.SurroundHpfHz)
		fmt.Printf("  surround LPF:  %.1f Hz\n", cfg.SurroundLpfHz)
		fmt.Printf("  decorr:        %.1f %%\n", cfg.DecorrPct)
	}

	return nil
}

func runUpmixOn(cmd *cobra.Command, args []string) error {
	return setUpmixEnabled(true)
}

func runUpmixOff(cmd *cobra.Command, args []string) error {
	return setUpmixEnabled(false)
}

func setUpmixEnabled(enabled bool) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	state := "off"
	if enabled {
		state = "on"
	}

	for _, d := range devices {
		if err := d.SetUpmixParam(dspi.UpmixParamEnabled, boolToFloat(enabled)); err != nil {
			slog.Error("setting upmix enabled failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: upmixer %s\n", d.Serial(), state)
	}

	return nil
}

func runUpmixSet(cmd *cobra.Command, args []string) error {
	param, err := dspi.ParseUpmixParam(args[0])
	if err != nil {
		return err
	}

	value, err := strconv.ParseFloat(args[1], 32)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetUpmixParam(param, float32(value)); err != nil {
			slog.Error("setting upmix parameter failed", "serial", d.Serial(), "param", args[0], "error", err)
			continue
		}
		fmt.Printf("%s: %s=%s\n", d.Serial(), dspi.UpmixParamName(param), args[1])
	}

	return nil
}

func boolToFloat(b bool) float32 {
	if b {
		return 1
	}
	return 0
}
