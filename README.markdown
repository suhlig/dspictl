# DSPi-Go

A command-line tool and Go library to control a [DSPi audio device](https://github.com/WeebLabs/DSPi) from the command line.

Pre-built binaries are available from [GitHub Releases](https://github.com/suhlig/dspi/releases).

Choose the binary that matches your platform:

| File | Architecture | Typical Hardware |
|---|---|---|
| `dspictl-linux-amd64` | x86_64 | Desktop PCs, laptops |
| `dspictl-linux-arm64` | ARM 64-bit | Raspberry Pi 3/4/5 (64-bit OS) |
| `dspictl-linux-armv7` | ARM 32-bit | Raspberry Pi 2/3/4/5 (32-bit OS) |
| `dspictl-darwin-amd64` | x86_64 | Intel Macs |
| `dspictl-darwin-arm64` | ARM 64-bit | Apple Silicon Macs |

After downloading, make it executable and run it:

```sh
chmod +x dspictl-*
./dspictl-*
```

> **Note:** You may need a [udev rule](#troubleshooting) for USB access as a regular user on Linux.

The macOS version needs to be allowed to be executed with `xattr -d com.apple.quarantine dspictl-darwin-arm64` once because it is not a signed binary.

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

## Local Compiling for Linux

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

## References

* The USB control protocol is documented at the [DSPi repository](https://github.com/WeebLabs/DSPi#usb-control-protocol)
