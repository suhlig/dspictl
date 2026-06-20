# ALSA Multi-Channel Output

This directory contains an ALSA configuration that routes a single 4-channel PCM stream across two separate stereo USB sound cards.

This is useful when the host computer needs to feed more channels than a single USB audio interface can provide, and the DSPi firmware does not support merging multiple USB inputs.

## How It Works

The `asound.conf` defines a `multi-out` PCM that:

1. Creates a 4-channel logical slave from two 2-channel hardware devices.
2. Maps the logical channels to physical devices via a routing table.
3. Sets the result as the system default so applications see a single 4-channel output.

## Configuration

Edit the `asound.conf` to match your system's card numbers (use `aplay -l` to list them):

```
pcm "hw:2,0"   # first stereo device
pcm "hw:3,0"   # second stereo device
```

Install the file:

```sh
sudo cp asound.conf /etc/asound.conf
```

## Verification

```sh
aplay -L | grep multi-out
speaker-test -c 4 -t sine
```

Each of the four channels should emit a test tone in sequence.
