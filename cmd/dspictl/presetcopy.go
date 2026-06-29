package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

// filterSnapshot holds the complete filter state for copying between sources.
type filterSnapshot struct {
	masterBypass   bool
	eqBands        []dspi.EQBand
	crossoverBands []dspi.CrossoverBand
	bandBypasses   []bandBypassEntry
}

type bandBypassEntry struct {
	Channel int
	Band    int
	Bypass  bool
}

func newPresetCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy DSP state between sources",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "filter <to-slot>",
		Short: "Copy filters from live state or another slot into a preset slot",
		Long: `Copy all filter settings (master EQ, output EQ, crossover, and their bypass
states) into a preset slot.

With one argument, copies the current live state's filters into the slot.
With two arguments, copies filters from the source slot into the target slot.

Examples:
  dspictl preset copy filter 2           # copy live filters to slot 2
  dspictl preset copy filter 0 2         # copy slot 0 filters to slot 2`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runPresetCopyFilter,
	})

	return cmd
}

func runPresetCopyFilter(cmd *cobra.Command, args []string) error {
	toSlot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid target slot: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	var lastErr error

	for _, d := range devices {
		var snap *filterSnapshot

		if len(args) == 2 {
			// Copy from another slot
			fromSlot, err := strconv.Atoi(args[1])
			if err != nil {
				slog.Error("invalid source slot", "serial", d.Serial(), "error", err)
				lastErr = err
				continue
			}

			// Capture filters from the source slot
			err = d.WithPresetSlot(fromSlot, func() error {
				var err error
				snap, err = captureLiveFilters(d)
				return err
			})
			if err != nil {
				slog.Error("capturing filters from source slot", "serial", d.Serial(), "slot", fromSlot, "error", err)
				lastErr = err
				continue
			}
		} else {
			// Copy from live state
			snap, err = captureLiveFilters(d)
			if err != nil {
				slog.Error("capturing live filters", "serial", d.Serial(), "error", err)
				lastErr = err
				continue
			}
		}

		// Apply filters to the target slot
		err = d.WithPresetSlot(toSlot, func() error {
			return applyFilters(d, snap)
		})
		if err != nil {
			slog.Error("applying filters to target slot", "serial", d.Serial(), "slot", toSlot, "error", err)
			lastErr = err
			continue
		}

		slog.Info("filters copied", "serial", d.Serial(), "to-slot", toSlot)
	}

	return lastErr
}

// captureLiveFilters reads all filter settings from the current live state.
func captureLiveFilters(d *dspi.Device) (*filterSnapshot, error) {
	snap := &filterSnapshot{}

	// Master EQ bypass
	bypass, err := d.GetMasterEQBypass()
	if err != nil {
		return nil, fmt.Errorf("getting master EQ bypass: %w", err)
	}
	snap.masterBypass = bypass

	maxBand, err := d.MaxBands()
	if err != nil {
		return nil, fmt.Errorf("getting max bands: %w", err)
	}

	maxChannel := d.MaxEQChannel()

	// Master EQ (channels 0-1) and output EQ (channels 2+)
	for ch := 0; ch <= maxChannel; ch++ {
		for band := range maxBand {
			eb, err := d.GetEQBand(ch, band)
			if err != nil {
				return nil, fmt.Errorf("getting EQ band ch %d band %d: %w", ch, band, err)
			}
			snap.eqBands = append(snap.eqBands, *eb)

			bp, err := d.GetBandBypass(ch, band)
			if err != nil {
				return nil, fmt.Errorf("getting band bypass ch %d band %d: %w", ch, band, err)
			}
			snap.bandBypasses = append(snap.bandBypasses, bandBypassEntry{Channel: ch, Band: band, Bypass: bp})
		}

		// Crossover bands on output channels only (ch >= 2)
		if ch >= 2 {
			for band := 20; band < 20+d.MaxCrossoverBands(); band++ {
				cb, err := d.GetCrossoverBand(ch, band)
				if err != nil {
					return nil, fmt.Errorf("getting crossover band ch %d band %d: %w", ch, band, err)
				}
				snap.crossoverBands = append(snap.crossoverBands, *cb)
			}
		}
	}

	return snap, nil
}

// applyFilters writes all filter settings to the device (assumes live state
// is already loaded with the target preset).
func applyFilters(d *dspi.Device, snap *filterSnapshot) error {
	for i := range snap.eqBands {
		if err := d.SetEQBand(&snap.eqBands[i]); err != nil {
			return fmt.Errorf("setting EQ band ch %d band %d: %w", snap.eqBands[i].Channel, snap.eqBands[i].Band, err)
		}
	}

	for i := range snap.crossoverBands {
		if err := d.SetCrossoverBand(&snap.crossoverBands[i]); err != nil {
			return fmt.Errorf("setting crossover band ch %d band %d: %w", snap.crossoverBands[i].Channel, snap.crossoverBands[i].Band, err)
		}
	}

	for _, bp := range snap.bandBypasses {
		if err := d.SetBandBypass(bp.Channel, bp.Band, bp.Bypass); err != nil {
			return fmt.Errorf("setting band bypass ch %d band %d: %w", bp.Channel, bp.Band, err)
		}
	}

	if err := d.SetMasterEQBypass(snap.masterBypass); err != nil {
		return fmt.Errorf("setting master EQ bypass: %w", err)
	}

	return nil
}
