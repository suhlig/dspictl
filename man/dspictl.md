% DSPICTL 1 "July 2026" "dspictl" "User Commands"

# NAME

**dspictl** - Control DSPi audio devices

# SYNOPSIS

**dspictl** [*--target* *serial*] *command* [*args*]

**dspictl mixer**

# DESCRIPTION

**dspictl** is a command-line tool to control one or more DSPi audio devices over USB.
It can query and modify every aspect of a DSPi device: volume, EQ, matrix routing,
presets, input source, crossover filters, and more.

A full-screen terminal UI is also available as **dspictl mixer**.

Device communication uses a USB control protocol implemented on the DSPi's
Pico SDK firmware. All commands communicate over control transfers, so no
kernel driver is needed beyond a udev rule for regular-user USB access.

## Multi-device

When multiple DSPi devices are connected, commands apply to **all** devices
simultaneously. Use *--target* to operate on a single device by serial number.

# GLOBAL OPTIONS

*--target* *serial*
: Operate on a specific device by serial number instead of all connected devices.

*--help*
: Show help for any command. Use `dspictl *command* --help` for command-specific help.

*--version*
: Print the dspictl version and exit.

# COMMANDS

## status

Show a summary of all connected DSPi devices. The output includes serial number, USB
bus address, platform type (e.g. RP2350), firmware version, master volume, active
preset slot, input source and sample rate, MCK status, loudness compensation status,
and CPU load.

```
dspictl status
```

## firmware

Firmware management.

### firmware version

Display the firmware version of all connected devices.

```
dspictl firmware version
```

### firmware upgrade *uf2-file*

Reboot a DSPi into its UF2 bootloader and flash new firmware.

Validation is performed before flashing: the UF2 file's target platform is checked
against the device. A *--target* is required when more than one device is connected.

```
dspictl firmware upgrade dspi-firmware.uf2
```

### firmware enter-bootloader

Reboot the device into UF2 bootloader mode without flashing. The device
reconnects as a mass storage device for manual firmware updates.

```
dspictl firmware enter-bootloader
```

## factory-reset

Reset the live DSP state to factory defaults. Stored presets are not affected.

```
dspictl factory-reset
```

## volume

Master volume get, set, mute, persistence mode, and save.

With no arguments, prints the current master volume in dB. With a single numeric
argument, sets the volume.

### volume get

Print the current master volume.

```
dspictl volume get
```

### volume set *db*

Set the master volume. Range is -128 to 0 dB. 0 dB is maximum output, -128 dB is
practically silent.

```
dspictl volume set -20
```

### volume mode [independent|preset]

Get or set volume persistence mode.

In *independent* mode (default), the volume setting is independent of preset
load/save operations. In *preset* mode, the volume becomes part of the preset
state and is saved/loaded along with other DSP parameters.

```
dspictl volume mode          # show current mode
dspictl volume mode preset   # switch to preset mode
```

### volume save

Save the current volume level as the boot default. On next power-up, the device
will use this volume instead of the firmware default.

```
dspictl volume save
```

### volume mute

Mute the master output immediately.

```
dspictl volume mute
```

### volume unmute

Unmute the master output. The volume returns to the firmware default of -20 dB.

```
dspictl volume unmute
```

### volume user [*db*]

Get or set the UAC/user volume. The user volume is independent of the master
volume and has a range of -60 to 0 dB. This is typically controlled by the
operating system's volume slider over USB Audio Class.

```
dspictl volume user         # show current user volume
dspictl volume user -30     # set user volume to -30 dB
```

### volume user-mute [on|off]

Get or set the vendor-channel user mute. Unlike the UAC1 mute (driven by the
OS over USB Audio Class), this flag is always honored regardless of input
source, so it works during S/PDIF, I2S, or ADAT playback.

```
dspictl volume user-mute       # show current user mute
dspictl volume user-mute on   # mute
```

### volume saved

Show the saved boot-default master volume that was persisted via
`volume save`.

```
dspictl volume saved
```

## preamp

Per-channel input preamp gain get/set.

With no arguments, shows the preamp gain for both input channels. With only a
channel argument, shows the gain for that channel. With channel and dB value,
sets the preamp.

### preamp get

Show preamp gain for all input channels, or for a single channel.

```
dspictl preamp get           # show both channels
dspictl preamp get 0         # show channel 0 only
```

### preamp set *channel* *db*

Set preamp gain for an input channel. Range depends on the device capabilities.

```
dspictl preamp set 0 -6.0   # set channel 0 to -6 dB
```

### preamp global [*db*]

Get or set the global preamp, applied before the per-channel preamps. This
affects the overall input level.

```
dspictl preamp global         # show current global preamp
dspictl preamp global -6      # set global preamp to -6 dB
```

## output

Per-output gain, mute, delay, enable, and disable.

Output channels are indexed starting at 0, corresponding to the physical output
connectors on the DSPi.

### output list

Show all output channels with their gain, mute state, delay, and enabled status.

```
dspictl output list
```

### output gain *channel* *db*

Set the gain for an output channel.

```
dspictl output gain 2 -3.5
```

### output mute *channel*

Mute an output channel.

```
dspictl output mute 3
```

### output unmute *channel*

Unmute an output channel.

```
dspictl output unmute 3
```

### output delay *channel* *ms*

Set a time-alignment delay on an output channel. Range is 0 to 85 ms.

Useful for aligning drivers in a multi-way speaker system.

```
dspictl output delay 3 2.5
```

### output enable *channel*

Enable an output channel.

```
dspictl output enable 4
```

