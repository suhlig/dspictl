package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newDiagnosticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "Device diagnostics and monitoring",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "buffer-stats",
		Short: "Read buffer fill statistics",
		RunE:  runDiagnosticsBufferStats,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "usb-errors",
		Short: "Read USB PHY error counters",
		RunE:  runDiagnosticsUSBErrors,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "core1",
		Short: "Query Core 1 operating mode",
		RunE:  runDiagnosticsCore1,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "clear-clips",
		Short: "Clear clip detection latches",
		RunE:  runDiagnosticsClearClips,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "spdif-rx-status",
		Short: "Show S/PDIF RX status (lock, audio, sample rate)",
		RunE:  runDiagnosticsSpdifRxStatus,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "spdif-rx-channel-status",
		Short: "Show raw S/PDIF RX channel status bytes",
		RunE:  runDiagnosticsSpdifRxChannelStatus,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "channels",
		Short: "Show the number of input channels the firmware advertises",
		Long: `Show the number of audio input channels the firmware advertises via USB.

This value is baked into the firmware's USB descriptor and cannot be changed
at runtime. It determines how many channels the host (e.g. a Mac) sees when
the device is enumerated.`,
		RunE: runDiagnosticsChannels,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset-buffer-stats",
		Short: "Reset buffer fill statistics counters",
		RunE:  runDiagnosticsResetBufferStats,
	})

	return cmd
}

func core1ModeName(mode int) string {
	switch mode {
	case 0:
		return "Idle"
	case 1:
		return "PDM"
	case 2:
		return "EQ Worker"
	default:
		return fmt.Sprintf("Unknown (%d)", mode)
	}
}

func runDiagnosticsBufferStats(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		stats, err := d.GetBufferStats()

		if err != nil {
			slog.Error("getting buffer stats failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:", d.Serial())

		for i, b := range stats.Data {
			if i%16 == 0 {
				fmt.Printf("\n  ")
			} else if i%4 == 0 {
				fmt.Printf(" ")
			}

			fmt.Printf("%02x", b)
		}

		fmt.Println()
	}

	return nil
}

func runDiagnosticsUSBErrors(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		stats, err := d.GetUSBErrorStats()

		if err != nil {
			slog.Error("getting USB error stats failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  CRC errors:     %d\n", stats.CRC)
		fmt.Printf("  Bit-stuff:      %d\n", stats.BitStuff)
		fmt.Printf("  Timeout:        %d\n", stats.Timeout)
		fmt.Printf("  Overflow:       %d\n", stats.Overflow)
		fmt.Printf("  Sequence:       %d\n", stats.Sequence)
		fmt.Printf("  Unknown:        %d\n", stats.Unknown)
	}

	return nil
}

func runDiagnosticsCore1(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		mode, err := d.GetCore1Mode()

		if err != nil {
			slog.Error("getting Core 1 mode failed", "serial", d.Serial(), "error", err)

			continue
		}

		conflict, err := d.GetCore1Conflict()

		if err != nil {
			slog.Error("getting Core 1 conflict failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: mode=%s", d.Serial(), core1ModeName(mode))

		if conflict {
			fmt.Printf("  conflict=true")
		}

		fmt.Println()
	}

	return nil
}

func runDiagnosticsClearClips(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.ClearClips()

		if err != nil {
			slog.Error("clearing clips failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: clips cleared\n", d.Serial())
	}

	return nil
}

func runDiagnosticsSpdifRxStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		status, err := d.GetSpdifRxStatus()

		if err != nil {
			slog.Error("getting S/PDIF RX status failed", "serial", d.Serial(), "error", err)

			continue
		}

		lockState := "no lock"

		if status.Locked {
			lockState = "locked"
		}

		audioState := "non-audio"

		if status.Audio {
			audioState = "audio"
		}

		fmt.Printf("%s: %s, %s", d.Serial(), lockState, audioState)

		if status.Rate > 0 {
			fmt.Printf(", %d Hz", status.Rate)
		}

		fmt.Println()
	}

	return nil
}

func runDiagnosticsSpdifRxChannelStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		status, err := d.GetSpdifRxChannelStatus()

		if err != nil {
			slog.Error("getting S/PDIF RX channel status failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:", d.Serial())

		for i, b := range status.Raw {
			if i%16 == 0 {
				fmt.Printf("\n  ")
			} else if i%4 == 0 {
				fmt.Printf(" ")
			}

			fmt.Printf("%02x", b)
		}

		fmt.Println()
	}

	return nil
}

func runDiagnosticsChannels(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		n, err := d.NumInputChannels()

		if err != nil {
			slog.Error("getting input channels failed", "serial", d.Serial(), "error", err)

			continue
		}

		total := 7

		if d.Platform() == dspi.PlatformRP2350 {
			total = 17
		}

		outputs := total - n

		fmt.Printf("%s:\n", d.Serial())
		fmt.Printf("  Input channels:  %d (USB)\n", n)
		fmt.Printf("  Output channels: %d\n", outputs)
		fmt.Printf("  Total channels:  %d\n", total)
	}

	return nil
}

func runDiagnosticsResetBufferStats(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.ResetBufferStats()

		if err != nil {
			slog.Error("resetting buffer stats failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: buffer stats reset\n", d.Serial())
	}

	return nil
}
