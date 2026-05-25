package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newMatrixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "matrix",
		Short: "Matrix mixer crosspoint control",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Show all crosspoints with gain, enabled, phase",
		RunE:  runMatrixList,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "get <input> <output>",
		Short:             "Show a single crosspoint",
		Args:              cobra.ExactArgs(2),
		RunE:              runMatrixGet,
		ValidArgsFunction: completeMatrixRoutes,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "set <input> <output> <db>",
		Short:             "Set crosspoint gain",
		Args:              cobra.ExactArgs(3),
		RunE:              runMatrixSet,
		ValidArgsFunction: completeMatrixRoutes,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "enable <input> <output>",
		Short:             "Enable a crosspoint",
		Args:              cobra.ExactArgs(2),
		RunE:              runMatrixEnable,
		ValidArgsFunction: completeMatrixRoutes,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "disable <input> <output>",
		Short:             "Disable a crosspoint",
		Args:              cobra.ExactArgs(2),
		RunE:              runMatrixDisable,
		ValidArgsFunction: completeMatrixRoutes,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "invert <input> <output>",
		Short:             "Toggle phase invert",
		Args:              cobra.ExactArgs(2),
		RunE:              runMatrixInvert,
		ValidArgsFunction: completeMatrixRoutes,
	})

	return cmd
}

func printCrosspoint(d *dspi.Device, in, out int) {
	route, err := d.GetMatrixRoute(in, out)

	if err != nil {
		slog.Error("getting matrix route failed", "serial", d.Serial(), "input", in, "output", out, "error", err)

		return
	}

	phaseStr := ""

	if route.PhaseInvert {
		phaseStr = " INV"
	}

	enableStr := "disabled"

	if route.Enabled {
		enableStr = "enabled"
	}

	fmt.Printf("  L(%d) -> Out%d: gain %s  %s%s\n", in, out, route.Gain, enableStr, phaseStr)
}

func withTwoArgs(args []string) (int, int, error) {
	in, err := strconv.Atoi(args[0])

	if err != nil {
		return 0, 0, fmt.Errorf("invalid input: %w", err)
	}

	out, err := strconv.Atoi(args[1])

	if err != nil {
		return 0, 0, fmt.Errorf("invalid output: %w", err)
	}

	return in, out, nil
}

func runMatrixList(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		numOutputs := 5

		if d.Platform() == dspi.PlatformRP2350 {
			numOutputs = 9
		}

		fmt.Printf("%s:\n", d.Serial())

		for in := range 2 {
			for out := 0; out < numOutputs; out++ {
				printCrosspoint(d, in, out)
			}
		}
	}

	return nil
}

func runMatrixGet(cmd *cobra.Command, args []string) error {
	in, out, err := withTwoArgs(args)

	if err != nil {
		return err
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		printCrosspoint(d, in, out)
	}

	return nil
}

func runMatrixSet(cmd *cobra.Command, args []string) error {
	in, out, err := withTwoArgs(args)

	if err != nil {
		return err
	}

	db, err := strconv.ParseFloat(args[2], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMatrixRoute(&dspi.MatrixRoute{
			Input:   in,
			Output:  out,
			Enabled: true,
			Gain:    dspi.NewGain(db),
		})

		if err != nil {
			slog.Error("setting matrix route failed", "serial", d.Serial(), "input", in, "output", out, "error", err)

			continue
		}

		fmt.Printf("%s: L(%d) -> Out%d: %s\n", d.Serial(), in, out, dspi.NewGain(db))
	}

	return nil
}

func runMatrixEnable(cmd *cobra.Command, args []string) error {
	return setMatrixEnabled(args, true)
}

func runMatrixDisable(cmd *cobra.Command, args []string) error {
	return setMatrixEnabled(args, false)
}

func setMatrixEnabled(args []string, enabled bool) error {
	in, out, err := withTwoArgs(args)

	if err != nil {
		return err
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		route, err := d.GetMatrixRoute(in, out)

		if err != nil {
			slog.Error("getting matrix route failed", "serial", d.Serial(), "input", in, "output", out, "error", err)

			continue
		}

		route.Enabled = enabled

		err = d.SetMatrixRoute(route)

		if err != nil {
			slog.Error("setting matrix route failed", "serial", d.Serial(), "input", in, "output", out, "error", err)

			continue
		}

		state := "enabled"

		if !enabled {
			state = "disabled"
		}

		fmt.Printf("%s: L(%d) -> Out%d: %s\n", d.Serial(), in, out, state)
	}

	return nil
}

func runMatrixInvert(cmd *cobra.Command, args []string) error {
	in, out, err := withTwoArgs(args)

	if err != nil {
		return err
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		route, err := d.GetMatrixRoute(in, out)

		if err != nil {
			slog.Error("getting matrix route failed", "serial", d.Serial(), "input", in, "output", out, "error", err)

			continue
		}

		route.PhaseInvert = !route.PhaseInvert

		err = d.SetMatrixRoute(route)

		if err != nil {
			slog.Error("setting matrix route failed", "serial", d.Serial(), "input", in, "output", out, "error", err)

			continue
		}

		state := "normal"

		if route.PhaseInvert {
			state = "inverted"
		}

		fmt.Printf("%s: L(%d) -> Out%d: phase %s\n", d.Serial(), in, out, state)
	}

	return nil
}
