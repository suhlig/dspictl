//go:build hwtest

package dspi_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

// hardwareDevice wraps an open DSPi device with its captured state snapshot.
type hardwareDevice struct {
	*dspi.Device
	snapshot *dspi.BulkParams
}

// discoverHardwareDevices enumerates connected DSPi devices and applies
// optional DSPI_TEST_SERIAL and DSPI_TEST_PLATFORM filters.
func discoverHardwareDevices() ([]*hardwareDevice, error) {
	infos, err := dspi.List()
	if err != nil {
		return nil, fmt.Errorf("listing devices: %w", err)
	}

	if len(infos) == 0 {
		return nil, fmt.Errorf("no DSPi devices found")
	}

	filterSerial := os.Getenv("DSPI_TEST_SERIAL")
	filterPlatform := os.Getenv("DSPI_TEST_PLATFORM")

	var devices []*hardwareDevice

	for _, info := range infos {
		if filterSerial != "" && info.Serial != filterSerial {
			continue
		}

		dev, err := dspi.Open(info)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", info.Serial, err)
		}

		if filterPlatform != "" && dev.Platform().String() != filterPlatform {
			dev.Close()
			continue
		}

		devices = append(devices, &hardwareDevice{Device: dev})
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no DSPi devices matched filters")
	}

	return devices, nil
}