### output disable *channel*

Disable an output channel. Disabled channels produce no audio.

```
dspictl output disable 4
```

## preset

Preset slot management. The DSPi has 10 preset slots (0-9). Each slot stores the
complete DSP state: volume, EQ, matrix routing, input configuration, crossover
filters, and channel names.

### preset list

Show all preset slots with their names and occupancy status.

```
dspictl preset list
```

### preset save *slot*

Save the current live DSP state into a preset slot.

```
dspictl preset save 2
```

### preset load *slot*

Load a preset slot into the live DSP state.

```
dspictl preset load 2
```

### preset delete *slot*

Delete (clear) a preset slot.

```
dspictl preset delete 2
```

### preset name *slot* [*name*]

Set a human-readable name for a preset slot. Use the *--name* flag as the
canonical form, or pass the name as a positional argument for backward
compatibility.

```
dspictl preset name 2 "Dinner Party"            # positional (backward compat)
dspictl preset name 2 --name "Dinner Party"      # flag (canonical form)
```

### preset active

Show the currently active preset slot number.

```
dspictl preset active
```

### preset startup-mode [specified|last]

Get or set the startup mode.

In *specified* mode, the device loads the default-slot on boot. In *last* mode,
the device loads the slot that was active when it was last powered off.

```
dspictl preset startup-mode             # show current mode
dspictl preset startup-mode specified   # use a fixed boot slot
```

### preset default-slot [*slot*]

Get or set the default boot slot (used when startup-mode is *specified*).

```
dspictl preset default-slot     # show current default
dspictl preset default-slot 0   # set to slot 0
```

### preset eq

Modify or list filter settings (master EQ, output EQ, crossover) in a preset
slot without permanently changing the live audio state. Write commands
internally load the slot, apply the change, save it back, and restore
the original live state. The `list` command loads the slot, reads the
filters, and restores — with a brief audio glitch while the preset is active.

#### preset eq list *slot*

Show all master EQ, output EQ, crossover filter bands, and their bypass
states stored in a preset slot.

```
dspictl preset eq list 2
```

#### preset eq master set *slot* *channel* *band*

Configure a master EQ band in a preset slot. Accepts the same flags as
`eq master set`.

```
dspictl preset eq master set 2 0 0 --type peak --freq 1000 --q 1.4 --gain 3.0
```

#### preset eq master clear *slot* *channel*

Reset all master EQ bands to flat in a preset slot.

```
dspictl preset eq master clear 2 0
```

#### preset eq master bypass *slot* [*on*|*off*]

Get or set the master EQ bypass state in a preset slot.

```
dspictl preset eq master bypass 2          # show bypass state
dspictl preset eq master bypass 2 on       # enable bypass
```

#### preset eq master band-bypass *slot* *channel* *band* [*on*|*off*]

Get or set bypass for a single master EQ band in a preset slot.

```
dspictl preset eq master band-bypass 2 0 0       # show bypass state
dspictl preset eq master band-bypass 2 0 0 on    # enable bypass
```

#### preset eq output set *slot* *channel* *band*

Configure an output EQ band in a preset slot. Accepts the same flags as
`eq output set`.

```
dspictl preset eq output set 2 0 0 --type highpass --freq 80 --q 0.7
```

#### preset eq output clear *slot* *channel*

Reset all output EQ bands to flat in a preset slot.

```
dspictl preset eq output clear 2 0
```

#### preset eq output band-bypass *slot* *channel* *band* [*on*|*off*]

Get or set bypass for a single output EQ band in a preset slot.

```
dspictl preset eq output band-bypass 2 0 0       # show bypass state
dspictl preset eq output band-bypass 2 0 0 on    # enable bypass
```

#### preset eq crossover set *slot* *channel* *band*

Configure a crossover band in a preset slot. Accepts the same flags as
`eq crossover set`.

```
dspictl preset eq crossover set 2 0 20 --type lr4-lp --freq 800
```

#### preset eq crossover clear *slot* *channel*

Reset all crossover bands to flat in a preset slot.

```
dspictl preset eq crossover clear 2 0
```

#### preset eq crossover bypass *slot* *channel* *band* [*on*|*off*]

Get or set bypass for a crossover band in a preset slot.

```
dspictl preset eq crossover bypass 2 0 20       # show bypass state
dspictl preset eq crossover bypass 2 0 20 on    # enable bypass
```

### preset copy filter *to-slot*

Copy all filter settings (master EQ, output EQ, crossover, and their bypass
states) from the current live state into a preset slot.

```
dspictl preset copy filter 2
```

### preset copy filter *from-slot* *to-slot*

Copy all filter settings from one preset slot into another.

```
dspictl preset copy filter 0 2
```

## matrix

Matrix mixer crosspoint control. The DSPi has a full matrix mixer with 2 inputs
and multiple outputs. Each crosspoint has gain, enable, and phase invert settings.

### matrix list

Show all matrix crosspoints with their gain, enable status, and phase.

```
dspictl matrix list
```

### matrix get *input* *output*

Show a single matrix crosspoint.

Input is 0 or 1. Output is 0 through the number of available outputs.

```
dspictl matrix get 0 2
```

### matrix set *input* *output* *db*

Set the gain of a matrix crosspoint.

```
dspictl matrix set 0 2 -6.0
```

### matrix enable *input* *output*

Enable a matrix crosspoint, allowing audio to pass from the input to the output.

```
dspictl matrix enable 0 2
```

### matrix disable *input* *output*

Disable a matrix crosspoint, silencing that path.

```
dspictl matrix disable 0 2
```

