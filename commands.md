# dspictl Command Reference

High-level design for the `dspictl` CLI, derived from the
[DSPi USB Control Protocol](https://github.com/WeebLabs/DSPi#usb-control-protocol).

## Design Conventions

- **Targeting:** Every command operates on **all connected devices** by default.
  Use `--target <serial>` to address a single device.
- **Numbering:** Channel and output indices are **0-based** (faithful to the protocol).
- **Flat vs. grouped:** Commands with 1-2 operations are flat subcommands of the
  root. Topics with 3+ operations get their own subcommand group.
- **EQ management** is deferred (not in this iteration).
- **Master mute** uses the `-128 dB` sentinel.

## Flat Commands

### `dspictl mute`

Sets master volume to the mute sentinel (-128 dB) on all targets. No arguments.

Example:
```
dspictl mute
dspictl --target E6614103E32C3B2D mute
```

### `dspictl unmute`

Resets master volume to the firmware default (-20 dB). No arguments.

Example:
```
dspictl unmute
```

### `dspictl status`

Queries and prints a summary of every connected device:

- Platform (RP2040 / RP2350)
- Serial number
- Master volume
- Active preset slot

Example:
```
$ dspictl status
Serial: E6614103E32C3B2D
Type: RP2350
Volume: -20 dB
Preset: 3
```

## Grouped Commands (subcommand groups)

### `dspictl volume`

Master volume get/set.

| Subcommand | Args | Description |
|---|---|---|
| `get` | — | Print current master volume of each target |
| `set` | `<db>` | Set master volume (-128 to 0 dB) |

Examples:
```
dspictl volume get
dspictl volume set -30
dspictl --target E6614103E32C3B2D volume set -6
```

### `dspictl preamp`

Per-channel input preamp get/set.

| Subcommand | Args | Description |
|---|---|---|
| `get` | `[<channel>]` | Show all channels, or one channel's preamp |
| `set` | `<channel> <db>` | Set preamp gain for a channel |

Example:
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

Examples:
```
dspictl output list
dspictl output gain 2 -6
dspictl output mute 3
dspictl output delay 4 5.2
dspictl output disable 9
```

### `dspictl preset`

Preset slot management (slots 0-9).

| Subcommand | Args | Description |
|---|---|---|
| `list` | — | Show all preset slots with names and occupancy |
| `save` | `<slot>` | Save current DSP state to slot |
| `load` | `<slot>` | Load slot into live state |
| `delete` | `<slot>` | Delete (clear) a slot |
| `name` | `<slot> <name>` | Set a slot name |
| `active` | — | Show the currently active preset slot |

Examples:
```
dspictl preset list
dspictl preset save 3
dspictl preset load 1
dspictl preset name 2 "2-Way + Sub"
dspictl preset active
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

Examples:
```
dspictl matrix list
dspictl matrix get 0 1
dspictl matrix set 0 1 -3.0
dspictl matrix enable 1 2
dspictl matrix invert 0 1
```

## Global Flag

- `--target <serial>` — Operate on a specific device by serial number.
  Without it, every command addresses all connected DSPi devices.
