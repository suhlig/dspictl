# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-08-06

❗️This major release adds full support for the V28 wire format introduced with DSPi v1.1.5. It is not backwards-compatible with older versions: dspictl 2.x requires the V16+ wire protocol (firmware v1.1.5-beta3 or later) and refuses all USB operations on older firmware with a clear error.

### Added

- Onboard signal generator: `siggen` command group with `start`, `stop`, `config`, `status`, and `types`
- Stereo upmixer (RP2350 only): `upmix` command group with `config`, `status`, `on|off`, and `set <param> <value>` (14 parameters incl. centre presence)
- ADAT bulk output (RP2350 only): `adat enable|pin|status`
- ADAT input (RP2350 only): `input source adat`, `input adat enable|pin|clock-mode|status`
- Psychoacoustic bass: `psybass` command group with cutoff, harmonics, drive, character, original, and output mask
- Control Surfaces: `cs` command group with bindings, capability/status reads, slot names, IR commands and learning, and save/revert
- UART and I2C external control interfaces: `ctrl uart|i2c|status`
- Fourth selectable S/PDIF input: `input source spdif4`
- S/PDIF multi-input configuration: `input spdif-enable <2|3|4>`, `input spdif-pin`, `input spdif-config`
- I2S input clock mode and external-clock slave status: `input clock-mode`, `input slave-status`
- I2S clock-pin mode (unified/split) and slave BCK pin role: `config clock-pin-mode`, `config bck-pin --role`
- Vendor-channel user mute: `volume user-mute`
- First-order low/high pass filter types: `eq master|output set ... --type lowpass1|highpass1`
- Linkwitz Transform filter type: `eq master|output set ... --type linkwitz --qp <Q>`
- Multichannel loudness, crossfeed, and volume leveller output/channel masks
- Named presets matching the console app (Night mode, Dialog boost, Front L/R, Headphones)
- `diagnostics channels` and `diagnostics reset-usb-errors` commands
- Firmware compatibility probe at open: devices whose firmware predates the V16 wire protocol fail every operation with a clear error (except entering the bootloader, so `firmware upgrade` keeps working); `status` and `firmware version` print a warning
- Add CODEBASE_MAP.md and link from README

### Changed