### matrix invert *input* *output*

Toggle the phase invert setting for a matrix crosspoint.

```
dspictl matrix invert 0 2
```

## channel-name

Read or write user-configurable channel names. Each channel can be assigned a
name of up to 31 characters.

With no arguments, shows all channel names. With a channel argument, shows the
name for that channel. With channel and name, sets the name.

### channel-name get

Show all channel names, or the name of a single channel.

```
dspictl channel-name get       # show all channels
dspictl channel-name get 2     # show channel 2 name
```

### channel-name set *channel* [*name*]

Set the name for a channel (max 31 characters). Use the *--name* flag as the
canonical form, or pass the name as a positional argument for backward
compatibility.

```
dspictl channel-name set 0 "Left In"              # positional (backward compat)
dspictl channel-name set 0 --name "Left In"        # flag (canonical form)
```

## channel

Channel gain, mute, unmute, and delay controls. Operates on any channel
(input or output) by index.

### channel gain *channel* *db*

Set the gain for a channel.

```
dspictl channel gain 0 -6
```

### channel mute *channel*

Mute a channel.

```
dspictl channel mute 2
```

### channel unmute *channel*

Unmute a channel.

```
dspictl channel unmute 2
```

### channel delay *channel* *ms*

Set a time-alignment delay on a channel. Range is 0 to 85 ms.

```
dspictl channel delay 1 3.5
```

## config

Hardware configuration commands for GPIO pins, I2S settings, and DSP state
export/import.

### config output-type *slot* [*spdif*|*i2s*]

Get or set the output type for a slot. Each slot can be configured as *spdif*
(S/PDIF digital output) or *i2s* (I2S digital audio). Use the *--type* flag as
the canonical form, or pass the type as a positional argument for backward
compatibility.

```
dspictl config output-type 0              # show current type
dspictl config output-type 0 spdif        # positional (backward compat)
dspictl config output-type 0 --type spdif # flag (canonical form)
```

### config output-pin *output* [*gpio*]

Get or set the GPIO pin used for a specific audio output.

```
dspictl config output-pin 2         # show current pin
dspictl config output-pin 2 12      # set to GPIO 12
```

### config bck-pin [*gpio*]

Get or set the shared I2S BCK (bit clock) GPIO pin. The LRCLK (word select) pin is always BCK + 1, as this is a PIO hardware constraint. Use `--role 1` to address the slave-mode pair (meaningful in split clock-pin mode).

```
dspictl config bck-pin                # show master-pair pin
dspictl config bck-pin 10             # set master BCK to GPIO 10 (LRCLK = 11)
dspictl config bck-pin --role 1       # show slave-pair pin
dspictl config bck-pin --role 1 26    # set slave BCK to GPIO 26
```

### config clock-pin-mode [unified|split]

Get or set the I2S clock-pin mode. *unified* (default) shares one BCK/LRCLK
pair for both master and slave clocking; *split* routes the slave role to its
own pair (`bck-pin --role 1`). The firmware returns a PIN_CONFIG_* status
byte.

```
dspictl config clock-pin-mode         # show current mode
dspictl config clock-pin-mode split   # separate slave clock pair
```

### config mck

I2S master clock sub-commands.

#### config mck enable [on|off]

Get or set the MCK (master clock) output state.

```
dspictl config mck enable          # show current state
dspictl config mck enable on       # enable MCK output
```

#### config mck pin [*gpio*]

Get or set the GPIO pin used for MCK output.

```
dspictl config mck pin         # show current pin
dspictl config mck pin 14      # set MCK to GPIO 14
```

#### config mck multiplier [128|256]

Get or set the MCK multiplier. The multiplier determines the master clock
frequency relative to the sample rate.

```
dspictl config mck multiplier         # show current multiplier
dspictl config mck multiplier 256     # set to 256x
```

### config i2s-rx-pin

Get or set the I2S RX (receive) data GPIO pin for an I2S data pair.
With no flags, shows pair 0 pin. Use *--pair* and *--pin* to specify
which pair to query or configure.

```
dspictl config i2s-rx-pin                   # show pair 0 pin
dspictl config i2s-rx-pin --pin 8           # set pair 0 to GPIO 8
dspictl config i2s-rx-pin --pair 2 --pin 12 # set pair 2 to GPIO 12
```

### config output-config-mode [independent|preset]

Get or set the output configuration persistence mode.

In *independent* mode, output settings (gain, mute, delay) are independent of
presets. In *preset* mode, they are part of the preset state.

```
dspictl config output-config-mode          # show current mode
dspictl config output-config-mode preset   # switch to preset mode
```

### config export

Export the complete DSP state to stdout as a binary dump. This includes all
parameters: volume, EQ, matrix, presets, channel names, etc.

```
dspictl config export
```

### config import

Import a complete DSP state from stdin. The input must be a binary dump in the
same format as produced by `config export`.

```
dspictl config export > backup.bin
cat backup.bin | dspictl config import
```

### config spdif-rx-pin [*gpio*]

Get or set the GPIO pin used for S/PDIF RX input.

```
dspictl config spdif-rx-pin        # show current pin
dspictl config spdif-rx-pin 5      # set to GPIO 5
```

### config save-output

Save the current output pin and type configuration to flash so it persists
across power cycles.

```
dspictl config save-output
```

### config save

Commit all current live DSP state (volume, EQ, matrix, routing, channel names)
to flash. This is the master persistence command.

```
dspictl config save
```

## input

Input source selection and I2S configuration.

### input source [usb|spdif|i2s|adat|spdif2|spdif3|spdif4]

