package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
)

func newPresetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Preset slot management (slots 0-9)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Show all preset slots with names and occupancy",
		RunE:  runPresetList,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "save <slot>",
		Short: "Save current DSP state to slot",
		Args:  cobra.ExactArgs(1),
		RunE:  runPresetSave,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "load <slot>",
		Short: "Load slot into live state",
		Args:  cobra.ExactArgs(1),
		RunE:  runPresetLoad,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <slot>",
		Short: "Delete (clear) a preset slot",
		Args:  cobra.ExactArgs(1),
		RunE:  runPresetDelete,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "name <slot> <name>",
		Short: "Set a slot name",
		Args:  cobra.ExactArgs(2),
		RunE:  runPresetName,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "active",
		Short: "Show the currently active preset slot",
		RunE:  runPresetActive,
	})

	return cmd
}

func runPresetList(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		dir, err := d.GetPresetDirectory()

		if err != nil {
			slog.Error("getting preset directory failed", "serial", d.Serial(), "error", err)

			continue
		}

		active, err := d.GetActivePreset()

		if err != nil {
			slog.Error("getting active preset failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s:\n", d.Serial())

		for slot := range 10 {
			occupied := dir.SlotOccupied&(1<<slot) != 0
			name, err := d.GetPresetName(slot)

			if err != nil {
				name = ""
			}

			marker := " "

			if slot == active {
				marker = "*"
			}

			status := "empty"

			if occupied {
				status = ""
			}

			fmt.Printf("  %s%d", marker, slot)

			if name != "" {
				fmt.Printf("  %s", name)
			}

			if status != "" {
				fmt.Printf("  (%s)", status)
			}

			fmt.Println()
		}
	}

	return nil
}

func runPresetSave(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.PresetSave(slot)

		if err != nil {
			slog.Error("saving preset failed", "serial", d.Serial(), "slot", slot, "error", err)

			continue
		}

		fmt.Printf("%s: saved to slot %d\n", d.Serial(), slot)
	}

	return nil
}

func runPresetLoad(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.PresetLoad(slot)

		if err != nil {
			slog.Error("loading preset failed", "serial", d.Serial(), "slot", slot, "error", err)

			continue
		}

		fmt.Printf("%s: loaded slot %d\n", d.Serial(), slot)
	}

	return nil
}

func runPresetDelete(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.PresetDelete(slot)

		if err != nil {
			slog.Error("deleting preset failed", "serial", d.Serial(), "slot", slot, "error", err)

			continue
		}

		fmt.Printf("%s: deleted slot %d\n", d.Serial(), slot)
	}

	return nil
}

func runPresetName(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	name := args[1]

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetPresetName(slot, name)

		if err != nil {
			slog.Error("setting preset name failed", "serial", d.Serial(), "slot", slot, "error", err)

			continue
		}

		fmt.Printf("%s: slot %d named \"%s\"\n", d.Serial(), slot, name)
	}

	return nil
}

func runPresetActive(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		slot, err := d.GetActivePreset()

		if err != nil {
			slog.Error("getting active preset failed", "serial", d.Serial(), "error", err)

			continue
		}

		name, _ := d.GetPresetName(slot)

		fmt.Printf("%s: slot %d", d.Serial(), slot)

		if name != "" {
			fmt.Printf(" (%s)", name)
		}

		fmt.Println()
	}

	return nil
}
