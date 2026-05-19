package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func newCore1Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "core1",
		Short: "Query Core 1 operating mode",
		RunE:  runCore1,
	}
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

func runCore1(cmd *cobra.Command, args []string) error {
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
