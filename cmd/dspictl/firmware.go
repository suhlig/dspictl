package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newFirmwareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firmware",
		Short: "Firmware management",
	}

	cmd.AddCommand(newFirmwareVersionCmd())
	cmd.AddCommand(newFirmwareUpdateCmd())

	return cmd
}

func newFirmwareVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show firmware version of connected devices",
		RunE:  runFirmwareVersion,
	}
}

func runFirmwareVersion(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	for _, d := range devices {
		bp, err := d.GetAllParams()

		if err != nil {
			slog.Error("getting firmware version failed", "serial", d.Serial(), "error", err)

			continue
		}

		fmt.Printf("%s: %d.%d\n", d.Serial(), bp.Header.FWMajor, bp.Header.FWMinor)
	}

	return nil
}

func newFirmwareUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update <uf2-file>",
		Short: "Update device firmware from a UF2 file",
		Long: `Reboots a DSPi into its UF2 bootloader and flashes new firmware.

A --target is required when more than one device is connected.
The command validates the UF2 file's target platform against the device
before proceeding.`,
		Args: cobra.ExactArgs(1),
		RunE: runFirmwareUpdate,
	}
}

func runFirmwareUpdate(cmd *cobra.Command, args []string) error {
	uf2Path := args[0]

	// Validate the UF2 file before touching any hardware.
	uf2, err := dspi.ParseUF2(uf2Path)

	if err != nil {
		return fmt.Errorf("invalid UF2 file: %w", err)
	}

	uf2Platform, err := dspi.PlatformForFamily(uf2.BoardFamily)

	if err != nil {
		return fmt.Errorf("unsupported UF2 board family 0x%08x — this file is not for a Raspberry Pi Pico", uf2.BoardFamily)
	}

	// Open the target device.
	devices, err := openDevices()

	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}

	defer closeDevices(devices)

	if len(devices) == 0 {
		return fmt.Errorf("no DSPi device found")
	}

	if len(devices) > 1 && targetSerial == "" {
		return fmt.Errorf("multiple devices connected — use --target to specify which device to update")
	}

	dev := devices[0]

	// Get current firmware version for the summary.
	oldVersion := "unknown"

	bp, err := dev.GetAllParams()

	if err == nil {
		oldVersion = fmt.Sprintf("%d.%d", bp.Header.FWMajor, bp.Header.FWMinor)
	}

	// Pre-flight checks.
	if dev.Platform() != uf2Platform {
		return fmt.Errorf(
			"platform mismatch: device is %s but UF2 file targets %s",
			dev.Platform(), uf2Platform,
		)
	}

	serial := dev.Serial()
	plat := dev.Platform()

	fmt.Printf("Device:   %s (%s, firmware %s)\n", serial, plat, oldVersion)
	fmt.Printf("UF2 file: %s (%s, %d blocks)\n", filepath.Base(uf2Path), uf2Platform, uf2.NumBlocks)
	fmt.Printf("Platform: OK\n\n")

	// Reboot into bootloader.
	fmt.Printf("Rebooting into bootloader...\n")

	err = dev.EnterBootloader()

	if err != nil {
		return fmt.Errorf("sending bootloader command: %w", err)
	}

	// Wait for the bootloader volume to appear.
	fmt.Printf("Waiting for bootloader volume...")

	vol, err := waitForBootloaderVolume(30 * time.Second)

	if err != nil {
		fmt.Println()

		return fmt.Errorf("bootloader volume did not appear: %w", err)
	}

	fmt.Printf(" %s\n", vol)

	// Copy UF2 file to the bootloader volume.
	fmt.Printf("Writing firmware...")

	err = copyUF2(uf2Path, vol)

	if err != nil {
		fmt.Println()

		return fmt.Errorf("writing firmware: %w", err)
	}

	fmt.Println(" done")

	// Wait for the volume to disappear (device reboots into new firmware).
	fmt.Printf("Device rebooting...")

	err = waitForVolumeGone(vol, 15*time.Second)

	if err != nil {
		fmt.Println()

		return fmt.Errorf("bootloader volume did not disappear: %w", err)
	}

	fmt.Println(" done")

	// Wait for the device to re-enumerate.
	fmt.Printf("Waiting for device...")

	d, err := waitForDevice(serial, 30*time.Second)

	if err != nil {
		fmt.Println()

		return fmt.Errorf("device did not reappear: %w", err)
	}

	defer d.Close()

	fmt.Println(" connected")

	// Report the new firmware version.
	newVersion := "unknown"

	bp, err = d.GetAllParams()

	if err == nil {
		newVersion = fmt.Sprintf("%d.%d", bp.Header.FWMajor, bp.Header.FWMinor)
	}

	fmt.Printf("\nUpdate complete: firmware %s → %s\n", oldVersion, newVersion)

	return nil
}

