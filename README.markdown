# dspictl

A command-line tool and Go library to control one or more [DSPi audio devices](https://github.com/WeebLabs/DSPi) from the command line. A full-screen terminal UI is also available as `dspictl mixer`.

Homebrew users can install it via `brew install suhlig/tap/dspictl`. For other platforms, pre-built archives are available from [GitHub Releases](https://github.com/suhlig/dspi/releases):

| File | Architecture | Typical Hardware |
|---|---|---|
| `dspictl-linux-amd64.tar.gz` | x86_64 | Desktop PCs, laptops |
| `dspictl-linux-arm64.tar.gz` | ARM 64-bit | Raspberry Pi 3/4/5 (64-bit OS) |
| `dspictl-linux-armv7.tar.gz` | ARM 32-bit | Raspberry Pi 2/3/4/5 (32-bit OS) |
| `dspictl-darwin-amd64.tar.gz` | x86_64 | Intel Macs |
| `dspictl-darwin-arm64.tar.gz` | ARM 64-bit | Apple Silicon Macs |

### Linux

```sh
arch=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/armv7l/armv7/') && curl -sL "https://github.com/suhlig/dspictl/releases/latest/download/dspictl-linux-${arch}.tar.gz" | tar xz && sudo mv dspictl /usr/local/bin
```

> **Note:** You may need a [udev rule](#troubleshooting) for USB access as a regular user on Linux.

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
watchexec --restart --wrap-process=none go run ./cmd/dspictl mixer
```

## Man Page

The man page is written in Markdown (`man/dspictl.md`), embedded into the binary at build time, and converted to troff format at runtime by `dspictl man [dir]`.

To preview the embedded man page (proving the content the user gets):

```sh
mkdir -p /tmp/dspictl-man/man1
go run ./cmd/dspictl man /tmp/dspictl-man/man1
MANPATH=/tmp/dspictl-man man dspictl
```

## Tests

The library includes two test suites:

1. **Unit tests** — mock-based, run anywhere, exercise request encoding and response decoding.
2. **Hardware tests** — require a real DSPi device connected via USB. They are non-destructive: each test captures the full device state at the start and restores it afterwards.

| Scenario | Command |
|---|---|
| Unit tests only (CI, no hardware) | `go tool ginkgo ./...` |
| Unit + hardware tests locally | `go tool ginkgo -tags=hwtest ./...` |
| Target only RP2350 | `DSPI_TEST_PLATFORM=RP2350 go tool ginkgo -tags=hwtest ./...` |
| Target specific serial | `DSPI_TEST_SERIAL=ABC123 go tool ginkgo -tags=hwtest ./...` |
| Start from factory reset | `DSPI_FACTORY_RESET=1 go tool ginkgo -tags=hwtest ./...` |

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

## Troubleshooting

### libusb: bad access [code -3]

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

### Package libusb-1.0 was not found in the pkg-config search path

```
Package libusb-1.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `libusb-1.0.pc' to the PKG_CONFIG_PATH environment variable
Package 'libusb-1.0', required by 'virtual:world', not found
```

You are missing `libusb-1.0-0-dev`. Install it with:

```sh
sudo apt-get install -y libusb-1.0-0-dev
```

### mpg123 not starting

```
[src/libout123/modules/alsa.c:open_alsa():181] error: cannot open device default
```

You have no or the wrong card set as default device in your ALSA configuration. Edit `/etc/asound.conf` or `~/.asoundrc` to set the correct card as default, e.g.:

```
defaults.pcm.card 2
defaults.pcm.device 0
```

### GitHub Actions

* Show the most recently failed GitHub Actions run:

  ```sh
  $ gh run view --log $(gh run list --workflow=ci.yml --status failure --json databaseId --jq '.[].databaseId')
  ```

## Hardware Tests on Self-Hosted Runners

The repository includes a `.github/workflows/hardware.yml` workflow that runs the full hardware test suite on a self-hosted runner. It is triggered manually via **Actions → Hardware Tests → Run workflow**.

To set up a Raspberry Pi (or any Linux machine) as a self-hosted runner:

1. Install the runner from your repository's **Settings → Actions → Runners → New self-hosted runner** page.
2. Install system dependencies:
   ```sh
   sudo apt-get update
   sudo apt-get install -y libusb-1.0-0-dev
   ```
3. Install Go (see [Requirements](#requirements)).
4. Add a [udev rule](#libusb-bad-access-code--3) so the runner user can access DSPi devices.
5. Connect the DSPi devices and start the runner service.

The workflow runs `go test -tags=hwtest -v ./...` against whatever devices are connected, reporting firmware versions and platform info in the logs.

## Releasing

> Requires [git-cliff](https://git-cliff.org) (e.g. `brew install git-cliff`)

1. Generate a starting point for the next version:

   ```sh
   git cliff --unreleased --bump --prepend CHANGELOG.md
   ```

1. Edit and then commit the updated `CHANGELOG.md`:

   ```sh
   git add CHANGELOG.md && git commit -m "Prepare changelog for $(git cliff --bumped-version)"
   ```

1. Push a tag to trigger the release workflow:

   ```sh
   version=$(git cliff --bumped-version)
   git tag "$version" && git push origin "$version"
   ```

## License

MIT

## References

* The USB control protocol is documented at the [DSPi repository](https://github.com/WeebLabs/DSPi#usb-control-protocol)
