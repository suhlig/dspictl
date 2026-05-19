# DSPi-Go

A Go library and example application for reading real-time telemetry from a [DSPi audio device](https://github.com/WeebLabs/DSPi) over USB.

## What It Does

DSPi-Go talks to a DSPi (a Raspberry Pi Pico-based digital audio platform) through its vendor-specific USB control interface. The library exposes per-channel peak levels, CPU load, and clip flags so you can build meters, monitors, or automation tools in Go.

The USB control protocol is documented at the [DSPi repository](https://github.com/WeebLabs/DSPi#usb-control-protocol).

## Example Application

`examples/mixer` is a full-screen terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss). It renders a live meter with colored bars, CPU usage, and clip indicators.

## Development

This project uses [pre-commit](https://pre-commit.com/) to check formatting, linting, and compilation before each commit. After cloning, install the hook:

```sh
brew install pre-commit && pre-commit install
```

Use [watchexec](https://github.com/watchexec/watchexec) to auto-restart the TUI on source changes:

```sh
brew install watchexec
```

Then run it with

```sh
watchexec --restart --wrap-process=none go run ./examples/mixer
```

## Requirements

- Go 1.26 or later
- [libusb-1.0](https://libusb.info/) installed on your system (required by `github.com/google/gousb`)
- A DSPi device connected via USB

## Compiling for Linux

As we have C code dependencies, the easiest is to compile for the target platform in a container:

```sh
docker build -t dspi-builder -f Dockerfile.linux .
```

This will give us a container image that we can use to compile the `dspictl` tool for Linux:

```sh
docker run --rm -v "$PWD":/src -e CGO_ENABLED=1 dspi-builder
```

We can then copy the compiled `dspictl` binary to the target device (e.g. `pi5`) and run it as a regular user:

```sh
scp dspictl pi5:bin && ssh pi5 dspictl
```

### Troubleshooting

You'll likely see the following error message when running as regular user on Linux:

```
Error: opening DSPi devices: enumerating DSPi devices: libusb: bad access [code -3]
```

The reason is that the kernel blocks regular users from raw USB access. You need a **udev rule**.

1. Create a rule file:

    ```sh
    echo 'SUBSYSTEM=="usb", ATTR{idVendor}=="2e8b", ATTR{idProduct}=="feaa", MODE="0666"' \
      | sudo tee /etc/udev/rules.d/99-dspi.rules
    ```

1. Apply it:

    ```sh
    sudo udevadm control --reload-rules
    sudo udevadm trigger
    ```

3. Unplug and reconnect the DSPi, and `dspictl` will work as a regular user.

A more security-conscious version uses a group instead of world-writable:

```sh
sudo groupadd -f dspi
sudo usermod -aG dspi suhlig # add user suhlig to the dspi group
echo 'SUBSYSTEM=="usb", ATTR{idVendor}=="2e8b", ATTR{idProduct}=="feaa", GROUP="dspi", MODE="0660"' \
  | sudo tee /etc/udev/rules.d/99-dspi.rules
```

## GitHub Actions

* Show the most recently failed GitHub Actions run:

  ```sh
  $ gh run view --log $(gh run list --workflow=ci.yml --status failure --json databaseId --jq '.[].databaseId')
  ```

## License

MIT
