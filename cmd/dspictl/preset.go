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

	var slotChoices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	cmd.AddCommand(&cobra.Command{
		Use:               "save <slot>",
		Short:             "Save current DSP state to slot",
		Args:              cobra.ExactArgs(1),
		RunE:              runPresetSave,
		ValidArgsFunction: completeChoices(slotChoices),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "load <slot>",
		Short:             "Load slot into live state",
		Args:              cobra.ExactArgs(1),
		RunE:              runPresetLoad,
		ValidArgsFunction: completeChoices(slotChoices),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "delete <slot>",
		Short:             "Delete (clear) a preset slot",
		Args:              cobra.ExactArgs(1),
		RunE:              runPresetDelete,
		ValidArgsFunction: completeChoices(slotChoices),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "name <slot> <name>",
		Short:             "Set a slot name",
		Args:              cobra.ExactArgs(2),
		RunE:              runPresetName,
		ValidArgsFunction: completeChoices(slotChoices, nil),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "active",
		Short: "Show the currently active preset slot",
		RunE:  runPresetActive,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "startup-mode [specified|last]",
		Short:             "Get or set startup mode",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runPresetStartupMode,
		ValidArgsFunction: completeChoices([]string{"specified", "last"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "default-slot [<slot>]",
		Short:             "Get or set default boot slot",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runPresetDefaultSlot,
		ValidArgsFunction: completeChoices(slotChoices),
	})

	cmd.AddCommand(newPresetEQCmd())
	cmd.AddCommand(newPresetCopyCmd())

	return cmd
}

func runPresetStartupMode(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mode, slot, err := d.GetPresetStartup()

			if err != nil {
				slog.Error("getting startup mode failed", "serial", d.Serial(), "error", err)

				continue
			}

			name := "specified"

			if mode == 1 {
				name = "last"
			}

			fmt.Printf("%s: %s (default slot %d)\n", d.Serial(), name, slot)
		}

		return nil
	}

	var mode int

	switch args[0] {
	case "specified":
		mode = 0
	case "last":
		mode = 1
	default:
		return fmt.Errorf("invalid mode: %s (expected specified or last)", args[0])
	}

	for _, d := range devices {
		_, currentSlot, err := d.GetPresetStartup()

		if err != nil {
			slog.Error("getting startup mode failed", "serial", d.Serial(), "error", err)

			continue
		}

		err = d.SetPresetStartup(mode, currentSlot)

		if err != nil {
			slog.Error("setting startup mode failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: startup-mode=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runPresetDefaultSlot(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			_, slot, err := d.GetPresetStartup()

			if err != nil {
				slog.Error("getting default slot failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: default slot %d\n", d.Serial(), slot)
		}

		return nil
	}

	slot, err := strconv.Atoi(args[0])

	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	for _, d := range devices {
		currentMode, _, err := d.GetPresetStartup()

		if err != nil {
			slog.Error("getting startup mode failed", "serial", d.Serial(), "error", err)

			continue
		}

		err = d.SetPresetStartup(currentMode, slot)

		if err != nil {
			slog.Error("setting default slot failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: default-slot=%d\n", d.Serial(), slot)
	}

	return nil
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