Get or set the active input source. The DSPi can receive audio from USB,
S/PDIF, I2S, ADAT, or the optional S/PDIF 2/3/4 inputs.

```
dspictl input source         # show current source
dspictl input source spdif   # switch to S/PDIF input
dspictl input source adat    # switch to ADAT input (RP2350 only)
```

### input rate [44100|48000|96000]

Get or set the I2S input sample rate (in Hz). This only applies when the input
source is set to *i2s*.

```
dspictl input rate          # show current rate
dspictl input rate 48000    # set to 48 kHz
```

### input channels [2|4|6|8]

Get or set the number of I2S input channels (2, 4, 6, or 8). This only applies
when the input source is set to *i2s*. The RP2350 supports up to 8 channels
over up to 4 I2S data pairs.

```
dspictl input channels           # show current channel count
dspictl input channels 8         # set to 8 channels
```

### input adat enable [on|off]

Get or set the ADAT input enable state. ADAT input is only available on the
RP2350 and requires a configured RX pin before it can be enabled.

```
dspictl input adat enable        # show current state
dspictl input adat enable on     # enable ADAT input
```

### input adat pin [<gpio>]

Get or set the ADAT input RX GPIO pin. Set to 255 to clear the pin assignment.
The pin may equal the ADAT output pin for loopback self-testing.

```
dspictl input adat pin      # show current pin
dspictl input adat pin 20   # assign GPIO 20
```

### input adat clock-mode [master|slave]

Get or set the ADAT input clock mode. In *master* mode (default) the device is
the rate authority; in *slave* mode the incoming ADAT stream clocks the DSPi.
The change is deferred if ADAT is the active source.

```
dspictl input adat clock-mode        # show current mode
dspictl input adat clock-mode slave  # slave to external ADAT clock
```

### input adat status

Show the ADAT input receiver status: lock state, clock mode, pin, rate_ok,
lock/loss/slip counts, and detected/measured sample rates.

```
dspictl input adat status
```

### input clock-mode [master|slave]

Get or set the I2S input clock mode. In *master* mode (default) the device
drives BCK/LRCLK and the rate is set via `input rate`; in *slave* mode an
external master drives the clocks and the rate is auto-detected. The change
is deferred to the main loop; the GET returns the live mode until then.

```
dspictl input clock-mode          # show current mode
dspictl input clock-mode slave    # slave to an external clock master
```

### input slave-status

Show the I2S external-clock slave lock status: lock state, detected and
measured rates, and lock/loss/slip counts.

```
dspictl input slave-status
```

### input spdif-enable *input* [on|off]

Get or set the enable state of an optional S/PDIF input (2, 3, or 4). Input 1
is always enabled; an input that is the live or pending source cannot be
disabled. The firmware returns a PIN_CONFIG_* status byte.

```
dspictl input spdif-enable 2         # show state of S/PDIF input 2
dspictl input spdif-enable 2 on      # enable S/PDIF input 2
```

### input spdif-pin *input* [*gpio*]

Get or set the S/PDIF RX GPIO pin of a specific input (1-4). Pass 255 as the
pin to restore that input's platform default (5 / 20 / 21 / 22).

```
dspictl input spdif-pin 2        # show pin of S/PDIF input 2
dspictl input spdif-pin 2 20     # assign GPIO 20 to S/PDIF input 2
```

### input spdif-config

List the S/PDIF input inventory read from the device: input count, enable
mask, and one GPIO per input.

```
dspictl input spdif-config
```

## eq

Parametric equalizer control. Three separate EQ groups are available: master EQ
for the input channels, per-output EQ, and crossover filters.

### eq master

Master EQ control for all input channels (0 and 1 on RP2040, up to 0-7 on RP2350).

#### eq master list

Show all active master EQ bands.

```
dspictl eq master list
```

#### eq master get *channel* *band*

Show a single master EQ band's settings.

```
dspictl eq master get 0 0
```

#### eq master set *channel* *band*

Configure an EQ band. The following flags are available:

*--type* (required)
: Filter type. One of: `flat`, `peak`, `lowshelf`, `highshelf`, `lowpass`, `highpass`, `notch`, `allpass`, `allpass1`, `lowshelf1`, `highshelf1`, `lowpass1`, `highpass1`, `linkwitz`.

*--freq* *hz*
: Center or corner frequency in Hz. For `linkwitz`, this is the driver resonance f0.

*--q* *q*
: Q-factor (bandwidth). For `linkwitz`, this is the driver Q0.

*--gain* *db*
: Gain in dB (not applicable for lowpass/highpass). For `linkwitz`, this is the target fp in Hz.

*--qp* *Q*
: Linkwitz Transform target pole Q (only for `--type linkwitz`).

```
dspictl eq master set 0 0 --type peak --freq 1000 --q 1.4 --gain 3.0
dspictl eq master set 0 0 --type lowshelf --freq 200 --gain -2.0
dspictl eq master set 0 0 --type linkwitz --freq 50 --q 0.5 --gain 25 --qp 0.707
```

#### eq master clear *channel*

Reset all master EQ bands to flat for a channel.

```
dspictl eq master clear 0
```

#### eq master bypass [on|off]

Get or set the global master EQ bypass.

```
dspictl eq master bypass        # show current state
dspictl eq master bypass on     # bypass master EQ
```

#### eq master band-bypass *channel* *band* [on|off]

Get or set bypass for a single master EQ band.

```
dspictl eq master band-bypass 0 0          # show bypass state
dspictl eq master band-bypass 0 0 on       # bypass band
```

