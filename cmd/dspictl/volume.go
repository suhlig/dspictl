package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume [db]",
		Short: "Master volume get/set, mute, persistence mode, and save",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runVolume,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Print current master volume (alias for `volume` with no args)",
		RunE:  runVolumeGet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <db>",
		Short: "Set master volume (-128 to 0 dB) (alias for `volume <db>`)",
		Args:  cobra.ExactArgs(1),
		RunE:  runVolumeSet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:               "mode [independent|preset]",
		Short:             "Get or set persistence mode",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runVolumeMode,
		ValidArgsFunction: completeChoices([]string{"independent", "preset"}),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "save",
		Short: "Save current volume as boot default",
		RunE:  runVolumeSave,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "mute",
		Short: "Mute master volume",
		RunE:  runVolumeMute,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "unmute",
		Short: "Unmute master volume to firmware default (-20 dB)",
		RunE:  runVolumeUnmute,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "user [<db>]",
		Short: "Get or set UAC/user volume (-60 to 0 dB)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runVolumeUser,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "saved",
		Short: "Show the saved boot-default master volume",
		RunE:  runVolumeSaved,
	})

	return cmd
}

func runVolumeMode(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			mode, err := d.GetMasterVolumeMode()

			if err != nil {
				slog.Error("getting volume mode failed", "serial", d.Serial(), "error", err)

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
		err := d.SetMasterVolumeMode(mode)

		if err != nil {
			slog.Error("setting volume mode failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: mode=%s\n", d.Serial(), args[0])
	}

	return nil
}

func runVolumeSave(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SaveMasterVolume()

		if err != nil {
			slog.Error("saving master volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: master volume saved\n", d.Serial())
	}

	return nil
}

func runVolume(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runVolumeGet(cmd, args)
	}

	return runVolumeSet(cmd, args)
}

func runVolumeGet(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		gain, err := d.GetMasterVolume()

		if err != nil {
			slog.Error("getting master volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), gain)
	}

	return nil
}

func runVolumeSet(cmd *cobra.Command, args []string) error {
	db, err := strconv.ParseFloat(args[0], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMasterVolume(dspi.NewGain(db))

		if err != nil {
			slog.Error("setting master volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %s\n", d.Serial(), dspi.NewGain(db))
	}

	return nil
}

func runVolumeMute(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMasterVolume(dspi.NewGain(-128))

		if err != nil {
			return fmt.Errorf("%s: %w", d.Serial(), err)
		}

		fmt.Printf("%s: muted\n", d.Serial())
	}

	return nil
}

func runVolumeUnmute(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		err := d.SetMasterVolume(dspi.NewGain(-20))

		if err != nil {
			return fmt.Errorf("%s: %w", d.Serial(), err)
		}

		fmt.Printf("%s: unmuted (-20 dB)\n", d.Serial())
	}

	return nil
}

func runVolumeUser(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(args) == 0 {
		for _, d := range devices {
			gain, err := d.GetUserVolume()

			if err != nil {
				slog.Error("getting user volume failed", "serial", d.Serial(), "error", err)

				continue
			}

			fmt.Printf("%s: user volume: %s\n", d.Serial(), gain)
		}

		return nil
	}

	db, err := strconv.ParseFloat(args[0], 64)

	if err != nil {
		return fmt.Errorf("invalid dB value: %w", err)
	}

	for _, d := range devices {
		err := d.SetUserVolume(db)

		if err != nil {
			slog.Error("setting user volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: user volume: %s\n", d.Serial(), dspi.NewGain(db))
	}

	return nil
}

func runVolumeSaved(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		gain, err := d.GetSavedMasterVolume()

		if err != nil {
			slog.Error("getting saved master volume failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: saved master volume: %s\n", d.Serial(), gain)
	}

	return nil
}
