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
- **EQ management** is deferred.
- **Loudness, crossfeed, and volume leveller** are deferred.
- **Master mute** uses the `-128 dB` sentinel.

## Flat Commands

### `dspictl status`

Queries and prints a summary of every connected device.

```
$ dspictl status
Serial: E6614103E32C3B2D
Type: RP2350
Volume: -20 dB
Preset: 3
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

Hardware configuration: output type, GPIO pins, I2S clocks.

| Subcommand | Args | Description |
|---|---|---|
| `output-type` | `<slot> [spdif\|i2s]` | Get or set slot output type |
| `output-pin` | `<output> [<gpio>]` | Get or set output GPIO pin |
| `i2s-rx-pin` | `[<gpio>]` | Get or set I2S RX data GPIO pin |
| `bck-pin` | `[<gpio>]` | Get or set shared I2S BCK pin (LRCLK = BCK + 1) |
| `output-config-mode` | `[independent\|preset]` | Get or set output configuration persistence |
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
```

## Global Flag

- `--target <serial>` — Operate on a specific device by serial number.
  Without it, every command addresses all connected DSPi devices.