### eq output

Per-output EQ control. Each output channel has its own independent EQ.

#### eq output list *channel*

Show all active EQ bands for an output channel.

```
dspictl eq output list 2
```

#### eq output get *channel* *band*

Show a single output EQ band's settings.

```
dspictl eq output get 2 0
```

#### eq output set *channel* *band*

Configure an output EQ band. Same flags as `eq master set`: *--type*, *--freq*,
*--q*, *--gain*, *--qp*.

```
dspictl eq output set 2 0 --type peak --freq 2500 --q 2.0 --gain -4.0
```

#### eq output clear *channel*

Reset all output EQ bands to flat for a channel.

```
dspictl eq output clear 2
```

#### eq output band-bypass *channel* *band* [on|off]

Get or set bypass for a single output EQ band.

```
dspictl eq output band-bypass 2 0 on
```

### eq crossover

Crossover filter control for output channels (bands 20-23). The DSPi supports
a variety of crossover filter types:

*Linkwitz-Riley:* lr2-lp, lr2-hp, lr4-lp, lr4-hp, lr6-lp, lr6-hp, lr8-lp, lr8-hp

*Butterworth:* bw1-lp, bw1-hp, bw2-lp, bw2-hp, bw3-lp, bw3-hp, bw4-lp, bw4-hp,
bw5-lp, bw5-hp, bw6-lp, bw6-hp, bw7-lp, bw7-hp, bw8-lp, bw8-hp

*Bessel:* bes2-lp, bes2-hp, bes4-lp, bes4-hp, bes6-lp, bes6-hp, bes8-lp, bes8-hp

#### eq crossover list *channel*

Show all crossover bands for an output channel.

```
dspictl eq crossover list 2
```

#### eq crossover get *channel* *band*

Show a single crossover band.

```
dspictl eq crossover get 2 20
```

#### eq crossover set *channel* *band*

Configure a crossover band. The following flags are available:

*--type* (required)
: Filter type (see list above), e.g. `lr4-lp`, `bw2-hp`, `bes6-lp`.

*--freq* *hz*
: Crossover frequency in Hz.

```
dspictl eq crossover set 2 20 --type lr4-lp --freq 800
dspictl eq crossover set 2 21 --type lr4-hp --freq 800
```

#### eq crossover clear *channel*

Reset all crossover bands to flat for an output channel.

```
dspictl eq crossover clear 2
```

#### eq crossover bypass *channel* *band* [on|off]

Get or set bypass for a crossover band.

```
dspictl eq crossover bypass 2 20          # show bypass state
dspictl eq crossover bypass 2 20 on       # bypass this band
```

## loudness

Loudness compensation based on ISO 226:2003 equal-loudness contours. This
adjusts the frequency response to compensate for the human ear's reduced
sensitivity at low volumes.

With no arguments, shows the current loudness status (on/off, reference SPL,
intensity).

### loudness on

Enable loudness compensation.

```
dspictl loudness on
```

### loudness off

Disable loudness compensation.

```
dspictl loudness off
```

### loudness reference [spl]

Get or set the reference SPL (40-100 dB). The reference SPL is the listening
level at which the frequency response should be flat. Lower values produce more
compensation.

```
dspictl loudness reference       # show current reference
dspictl loudness reference 75    # set to 75 dB SPL
```

### loudness reference get

Show the current reference SPL.

```
dspictl loudness reference get
```

### loudness reference set *spl*

Set the reference SPL.

```
dspictl loudness reference set 80
```

### loudness intensity [pct]

Get or set the loudness compensation intensity (0-200%). At 0% no compensation
is applied, at 100% full ISO 226 correction is applied.

```
dspictl loudness intensity         # show current intensity
dspictl loudness intensity 80      # set to 80%
```

### loudness intensity get

Show the current compensation intensity.

```
dspictl loudness intensity get
```

### loudness intensity set *pct*

Set the compensation intensity.

```
dspictl loudness intensity set 100
```

### loudness outputs [on|off] [<channels...>]

Get or set which output channels are compensated (V19+).

With no arguments, shows the current active outputs. With `on` or `off`
followed by channel numbers, toggles specific outputs. With a preset name,
sets the mask to a predefined value.

Presets:

- **all** – all outputs (default)
- **none** – disable all outputs

```
dspictl loudness outputs            # show active outputs
dspictl loudness outputs on 0 1     # enable outputs 0 and 1
dspictl loudness outputs off 2      # disable output 2
dspictl loudness outputs all        # enable all outputs
dspictl loudness outputs none       # disable all
```

## diagnostics

Device diagnostics and monitoring. These commands expose internal device metrics
useful for debugging.

### diagnostics buffer-stats

Read buffer fill statistics from the DSPi firmware.

```
dspictl diagnostics buffer-stats
```

### diagnostics usb-errors

Read USB PHY error counters from the device.

```
dspictl diagnostics usb-errors
```

### diagnostics reset-usb-errors

Reset the USB PHY error counters (a no-op under TinyUSB, acknowledged
with a status byte).

```
dspictl diagnostics reset-usb-errors
```

### diagnostics core1

Query Core 1 operating mode (on dual-core platforms).

```
dspictl diagnostics core1
```

### diagnostics clear-clips

Clear clip detection latches that may have been triggered by signal overload.

```
dspictl diagnostics clear-clips
```

### diagnostics spdif-rx-status

Show S/PDIF RX status including lock state, audio/non-audio detection, and
sample rate.

```
dspictl diagnostics spdif-rx-status
```

