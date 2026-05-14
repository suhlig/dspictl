# DSPi-Go

A Go library and example application for reading real-time telemetry from a [DSPi audio device](https://github.com/WeebLabs/DSPi) over USB.

## What It Does

DSPi-Go talks to a DSPi (a Raspberry Pi Pico-based digital audio platform) through its vendor-specific USB control interface. The library exposes per-channel peak levels, CPU load, and clip flags so you can build meters, monitors, or automation tools in Go.

The USB control protocol is documented at the [DSPi repository](https://github.com/WeebLabs/DSPi#usb-control-protocol).

## Example Application

`examples/mixer` is a full-screen terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss). It renders a live meter with colored bars, CPU usage, and clip indicators.

## Development

This project uses [pre-commit](https://pre-commit.com/) to check formatting, linting, and compilation before each commit. After cloning, install the hook:

```command
brew install pre-commit && pre-commit install
```

Use [watchexec](https://github.com/watchexec/watchexec) to auto-restart the TUI on source changes:

```command
brew install watchexec
```

Then run it with

```command
watchexec --restart go run ./examples/mixer
```

## Requirements

- Go 1.26 or later
- [libusb-1.0](https://libusb.info/) installed on your system (required by `github.com/google/gousb`)
- A DSPi device connected via USB

## License

MIT
