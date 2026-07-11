# dspictl Roadmap

High-level design for the `dspictl` CLI, derived from the
[DSPi USB Control Protocol](https://github.com/WeebLabs/DSPi#usb-control-protocol).

# Commands

## Design Conventions

- **Targeting:** Every command operates on **all connected devices** by default.
  Use `--target <serial>` to address a single device.
- **Numbering:** Channel and output indices are **0-based** (faithful to the protocol).
- **Flat vs. grouped:** Commands with 1-2 operations are flat subcommands of the
  root. Topics with 3+ operations get their own subcommand group.
- **Master mute** uses the `-128 dB` sentinel.
- **Get / Set Overloading:** Whenever a command's setter takes only positional arguments (no flags), the command itself is overloaded: called with no arguments it acts as a getter; called with the positional arguments it acts as a setter.  Explicit `get` and `set` subcommands are kept as discoverable aliases.  Commands whose setter requires flags (e.g. `eq master set` with `--type`, `--freq`, etc.) keep the explicit `get`/`set` subcommands as the only path.  This convention makes the CLI concise and predictable.

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
  Sample Rate (raw): 48000 Hz
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

| Usage | Description |
|---|---|
| `volume [db]` | Print current volume, or set it to `<db>` |
| `volume get` | Alias for `volume` (getter) |
| `volume set <db>` | Alias for `volume <db>` |
| `volume mute` | Mute master volume (-128 dB) |
| `volume unmute` | Unmute to firmware default (-20 dB) |
| `volume mode [independent\|preset]` | Get or set persistence mode |
| `volume save` | Save current volume as boot default (mode 0) |
| `volume user [<db>]` | Get or set UAC/user volume (-60 to 0 dB) |
| `volume saved` | Show the saved boot-default master volume |

```
dspictl volume           # get
dspictl volume -30       # set
dspictl volume set -30   # same as above
dspictl volume mute
dspictl volume unmute
dspictl volume mode
dspictl volume mode preset
dspictl volume save
dspictl volume user      # get user volume
dspictl volume user -30  # set user volume
dspictl volume saved     # show saved volume
```

### `dspictl preamp`

Per-channel input preamp get/set.

| Usage | Description |
|---|---|
| `preamp [<channel> [<db>]]` | Show all channels, one channel, or set a channel's preamp |
| `preamp get [channel]` | Alias for `preamp [channel]` |
| `preamp set <channel> <db>` | Alias for `preamp <channel> <db>` |
| `preamp global [<db>]` | Get or set the global preamp (applied before per-channel) |

```
dspictl preamp              # show all channels
dspictl preamp 0            # show USB L only
dspictl preamp 0 -3.5       # USB L to -3.5 dB
dspictl preamp get 0        # same as above
dspictl preamp set 0 -3.5   # same as above
dspictl preamp global       # show global preamp
dspictl preamp global -6    # set global preamp to -6 dB
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
| `eq list` | `<slot>` | Show all filters stored in a preset slot |
| `eq master set` | `<slot> <channel> <band>` | Configure a master EQ band in a preset slot |
| `eq master clear` | `<slot> <channel>` | Reset all master bands to flat in a preset slot |
| `eq master bypass` | `<slot> [true\|false]` | Get or set master EQ bypass in a preset slot |
| `eq master band-bypass` | `<slot> <channel> <band> [true\|false]` | Get or set bypass for a single master band in a preset slot |
| `eq output set` | `<slot> <channel> <band>` | Configure an output EQ band in a preset slot |
| `eq output clear` | `<slot> <channel>` | Reset all output bands to flat in a preset slot |
| `eq output band-bypass` | `<slot> <channel> <band> [true\|false]` | Get or set bypass for a single output band in a preset slot |
| `eq crossover set` | `<slot> <channel> <band>` | Configure a crossover band in a preset slot |
| `eq crossover clear` | `<slot> <channel>` | Reset all crossover bands to flat in a preset slot |
| `eq crossover bypass` | `<slot> <channel> <band> [true\|false]` | Get or set bypass for a crossover band in a preset slot |
| `copy filter` | `<to-slot>` or `<from-slot> <to-slot>` | Copy filters from live state or another slot |

```
dspictl preset list
dspictl preset save 3
dspictl preset load 1
dspictl preset name 2 "2-Way + Sub"
dspictl preset startup-mode last
dspictl preset default-slot 0
dspictl preset eq master set 2 0 0 --type peak --freq 1000 --q 1.4 --gain 3.0
dspictl preset eq master bypass 2 true
dspictl preset eq output set 2 0 0 --type highpass --freq 80 --q 0.7
dspictl preset eq crossover set 2 0 20 --type lr4-lp --freq 800
dspictl preset copy filter 2
dspictl preset copy filter 0 2
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

| Usage | Description |
|---|---|
| `channel-name [<channel> [<name>]]` | Show all names, one name, or set a name |
| `channel-name get [channel]` | Alias for `channel-name [channel]` |
| `channel-name set <channel> <name>` | Alias for `channel-name <channel> <name>` |

```
dspictl channel-name              # show all names
dspictl channel-name 0            # show channel 0 name
dspictl channel-name 2 "Front Left"  # set name
dspictl channel-name get 0        # same as above
dspictl channel-name set 2 "Front Left"  # same as above
```

### `dspictl channel`

Channel gain, mute, unmute, and delay controls. Operates on any channel
(input or output) by index.

| Subcommand | Args | Description |
|---|---|---|
| `gain` | `<channel> <db>` | Set channel gain |
| `mute` | `<channel>` | Mute a channel |
| `unmute` | `<channel>` | Unmute a channel |
| `delay` | `<channel> <ms>` | Set channel delay (0-85 ms) |

```
dspictl channel gain 0 -6
dspictl channel mute 2
dspictl channel unmute 2
dspictl channel delay 1 3.5
```

### `dspictl crossfeed`

Crossfeed (headphone spatialization) control.

| Subcommand | Args | Description |
|---|---|---|
| *(no subcommand)* | — | Show crossfeed status with all parameters |
| `enable` | `[on\|off]` | Get or set crossfeed enable state |
| `preset` | `[<n>]` | Get or set crossfeed preset (0-4) |
| `freq` | `[<hz>]` | Get or set crossover frequency in Hz |
| `feed` | `[<db>]` | Get or set feed level in dB |
| `itd` | `[on\|off]` | Get or set interaural time delay |
| `outputs` | `[on\|off] [<pairs...>]` | Get or set crossfeed output-pair mask |

```
dspictl crossfeed
dspictl crossfeed enable on
dspictl crossfeed preset 2
dspictl crossfeed freq 700
dspictl crossfeed feed -3.0
dspictl crossfeed itd on
dspictl crossfeed outputs              # show active pairs
dspictl crossfeed outputs on 0 1      # enable pairs
dspictl crossfeed outputs off 2       # disable pair
dspictl crossfeed outputs all         # enable all
dspictl crossfeed outputs headphones  # pair 1 only
dspictl crossfeed outputs none        # disable all
```

### `dspictl diagnostics`

Device diagnostics and monitoring.

| Subcommand | Args | Description |
|---|---|---|
| `buffer-stats` | — | Read buffer fill statistics |
| `usb-errors` | — | Read USB PHY error counters |
| `core1` | — | Query Core 1 operating mode |
| `clear-clips` | — | Clear clip detection latches |
| `spdif-rx-status` | — | Show S/PDIF RX status (lock, audio, sample rate) |
| `spdif-rx-channel-status` | — | Show raw S/PDIF RX channel status bytes |
| `channels` | — | Show number of input channels the firmware advertises |
| `reset-buffer-stats` | — | Reset buffer fill statistics counters |

### `dspictl config`

Hardware configuration: output type, GPIO pins, I2S clocks, bulk state
export/import, and persistence commands.

| Subcommand | Args | Description |
|---|---|---|
| `output-type` | `<slot> [--type spdif\|i2s]` | Get or set slot output type |
| `output-pin` | `<output> [<gpio>]` | Get or set output GPIO pin |
| `i2s-rx-pin` | `[--pair <n> --pin <gpio>]` | Get or set I2S RX data GPIO pin for an I2S pair |
| `bck-pin` | `[<gpio>]` | Get or set shared I2S BCK pin (LRCLK = BCK + 1) |
| `output-config-mode` | `[independent\|preset]` | Get or set output configuration persistence |
| `export` | — | Export complete DSP state to stdout |
| `import` | — | Import complete DSP state from stdin |
| `spdif-rx-pin` | `[<gpio>]` | Get or set the S/PDIF RX input GPIO pin |
| `save-output` | — | Save output pin/type configuration to flash |
| `save` | — | Commit all current DSP state (volume, EQ, routing) to flash |
| `mck` | *(see below)* | I2S master clock configuration |

`dspictl config mck` sub-subcommands:

| Subcommand | Args | Description |
|---|---|---|
| `enable` | `[on\|off]` | Get or set MCK output state |
| `pin` | `[<gpio>]` | Get or set MCK GPIO pin |
| `multiplier` | `[128\|256]` | Get or set MCK multiplier |

```
dspictl config output-type 0 i2s
dspictl config output-pin 3 8
dspictl config i2s-rx-pin --pin 15
dspictl config bck-pin 14        # BCK on GPIO 14, LRCLK on GPIO 15
dspictl config output-config-mode preset
dspictl config mck enable on
dspictl config mck pin 13
dspictl config mck multiplier 256
dspictl config export > backup.bin
cat backup.bin | dspictl config import
dspictl config spdif-rx-pin 5
dspictl config save-output
dspictl config save
```

### `dspictl input`

Input source selection and I2S sample rate configuration.

| Subcommand | Args | Description |
|---|---|---|
| `source` | `[usb\|spdif\|i2s]` | Get or set the active input source |
| `rate` | `[44100\|48000\|96000]` | Get or set the I2S input sample rate |
| `channels` | `[2\|4\|6\|8]` | Get or set the number of I2S input channels |

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

### `dspictl dac-mute`

DAC hardware mute control. Manages the mute GPIO pin and timing for
pop-free output muting on power state transitions.

| Usage | Description |
|---|---|
| `dac-mute` | Show current DAC mute configuration |
| `dac-mute on` | Enable DAC hardware mute |
| `dac-mute off` | Disable DAC hardware mute |
| `dac-mute config` | `--enabled --active-low --pin --hold-ms --release-ms` | Configure all DAC mute parameters |
| `dac-mute test` | Run a DAC mute test cycle |

```
dspictl dac-mute
dspictl dac-mute on
dspictl dac-mute config --enabled --active-low --pin 255 --hold-ms 100 --release-ms 50
dspictl dac-mute test
```

### `dspictl loudness`

Loudness compensation (ISO 226:2003 equal-loudness contours).

| Usage | Description |
|---|---|
| `loudness` | Show loudness status |
| `loudness on` | Enable loudness compensation |
| `loudness off` | Disable loudness compensation |
| `loudness reference [spl]` | Get or set reference SPL (40–100 dB) |
| `loudness reference get` | Alias for `loudness reference` |
| `loudness reference set <spl>` | Alias for `loudness reference <spl>` |
| `loudness intensity [pct]` | Get or set intensity (0–200%) |
| `loudness intensity get` | Alias for `loudness intensity` |
| `loudness intensity set <pct>` | Alias for `loudness intensity <pct>` |
| `loudness outputs [on\|off] [<ch...>]` | Get or set per-output loudness mask |

```
dspictl loudness
dspictl loudness on
dspictl loudness reference        # get
dspictl loudness reference 80     # set
dspictl loudness reference get    # same as above
dspictl loudness reference set 80 # same as above
dspictl loudness intensity        # get
dspictl loudness intensity 100    # set
dspictl loudness intensity get    # same as above
dspictl loudness intensity set 100 # same as above
dspictl loudness outputs            # show active outputs
dspictl loudness outputs on 0 1     # enable outputs
dspictl loudness outputs off 2      # disable output
dspictl loudness outputs all        # enable all
dspictl loudness outputs none       # disable all
```

### `dspictl leveller`

Dynamic range compression (leveller) control.

| Subcommand | Args | Description |
|---|---|---|
| *(no subcommand)* | — | Show leveller status with all parameters |
| `enable` | `[on\|off]` | Get or set leveller enable state |
| `amount` | `[<value>]` | Get or set compression amount |
| `speed` | `[<n>]` | Get or set attack/release speed |
| `maxgain` | `[<db>]` | Get or set maximum gain reduction in dB |
| `lookahead` | `[on\|off]` | Get or set lookahead enable |
| `gate` | `[<db>]` | Get or set noise gate threshold in dB |
| `detector-mask [on\|off] [<ch...>]` | Get or set detector input-channel mask |
| `apply-mask [on\|off] [<ch...>]` | Get or set apply input-channel mask |

```
dspictl leveller
dspictl leveller enable on
dspictl leveller amount 10
dspictl leveller speed 3
dspictl leveller maxgain -12
dspictl leveller lookahead on
dspictl leveller gate -80
dspictl leveller detector-mask              # show active inputs
dspictl leveller detector-mask on 0 1       # enable inputs
dspictl leveller detector-mask off 2        # disable input
dspictl leveller detector-mask all          # all inputs (Night mode)
dspictl leveller detector-mask center       # center only (Dialog boost)
dspictl leveller detector-mask front-lr     # front L/R only
dspictl leveller detector-mask none         # disable all
dspictl leveller apply-mask                 # show active inputs
dspictl leveller apply-mask all             # all inputs
```

### `dspictl lg-sound-sync`

LG Sound Sync control for TV audio return over optical. Allows the TV
remote to control volume on the DSPi.

| Usage | Description |
|---|---|
| `lg-sound-sync` | Show current enable state |
| `lg-sound-sync enable [on\|off]` | Get or set LG Sound Sync enable |
| `lg-sound-sync status` | Show detailed status (present, muted, volume) |

```
dspictl lg-sound-sync
dspictl lg-sound-sync enable on
dspictl lg-sound-sync status
```

### `dspictl firmware`

Firmware management.

| Subcommand | Args | Description |
|---|---|---|
| `version` | — | Show firmware version of connected devices |
| `upgrade` | `<uf2-file>` | Upgrade firmware from a UF2 file |
| `enter-bootloader` | — | Reboot device into UF2 bootloader mode |

```
dspictl firmware version
dspictl firmware upgrade dspictl.uf2
dspictl firmware enter-bootloader
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

## Firmware Source

* Do not assume about any features the DSPi firmware might or might not have; always [check the source code](https://github.com/WeebLabs/DSPi/blob/crossover-refactor/Documentation/codebase_map.md#64-dsp-feature-modules). Ask me if you are unsure which branch is current.

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

# Future

## Cross-device copy

The current `preset copy filter` command operates within a single device (or all targeted devices). A natural extension would support sourcing from a different device than the target using `serial:slot` notation:

```
dspictl preset copy filter 0 1                                    # device-local (today)
dspictl preset copy filter D3615A4863503A79:0 1                   # from serial to all targets
dspictl preset copy filter D3615A4863503A79:0 A6353A7489D30615:1  # serial to serial
```

The `--target` flag narrows which devices you operate on, while `serial:slot`
identifies a source device that may differ from the target(s). They are
complementary.

This pattern is not limited to filters — it could extend to matrix, output
config, or other preset aspects if a need arises.