### diagnostics spdif-rx-channel-status

Show the raw S/PDIF RX channel status bytes (24 bytes of IEC 60958 channel
status data).

```
dspictl diagnostics spdif-rx-channel-status
```

### diagnostics adat-input-status

Show ADAT input receiver status including lock state, clock mode, configured
pin, rate_ok flag, lock/loss/slip counts, and detected/measured sample rates.
Only available on the RP2350.

```
dspictl diagnostics adat-input-status
```

### diagnostics channels

Show the number of audio input channels the firmware advertises over USB.
This value is baked into the firmware's USB descriptor and is fixed at
compile time. It determines how many channels the host (e.g. a Mac) sees
when the device is enumerated.

```
dspictl diagnostics channels
```

### diagnostics reset-buffer-stats

Reset the buffer fill statistics counters.

```
dspictl diagnostics reset-buffer-stats
```

## psybass

Psychoacoustic bass enhancement (missing-fundamental harmonics). Synthesizes
higher harmonics from low-frequency content so the ear perceives bass that the
physical speaker cannot reproduce.

With no subcommand, shows the psybass status with all parameters.

```
dspictl psybass
```

### psybass on

Enable psychoacoustic bass on all outputs.

```
dspictl psybass on
```

### psybass off

Disable psychoacoustic bass.

```
dspictl psybass off
```

### psybass cutoff [<hz>]

Get or set the speaker low-frequency limit in Hz.

```
dspictl psybass cutoff        # show current cutoff
dspictl psybass cutoff 80     # set to 80 Hz
```

### psybass harmonics [<db>]

Get or set the harmonic mix level in dB.

```
dspictl psybass harmonics       # show current level
dspictl psybass harmonics -12   # set to -12 dB
```

### psybass drive [<db>]

Get or set the odd-path clipper drive in dB.

```
dspictl psybass drive     # show current drive
dspictl psybass drive 6   # set to 6 dB
```

### psybass character [<pct>]

Get or set the even/odd harmonic blend percentage.

```
dspictl psybass character      # show current blend
dspictl psybass character 50   # set to 50%
```

### psybass original [<db>]

Get or set the original low-band level in dB.

```
dspictl psybass original        # show current level
dspictl psybass original -30    # set to -30 dB
```

### psybass outputs [on|off] [<channels...>]

Get or set which output channels are processed.

With no arguments, shows the current active outputs. With `on` or `off`
followed by channel numbers, toggles specific outputs. With a preset name, sets
the mask to a predefined value.

Presets:

- **all** – all outputs (default)
- **none** – disable all outputs

```
dspictl psybass outputs            # show active outputs
dspictl psybass outputs on 0 1     # enable outputs 0 and 1
dspictl psybass outputs off 2      # disable output 2
dspictl psybass outputs all        # enable all outputs
dspictl psybass outputs none       # disable all
```

## crossfeed

Crossfeed (headphone spatialization) control. Crossfeed mixes a small amount
of the left channel into the right and vice versa, with frequency-dependent
filtering and delay, to simulate the natural crosstalk that occurs when
listening to loudspeakers.

With no subcommand, shows the crossfeed status with all parameters.

### crossfeed enable [on|off]

Get or set crossfeed enable state.

```
dspictl crossfeed enable on
```

### crossfeed preset [*n*]

Get or set crossfeed preset (0-4).

```
dspictl crossfeed preset 2
```

### crossfeed freq [*hz*]

Get or set the crossfeed crossover frequency in Hz.

```
dspictl crossfeed freq 700
```

### crossfeed feed [*db*]

Get or set the crossfeed feed level in dB.

```
dspictl crossfeed feed -3.0
```

### crossfeed itd [on|off]

Get or set the interaural time delay (ITD) feature.

```
dspictl crossfeed itd on
```

### crossfeed outputs [on|off] [<pairs...>]

Get or set which output pairs are crossfed (V20+). With no arguments, shows
the current active pairs. With `on` or `off` followed by pair numbers,
toggles specific pairs. With a preset name, sets the mask to a predefined
value.

Presets:

- **all** – all pairs
- **headphones** – pair 1 only (typical headphone setup)
- **none** – disable all pairs

```
dspictl crossfeed outputs              # show active pairs
dspictl crossfeed outputs on 0 1       # enable pairs 0 and 1
dspictl crossfeed outputs off 2        # disable pair 2
dspictl crossfeed outputs all          # enable all
dspictl crossfeed outputs headphones   # pair 1 only
dspictl crossfeed outputs none         # disable all
```

## dac-mute

DAC hardware mute control. Manages the mute GPIO pin and timing for pop-free
output muting on power state transitions.

With no arguments, shows the current DAC mute configuration.

### dac-mute on

Enable DAC hardware mute.

```
dspictl dac-mute on
```

### dac-mute off

Disable DAC hardware mute.

```
dspictl dac-mute off
```

### dac-mute config

Configure all DAC hardware mute GPIO parameters at once using flags.
All flags are required:

- *--enabled*: on/off (or true/false)
- *--active-low*: on/off (true=active low, false=active high)
- *--pin*: GPIO pin number, or 255 to keep current
- *--hold-ms*: hold time in milliseconds
- *--release-ms*: release time in milliseconds

```
dspictl dac-mute config --enabled --active-low --pin 255 --hold-ms 100 --release-ms 50
```

### dac-mute test

Run a DAC mute test cycle to verify the mute GPIO configuration.

```
dspictl dac-mute test
```

## leveller

Dynamic range compression (leveller) control for automatic level management.

With no subcommand, shows the leveller status with all parameters.