var _ = Describe("Hardware", Ordered, func() {
	var devices []*hardwareDevice

	BeforeAll(func() {
		var err error
		devices, err = discoverHardwareDevices()
		Expect(err).ToNot(HaveOccurred())
		Expect(devices).ToNot(BeEmpty())

		for _, d := range devices {
			snap, err := d.GetAllParams()
			Expect(err).ToNot(HaveOccurred(), "capturing state for %s", d.Serial())
			d.snapshot = snap

			By(fmt.Sprintf(
				"Device %s (%s) firmware %d.%d — %d channels, %d outputs, payload %d bytes",
				d.Serial(),
				d.Platform(),
				snap.Header.FWMajor,
				snap.Header.FWMinor,
				snap.Header.NumChannels,
				snap.Header.NumOutputs,
				snap.Header.PayloadLength,
			))
		}

		if os.Getenv("DSPI_FACTORY_RESET") == "1" {
			for _, d := range devices {
				By(fmt.Sprintf("Factory resetting %s", d.Serial()))
				Expect(d.FactoryReset()).To(Succeed())
			}
		}
	})

	AfterAll(func() {
		for _, d := range devices {
			if d.snapshot != nil {
				By(fmt.Sprintf("Restoring state on %s", d.Serial()))
				_ = d.SetAllParams(d.snapshot)
			}
			d.Close()
		}
	})

	Describe("Identity", func() {
		It("has a valid platform", func() {
			for _, d := range devices {
				Expect(d.Platform()).To(Or(Equal(dspi.PlatformRP2040), Equal(dspi.PlatformRP2350)))
			}
		})

		It("has a non-empty serial", func() {
			for _, d := range devices {
				Expect(d.Serial()).ToNot(BeEmpty())
			}
		})

		It("has valid bus and address", func() {
			for _, d := range devices {
				Expect(d.Bus()).To(BeNumerically(">=", 0))
				Expect(d.Address()).To(BeNumerically(">=", 0))
			}
		})
	})

	Describe("ReadMeter", func() {
		It("returns channel count matching platform", func() {
			for _, d := range devices {
				snap := d.ReadMeter()
				Expect(snap.Err()).ToNot(HaveOccurred())

				// V16+ unified channel model: inputs + outputs
				// (RP2040: 2 + 5 = 7, RP2350: 8 + 9 = 17).
				expected := 7
				if d.Platform() == dspi.PlatformRP2350 {
					expected = 17
				}
				Expect(snap.Channels).To(Equal(expected))
			}
		})
	})

	Describe("Channels", func() {
		It("returns consistent channel names", func() {
			for _, d := range devices {
				chans, err := d.Channels()
				Expect(err).ToNot(HaveOccurred())

				// V16+ unified channel model: inputs + outputs
				// (RP2040: 2 + 5 = 7, RP2350: 8 + 9 = 17).
				expectedCount := 7
				if d.Platform() == dspi.PlatformRP2350 {
					expectedCount = 17
				}
				Expect(chans).To(HaveLen(expectedCount))

				for i, ch := range chans {
					Expect(ch.Index).To(Equal(i))
					Expect(ch.Name).ToNot(BeEmpty())
				}
			}
		})
	})

	Describe("MasterVolume", func() {
		It("reads a valid value", func() {
			for _, d := range devices {
				vol, err := d.GetMasterVolume()
				Expect(err).ToNot(HaveOccurred())
				Expect(vol.DB()).To(BeNumerically(">=", -128))
				Expect(vol.DB()).To(BeNumerically("<=", 0))
			}
		})
	})

	Describe("MasterVolumeMode", func() {
		It("reads a valid mode", func() {
			for _, d := range devices {
				mode, err := d.GetMasterVolumeMode()
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(Or(Equal(0), Equal(1)))
			}
		})
	})

	Describe("ActivePreset", func() {
		It("reads a valid preset slot", func() {
			for _, d := range devices {
				slot, err := d.GetActivePreset()
				Expect(err).ToNot(HaveOccurred())
				Expect(slot).To(BeNumerically(">=", 0))
				Expect(slot).To(BeNumerically("<", 8))
			}
		})
	})

	Describe("PresetDirectory", func() {
		It("returns valid metadata", func() {
			for _, d := range devices {
				dir, err := d.GetPresetDirectory()
				Expect(err).ToNot(HaveOccurred())
				Expect(dir.StartupMode).To(Or(Equal(0), Equal(1)))
				Expect(dir.DefaultSlot).To(BeNumerically(">=", 0))
				Expect(dir.DefaultSlot).To(BeNumerically("<", 8))
				Expect(dir.LastActive).To(BeNumerically(">=", 0))
				Expect(dir.LastActive).To(BeNumerically("<", 8))
			}
		})
	})

	Describe("Core1", func() {
		It("reads a valid mode", func() {
			for _, d := range devices {
				mode, err := d.GetCore1Mode()
				Expect(err).ToNot(HaveOccurred())
				Expect(mode).To(BeNumerically(">=", 0))
				Expect(mode).To(BeNumerically("<=", 2))
			}
		})

		It("reads conflict status without error", func() {
			for _, d := range devices {
				_, err := d.GetCore1Conflict()
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})

	Describe("BufferStats", func() {
		It("returns non-empty data", func() {
			for _, d := range devices {
				stats, err := d.GetBufferStats()
				Expect(err).ToNot(HaveOccurred())
				Expect(stats.Data).ToNot(BeEmpty())
			}
		})
	})

	Describe("USBErrorStats", func() {
		It("returns counters without error", func() {
			for _, d := range devices {
				stats, err := d.GetUSBErrorStats()
				Expect(err).ToNot(HaveOccurred())
				Expect(stats).ToNot(BeNil())
			}
		})
	})

	Describe("GetAllParams", func() {
		It("has a valid header", func() {
			for _, d := range devices {
				snap, err := d.GetAllParams()
				Expect(err).ToNot(HaveOccurred())
				Expect(snap.Header.FormatVersion).To(BeNumerically(">", 0))
				Expect(snap.Header.Platform).To(Equal(d.Platform()))
				Expect(snap.Header.NumChannels).To(BeNumerically(">", 0))
				Expect(snap.Header.NumOutputs).To(BeNumerically(">", 0))
				Expect(snap.Header.PayloadLength).To(BeNumerically(">=", 16))
				Expect(snap.Header.FWMajor).To(BeNumerically(">", 0))
				Expect(snap.Header.FWMinor).To(BeNumerically(">", 0))
				Expect(len(snap.Raw)).To(BeNumerically("==", snap.Header.PayloadLength))
			}
		})
	})
})
