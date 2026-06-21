# dspictl Roadmap

High-level design for the `dspictl` CLI, derived from the
[DSPi USB Control Protocol](https://github.com/WeebLabs/DSPi#usb-control-protocol).

# TODO

- `REQ_SAVE_PARAMS` (0x51) and `REQ_LOAD_PARAMS` (0x52)
- Crossfeed
- Volume leveller

# Commands

## Design Conventions

- **Targeting:** Every command operates on **all connected devices** by default.
  Use `--target <serial>` to address a single device.
- **Numbering:** Channel and output indices are **0-based** (faithful to the protocol).
- **Flat vs. grouped:** Commands with 1-2 operations are flat subcommands of the
  root. Topics with 3+ operations get their own subcommand group.
- **Master mute** uses the `-128 dB` sentinel.

## Flat Commands

### `dspictl status`

Queries and prints a summary of every connected device.

```
$ dspictl status
Serial: E6614103E32C3B2D
  USB Bus Number: 1
  USB Device Address: 42
  Type: RP2350
  Firmware: 2.1.0
  Volume: -20 dB
  Preset: 3
  Input: USB
  Rate: 48000 Hz
  MCK: true (GPIO 13, 128×)
  Loudness: disabled
  CPU: 12% / 8%
```

### `dspictl factory-reset`

Resets the live DSP state to factory defaults on each target. Does NOT erase
any preset slots. No arguments.

## Grouped Commands (subcommand groups)

### `dspictl volume`

Master volume get/set, mute, persistence mode, and save.

| Subcommand | Args | Description |
|---|---|---|
| `get` | — | Print current master volume |
| `set` | `<db>` | Set master volume (-128 to 0 dB) |
| `mute` | — | Mute master volume (-128 dB) |
| `unmute` | — | Unmute to firmware default (-20 dB) |
| `mode` | `[independent\|preset]` | Get or set persistence mode |
| `save` | — | Save current volume as boot default (mode 0) |

```
dspictl volume get
dspictl volume set -30
dspictl volume mute
dspictl volume unmute
dspictl volume mode
dspictl volume mode preset
dspictl volume save
```

### `dspictl preamp`

Per-channel input preamp get/set.

| Subcommand | Args | Description |
|---|---|---|
| `get` | `[<channel>]` | Show all channels, or one channel's preamp |
| `set` | `<channel> <db>` | Set preamp gain for a channel |

```
dspictl preamp get          # show all channels
dspictl preamp get 0        # show USB L only
dspictl preamp set 0 -3.5   # USB L to -3.5 dB
```

### `dspictl output`

Per-output gain, mute, delay, and enable/disable.

| Subcommand | Args | Description |
|---|---|---|
| `list` | — | Show all outputs with gain, mute, delay, enable |
| `gain` | `<channel> <db>` | Set output gain |
| `mute` | `<channel>` | Mute output |
| `unmute` | `<channel>` | Unmute output |
| `delay` | `<channel> <ms>` | Set time alignment delay (0-85 ms) |
| `enable` | `<channel>` | Enable output channel |
| `disable` | `<channel>` | Disable output channel |

```
dspictl output list
dspictl output gain 2 -6
dspictl output mute 3
dspictl output delay 4 5.2
dspictl output disable 9
```

### `dspictl preset`

Preset slot management (slots 0-9) and startup configuration.

| Subcommand | Args | Description |
|---|---|---|
| `list` | — | Show all preset slots with names and occupancy |
| `save` | `<slot>` | Save current DSP state to slot |
| `load` | `<slot>` | Load slot into live state |
| `delete` | `<slot>` | Delete (clear) a slot |
| `name` | `<slot> <name>` | Set a slot name |
| `active` | — | Show the currently active preset slot |
| `startup-mode` | `[specified\|last]` | Get or set startup mode |
| `default-slot` | `[<slot>]` | Get or set the default boot slot |

```
dspictl preset list
dspictl preset save 3
dspictl preset load 1
dspictl preset name 2 "2-Way + Sub"
dspictl preset startup-mode last
dspictl preset default-slot 0
```

### `dspictl matrix`

Matrix mixer crosspoint control.

| Subcommand | Args | Description |
|---|---|---|
| `list` | — | Show all crosspoints with gain, enabled, phase |
| `get` | `<input> <output>` | Show a single crosspoint |
| `set` | `<input> <output> <db>` | Set crosspoint gain |
| `enable` | `<input> <output>` | Enable a crosspoint |
| `disable` | `<input> <output>` | Disable a crosspoint |
| `invert` | `<input> <output>` | Toggle phase invert |

```
dspictl matrix list
dspictl matrix get 0 1
dspictl matrix set 0 1 -3.0
dspictl matrix enable 1 2
dspictl matrix invert 0 1
```

### `dspictl channel-name`

Read or write user-configurable names for audio channels. Names live in RAM
and are persisted via `preset save`.

| Subcommand | Args | Description |
|---|---|---|
| `get` | `[<channel>]` | Show all channel names, or one channel |
| `set` | `<channel> <name>` | Set a channel name (max 31 chars) |

```
dspictl channel-name get
dspictl channel-name get 0
dspictl channel-name set 2 "Front Left"
```

### `dspictl diagnostics`

Device diagnostics and monitoring.

| Subcommand | Args | Description |
|---|---|---|
| `buffer-stats` | — | Read buffer fill statistics |
| `usb-errors` | — | Read USB PHY error counters |
| `core1` | — | Query Core 1 operating mode |
| `clear-clips` | — | Clear clip detection latches |

### `dspictl config`

Hardware configuration: output type, GPIO pins, I2S clocks, and bulk state
export/import.

| Subcommand | Args | Description |
|---|---|---|
| `output-type` | `<slot> [spdif\|i2s]` | Get or set slot output type |
| `output-pin` | `<output> [<gpio>]` | Get or set output GPIO pin |
| `i2s-rx-pin` | `[<gpio>]` | Get or set I2S RX data GPIO pin |
| `bck-pin` | `[<gpio>]` | Get or set shared I2S BCK pin (LRCLK = BCK + 1) |
| `output-config-mode` | `[independent\|preset]` | Get or set output configuration persistence |
| `export` | — | Export complete DSP state to stdout |
| `import` | — | Import complete DSP state from stdin |
| `mck` | *(see below)* | I2S master clock configuration |

`dspictl config mck` sub-subcommands:

| Subcommand | Args | Description |
|---|---|---|
| `enable` | `[true\|false]` | Get or set MCK output state |
| `pin` | `[<gpio>]` | Get or set MCK GPIO pin |
| `multiplier` | `[128\|256]` | Get or set MCK multiplier |

```
dspictl config output-type 0 i2s
dspictl config output-pin 3 8
dspictl config i2s-rx-pin 15
dspictl config bck-pin 14        # BCK on GPIO 14, LRCLK on GPIO 15
dspictl config output-config-mode preset
dspictl config mck enable true
dspictl config mck pin 13
dspictl config mck multiplier 256
dspictl config export > backup.bin
cat backup.bin | dspictl config import
```

### `dspictl input`

Input source selection and I2S sample rate configuration.

| Subcommand | Args | Description |
|---|---|---|
| `source` | `[usb\|spdif\|i2s]` | Get or set the active input source |
| `rate` | `[44100\|48000\|96000]` | Get or set the I2S input sample rate |

```
dspictl input source
dspictl input source i2s
dspictl input rate 48000
```

### `dspictl eq`

Parametric equalizer for master channels and per-output channels.

| Subcommand | Args | Description |
|---|---|---|
| `master list` | — | Show all active master EQ bands |
| `master get` | `<channel> <band>` | Show a single master EQ band |
| `master set` | `<channel> <band>` | Configure a master EQ band (requires `--type`) |
| `master clear` | `<channel>` | Reset all master bands to flat |
| `master bypass` | `[true\|false]` | Get or set master EQ bypass |
| `master band-bypass` | `<channel> <band> [true\|false]` | Get or set bypass for a single master band |
| `output list` | `<channel>` | Show all active EQ bands for an output |
| `output get` | `<channel> <band>` | Show a single output EQ band |
| `output set` | `<channel> <band>` | Configure an output EQ band (requires `--type`) |
| `output clear` | `<channel>` | Reset all output bands to flat |
| `output band-bypass` | `<channel> <band> [true\|false]` | Get or set bypass for a single output band |

Band set flags: `--type <flat|peak|lowshelf|highshelf|lowpass|highpass>`, `--freq <Hz>`, `--q <factor>`, `--gain <dB>`.

```
dspictl eq master list
dspictl eq master set 0 0 --type peak --freq 1000 --q 1.0 --gain -3.0
dspictl eq output clear 2
dspictl eq master bypass true
```

| `crossover list` | `<channel>` | Show all crossover bands for an output (bands 20-23) |
| `crossover get` | `<channel> <band>` | Show a single crossover band |
| `crossover set` | `<channel> <band>` | Configure a crossover band (requires `--type`, `--freq`) |
| `crossover clear` | `<channel>` | Reset all crossover bands to flat |
| `crossover bypass` | `<channel> <band> [true\|false]` | Get or set bypass for a crossover band |

Crossover set flags: `--type <lr2-lp|lr2-hp|lr4-lp|lr4-hp|…|bes8-lp|bes8-hp>`, `--freq <Hz>`, `--bypass`.

```
dspictl eq crossover list 0
dspictl eq crossover set 0 20 --type lr4-lp --freq 800
dspictl eq crossover bypass 0 20 true
```

### `dspictl loudness`

Loudness compensation (ISO 226:2003 equal-loudness contours).

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | Show loudness status |
| `on` | — | Enable loudness compensation |
| `off` | — | Disable loudness compensation |
| `reference` | `[spl]` | Get or set reference SPL (40–100 dB) |
| `intensity` | `[pct]` | Get or set intensity (0–200%) |

```
dspictl loudness
dspictl loudness on
dspictl loudness reference 80
dspictl loudness intensity 100
```

### `dspictl firmware`

Firmware management.

| Subcommand | Args | Description |
|---|---|---|
| `version` | — | Show firmware version of connected devices |
| `upgrade` | `<uf2-file>` | Upgrade firmware from a UF2 file |

```
dspictl firmware version
dspictl firmware upgrade dspictl.uf2
```

### `dspictl mixer`

Interactive mixer TUI.

```
dspictl mixer
```

## Global Flag

- `--target <serial>` — Operate on a specific device by serial number.
  Without it, every command addresses all connected DSPi devices.

# Coding Conventions

## Currency

* Do not assume you know about the latest Go version. Always check the [release history](https://go.dev/doc/devel/release) and the release notes of all releases since the version you thought was current, e.g. for [Go 1.25](https://go.dev/doc/go1.25) and [1.26](https://go.dev/doc/go1.26).
* Use the latest stable Go version that's available

## Test

* Implement tests for Ginkgo and Gomega.
* Run tests with

  ```command
  $ go tool ginkgo -tags=hwtest ./...
  ```

  Fall back to

  ```command
  $ go tool ginkgo ./...
  ```

  if there is no hardware available locally.