### leveller enable [on|off]

Get or set leveller enable state.

```
dspictl leveller enable on
```

### leveller amount [*value*]

Get or set compression amount.

```
dspictl leveller amount 10
```

### leveller speed [*n*]

Get or set attack/release speed.

```
dspictl leveller speed 3
```

### leveller maxgain [*db*]

Get or set maximum gain reduction in dB.

```
dspictl leveller maxgain -12
```

### leveller lookahead [on|off]

Get or set lookahead enable.

```
dspictl leveller lookahead on
```

### leveller gate [*db*]

Get or set noise gate threshold in dB.

```
dspictl leveller gate -80
```

### leveller detector-mask [on|off] [<channels...>]

Get or set which input channels feed the shared RMS detector (V18+). With
no arguments, shows the current active inputs. With `on` or `off` followed
by channel numbers, toggles specific inputs. With a preset name, sets the
mask to a predefined value.

Presets:

- **all** / **night** – all inputs (Night mode)
- **center** – center channel only (Dialog boost)
- **front-lr** – front L/R only
- **none** – disable all

```
dspictl leveller detector-mask              # show active inputs
dspictl leveller detector-mask on 0 1       # enable inputs 0 and 1
dspictl leveller detector-mask off 2        # disable input 2
dspictl leveller detector-mask all          # all inputs
dspictl leveller detector-mask center       # center only
dspictl leveller detector-mask front-lr     # front L/R only
dspictl leveller detector-mask none         # disable all
```

### leveller apply-mask [on|off] [<channels...>]

Get or set which input channels receive the computed gain (V18+). Same
structure and presets as detector-mask.

Presets:

- **all** / **night** – all inputs
- **center** – center channel only
- **front-lr** – front L/R only
- **none** – disable all

```
dspictl leveller apply-mask              # show active inputs
dspictl leveller apply-mask on 0 1       # enable inputs 0 and 1
dspictl leveller apply-mask all          # all inputs
```

## lg-sound-sync

LG Sound Sync control for TV audio return over optical. When enabled, the LG
TV remote can control the DSPi volume via the S/PDIF connection.

### lg-sound-sync enable [on|off]

Get or set LG Sound Sync enable state.

```
dspictl lg-sound-sync enable on
```

### lg-sound-sync status

Show detailed LG Sound Sync status including whether a compatible TV is
present, mute state, and volume level.

```
dspictl lg-sound-sync status
```

## siggen

Control the onboard test signal generator. The generator can inject one of 15
measurement/diagnostic signals directly into the output pipeline without a host
audio stream.

### siggen types

List the signal types supported by the connected device, including parameter
ranges.

```
dspictl siggen types
```

### siggen status

Show the current generator state.

```
dspictl siggen status
```

### siggen start

Configure and start the generator. At minimum, `--type` and `--channels` are
required. The default level is -20 dBFS.

```
dspictl siggen start --type sine --channels 0,1 --level -20 --freq 1000
dspictl siggen start --type sweep-log --channels 0 --f1 20 --f2 20000 --duration 10000
dspictl siggen start --type channel-id --channels 0 1 2 3
```

Use `--raw` to bypass per-channel crossover + PEQ on the generator channels,
`--walk` to play one channel at a time, and `--decorr` for independent noise
generators on each channel.

### siggen config

Stage a configuration without starting the generator.

```
dspictl siggen config --type pink --channels 0 --level -30
```

### siggen stop

Stop the generator. Use `--now` for an immediate hard stop without fade.

```
dspictl siggen stop
dspictl siggen stop --now
```

## adat

ADAT bulk output control (RP2350 only). Streams all 8 post-gain output
channels as one ADAT lightpipe signal on a single GPIO.

### adat enable [on|off]

Get or set the ADAT bulk output enable state.

```
dspictl adat enable        # show current state
dspictl adat enable on     # enable ADAT output
```

### adat pin [*gpio*]

Get or set the ADAT output GPIO pin. The platform default is GPIO 12; pass 255
to restore it. Re-routing is allowed even while enabled (the stream moves
under mute).

```
dspictl adat pin       # show current pin
dspictl adat pin 12    # set ADAT output to GPIO 12
```

### adat status

Show the ADAT output stream status: configured enable, stream activity, pin,
rate_ok, and resync/slip counters.

```
dspictl adat status
```

## upmix

Stereo upmixer control (RP2350 only). Derives Centre + Left/Right Surround
virtual source channels from the stereo input pair, routeable via the matrix.

### upmix status

Show live upmixer telemetry: active/parked state and the smoothed
correlation, balance, and steering gains.

```
dspictl upmix status
```

### upmix config

Show the full upmixer configuration: centre/surround modes, presence,
strength, centre width, threshold, attack/release, detector HPF, surround
delay/HPF/LPF, and decorrelation.

```
dspictl upmix config
```

### upmix on / upmix off

Enable or disable the upmixer.

```
dspictl upmix on
dspictl upmix off
```

### upmix set *param* *value*

Set a single upmixer parameter. Parameters: `enabled`, `center-mode`
(0=passive, 1=adaptive, 2=off), `surround-mode` (0=off, 1=passive, 2=adaptive),
`strength`, `center-width`, `threshold`, `attack`, `release`, `det-hpf`,
`surround-delay`, `surround-hpf`, `surround-lpf`, `decorr`, `presence`.

```
dspictl upmix set presence -4
dspictl upmix set center-mode 2
```

## ctrl