// waitForBootloaderVolume polls for an RP2350 or RPI-RP2 volume to appear.
func waitForBootloaderVolume(timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		vol, ok := findBootloaderVolume()

		if ok {
			return vol, nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("timed out after %s", timeout)
}

// waitForVolumeGone polls until the given volume path no longer exists.
func waitForVolumeGone(vol string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if _, err := os.Stat(vol); os.IsNotExist(err) {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timed out after %s", timeout)
}

// waitForDevice polls for a DSPi device with the given serial to re-appear.
func waitForDevice(serial string, timeout time.Duration) (*dspi.Device, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		infos, err := dspi.List()

		if err != nil {
			slog.Warn("listing devices", "error", err)
			time.Sleep(1 * time.Second)

			continue
		}

		for _, info := range infos {
			if info.Serial == serial {
				return dspi.Open(info)
			}
		}

		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("timed out after %s", timeout)
}

// findBootloaderVolume looks for an RP2350 or RPI-RP2 volume on the system.
func findBootloaderVolume() (string, bool) {
	patterns := []string{
		"/Volumes/RP2350*",
		"/Volumes/RPI-RP2*",
		"/media/*/RP2350*",
		"/media/*/RPI-RP2*",
		"/run/media/*/RP2350*",
		"/run/media/*/RPI-RP2*",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)

		if err != nil {
			slog.Warn("globbing for bootloader volume", "pattern", pattern, "error", err)

			continue
		}

		for _, m := range matches {
			// Filter out system paths that happen to match (e.g. /Volumes/Macintosh HD).
			base := filepath.Base(m)

			if !strings.HasPrefix(base, "RP2350") && !strings.HasPrefix(base, "RPI-RP2") {
				continue
			}

			return m, true
		}
	}

	return "", false
}

// copyUF2 copies a UF2 file to the bootloader volume.
// It refuses to write to any volume whose name doesn't start with RP2350 or RPI-RP2.
// On macOS, binaries run from temp directories (via go run) may lack permission
// to write to removable volumes — it falls back to /bin/cp in that case.
func copyUF2(src, vol string) error {
	base := filepath.Base(vol)

	if !strings.HasPrefix(base, "RP2350") && !strings.HasPrefix(base, "RPI-RP2") {
		return fmt.Errorf("refusing to write to non-bootloader volume: %s", vol)
	}

	name := filepath.Base(src)
	dst := filepath.Join(vol, name)

	// Try Go's file copy first.
	if err := copyFileOS(src, dst); err == nil {
		return nil
	}

	// Fall back to /bin/cp (handles macOS TCC restrictions on temp-dir binaries).
	slog.Warn("direct file copy failed, falling back to /bin/cp")

	cp := exec.Command("cp", src, dst)

	if out, err := cp.CombinedOutput(); err != nil {
		return fmt.Errorf("cp failed: %w\n%s", err, string(out))
	}

	return nil
}

// copyFileOS copies a file using the OS kernel.
func copyFileOS(src, dst string) error {
	srcFile, err := os.Open(src)

	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}

	dstFile, err := os.Create(dst)

	if err != nil {
		_ = srcFile.Close()

		return fmt.Errorf("creating destination: %w", err)
	}

	_, err = io.Copy(dstFile, srcFile)

	if closeErr := dstFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	if closeErr := srcFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	if err != nil {
		return fmt.Errorf("copying: %w", err)
	}

	return nil
}