- Update wire format from V24 (5900 B) to V28 (5944 B): append the 44-byte stereo-upmixer section (V25) with its presence byte (V26) and centre-OFF mode (V27), and grow the optional S/PDIF input pins to three entries (V28); the bulk payload also exposes the loudness reference SPL and intensity
- First-order filter types no longer require a Q factor (the firmware ignores it for them)
- Bulk transfers retry with backoff (0.15/0.3/0.6 s, matching the console app) when the firmware is still applying a previous upload
- SetAllParams rejects a snapshot that is not the current wire size with a clear error instead of a STALL
- Update the man page for all new commands and add a troubleshooting entry for pre-V16 firmware
- Update module github.com/cpuguy83/go-md2man/v2 to v2.0.7 (#28)
- Update mislav/bump-homebrew-formula-action action to v4.2 (#33)

### Fixed

- `config export/import` and preset restore now work end to end against v1.1.5 firmware: the snapshot size and field offsets match the V28 layout
- Hardware tests expect the V16+ unified channel model (17 channels on RP2350)

## [2.0.0-rc.2] - 2026-07-18

❗️This is a release candidate for the major new release with support for the V24 wire format introduced with DSPi v1.1.5. It is not backwards-compatible with older versions.

### Added

- Add onboard signal generator support
- Add CODEBASE_MAP.md and link from README

### Changed

- Update mislav/bump-homebrew-formula-action action to v4.2 (#33)
- Tag and push in one line

## [2.0.0-rc.1] - 2026-07-17

❗️This is a release candidate for the major new release with support for the V24 wire format introduced with DSPi v1.1.5. It is not backwards-compatible with older versions.

### Added

- ADAT input support: `input source adat`, `input adat enable|pin|clock-mode|status` (RP2350 only)
- Psychoacoustic bass: `psybass` command group with cutoff, harmonics, drive, character, original, and output mask
- Linkwitz Transform filter type: `eq master|output set ... --type linkwitz --qp <Q>`
- `input source` now accepts `spdif2` and `spdif3` as aliases for the optional S/PDIF inputs

### Changed

- Update wire format from V20 (5876 B) to V24 (5900 B)

## [2.0.0-rc.0] - 2026-07-11

❗️This is a release candidate for the major new release with support for the V16/V20 wire format introduced with DSPi v1.1.5. It is not backwards-compatible with older versions.

### Added

- Multichannel loudness: per-output mask (`loudness outputs on|off|all|none`)
- Multichannel crossfeed: per-pair output mask (`crossfeed outputs on|off|all|headphones|none`)
- Multichannel volume leveller: detector and apply channel masks (`leveller detector-mask`, `leveller apply-mask`)
- Named presets matching the console app (Night mode, Dialog boost, Front L/R, Headphones)
- `diagnostics channels` command

### Changed

- Update wire format from V16 (5864 B) to V20 (5876 B)
- Update module github.com/cpuguy83/go-md2man/v2 to v2.0.7 (#28)

## [1.7.0] - 2026-07-04

### Added

- `preset eq list` command

### Changed

- Rewrite man page

## [1.6.2] - 2026-06-21

### Fixed

- Silently print `(unknown)` for older firmware without input rate

## [1.6.1] - 2026-06-21

### Added

- Add I2S input support
- Add man pages
- Document that LRCLK is always BCK + 1
- Implement crossover filters

### Changed

- Update module charm.land/lipgloss/v2 to v2.0.4 (#14)
- Update module github.com/onsi/ginkgo/v2 to v2.31.0 (#17)
- Update module github.com/onsi/gomega to v1.42.0 (#18)
- Fix SetMCKEnable, SetMCKMultiplier and SetMCKPin
- Allow negative gain for EQ
- Read MAX_BANDS using USB wire protocol
- Prevent firmware from decoding the wrong EQ band index

## [1.6.0] - 2026-06-09

### Changed

- Update USB protocol to [DSPi@bbfd91a](https://github.com/WeebLabs/DSPi/commit/bbfd91a1642c3eed47e2833a88343de797457972)
- Fix decoding of firmware patch version
- Explain how to deal with no default ALSA card defined
- Update module `charm.land/bubbletea/v2` to `v2.0.7` (#12)
- Update module `github.com/spf13/pflag` to `v1.0.10` (#11)

## [1.5.0] - 2026-05-27

### Changed

- Add loudness support

## [1.4.0] - 2026-05-27

### Changed

- Print firmware version in `status`
- Add `firmware status` command
- Add `firmware upgrade` command (replaces the obsolete `bootloader` command)

## [1.3.0] - 2026-05-26

### Added

- EQ management

### Fixed

- Bump and fix actions

## [1.2.1] - 2026-05-25

### Fixed

- Add missing completion for channels

## [1.2.0] - 2026-05-25

### Changed

- Complete arg values on the command-line, too

## [1.1.1] - 2026-05-23

### Added

- Add Linux one-liner install instructions

### Changed

- Publish unversioned release archive names

## [1.1.0] - 2026-05-23

### Changed

- Promote the TUI mixer to be part of dspictl

## [1.0.4] - 2026-05-23

### Added

- Add homenbrew install instructions
- Add bulk transfer and config export / import commands
- Add hardware tests

### Changed

- Indent device details
- Use git-cliff for changelogs

## [1.0.3] - 2026-05-20

### Changed

- Fix homebrew download URL

## [1.0.3] - 2026-05-20

### Changed

- Fix homebrew download URL

## [1.0.2] - 2026-05-20

### Added

- Add tests for the USB control transfer

### Changed

- Bump homebrew tap on release
- Work around race condition in release creation

## [1.0.1] - 2026-05-20

### Added

- Add macOS to build matrix

### Changed

- Minimize binaries
- Use latest available macOS Intel runner

## [1.0.0] - 2026-05-19

Initial release

<!-- generated by git-cliff -->