External control interfaces: a UART and an I2C target that expose the same
vendor command surface over wires (see `control_interfaces_spec.md`). Both
ship disabled; the SET commands are USB-only (the firmware refuses them over
UART/I2C so a controller can never lock itself out).

### ctrl uart [on|off]

Get or set the UART control interface. Flags: `--tx <gpio>`, `--rx <gpio>`,
`--baud <rate>` (9600..1000000), `--notify`. The apply outcome is read back
via `ctrl status`.

```
dspictl ctrl uart
dspictl ctrl uart on --tx 16 --rx 17 --baud 115200
```

### ctrl i2c [on|off]

Get or set the I2C target control interface. Flags: `--sda <gpio>`, `--scl
<gpio>`, `--address <0x08..0x77>`.

```
dspictl ctrl i2c
dspictl ctrl i2c on --sda 18 --scl 19 --address 0x42
```

### ctrl status

Show the live UART/I2C interface status and the last apply results.

```
dspictl ctrl status
```

## cs

Control Surfaces: user-wired physical controls and indicators (buttons,
switches, pots, rotary encoders, LEDs, IR receivers) on spare GPIOs, each
bound to one firmware parameter (see `control_surfaces_spec.md`). SETs are
apply-live-only previews; `cs save` persists them and `cs revert` discards
them. Outcomes are polled via `cs status`.

### cs status

Show the status packet: last SET result, dirty flag, active bindings, IR
learn state.

```
dspictl cs status
```

### cs caps

Show the capability header (format version, counts) and the per-type action
masks.

```
dspictl cs caps
```

### cs binding get *slot*

Show the live binding of a slot (0-15).

```
dspictl cs binding get 3
```

### cs binding set *slot*

Upload a binding. Flags: `--type <none|button|switch|pot|encoder|led|led-pwm|ir>`,
`--noun <noun>` (e.g. `user-volume`, `master-volume`, `filter-freq`, `preset`),
`--action <adjust|step|inc|dec|toggle|set|follow|trigger|ind-equals|momentary|ind-above|ind-level>`,
`--gpio <pin[,pin]>`, `--event <press|long|double>`, `--target <ch>`,
`--index <band>`, `--value`, `--step`, `--range-min`, `--range-max`,
`--invert`.

```
dspictl cs binding set 0 --type button --noun user-mute --action toggle --gpio 26
dspictl cs binding set 1 --type pot --noun master-volume --action adjust --gpio 27
```

### cs name get *slot* / cs name set *slot* *name*

Get or set a slot's user label (up to 31 bytes). Names survive binding
changes and slot clears.

```
dspictl cs name set 0 "Mute All"
dspictl cs name get 0
```

### cs ir get *subslot* / cs ir set *subslot*

Get or set an IR remote command (sub-slots 0-15). The `ir set` flags mirror
`binding set` plus `--protocol <nec|rc5|rc6|hash>` and `--code <hex>` (the
learned code).

```
dspictl cs ir set 0 --noun preset --action set --protocol nec --code 0x00FF10EF --value 1
dspictl cs ir get 0
```

### cs ir learn [arm|cancel|read]

*arm* listens for the next decoded press and captures it; *read* returns the
captured protocol + code; *cancel* aborts. Requires a live CS_TYPE_IR binding.

```
dspictl cs ir learn arm
dspictl cs ir learn read
```

### cs save / cs revert

Persist the whole live CS config (bindings, IR commands, slot names) to flash,
or discard the preview and re-apply the stored config. Both are deferred;
poll `cs status` for the outcome.

```
dspictl cs save
dspictl cs revert
```

## mixer

Launch an interactive full-screen terminal UI mixer.

```
dspictl mixer
```

The mixer provides a visual interface for controlling volume levels, mute
status, and other real-time parameters. Use arrow keys and mouse to interact;
press *?* or *q* for help and quit.

# FILES

*dspictl* communicates with DSPi devices via USB control transfers. The devices
appear as USB audio class 2.0 interfaces and require no kernel driver on modern
Linux or macOS systems.

On Linux, a udev rule may be needed for regular-user USB access (see
TROUBLESHOOTING below).

# ENVIRONMENT

No environment variables are required. Log output is written to stderr as JSON
and can be controlled via the usual `slog` environment conventions.

# TROUBLESHOOTING

### libusb: bad access [code -3]

This error indicates the kernel is blocking regular-user USB access. Create a
udev rule:

    echo 'SUBSYSTEM=="usb", ATTR{idVendor}=="2e8b", ATTR{idProduct}=="feaa", MODE="0666"' \
      | sudo tee /etc/udev/rules.d/99-dspi.rules
    sudo udevadm control --reload-rules
    sudo udevadm trigger

### No devices found

Ensure the DSPi is connected via USB, powered on, and that you have permission
to access the device (see udev rule above). Use `dspictl status` to confirm
the device is detected.

### libusb: pipe error [code -9] / firmware predates the V16 wire protocol

dspictl 2.x requires the V16+ wire protocol, shipped since firmware
v1.1.5-beta3.  On older firmware every command fails with a clear
"predates the V16 wire protocol" error (previously a cryptic STALL / pipe
error).  `dspictl firmware version` and `dspictl status` report the device's
version and the incompatibility; `dspictl firmware upgrade` still works on
such devices — the bootloader command is the one operation allowed through.

# BUGS

Report bugs and feature requests at:

    https://github.com/suhlig/dspi/issues

# SEE ALSO

**DSPi USB Control Protocol**: https://github.com/WeebLabs/DSPi#usb-control-protocol

**DSPi Project Homepage**: https://github.com/WeebLabs/DSPi
