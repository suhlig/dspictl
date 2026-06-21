//go:build hwtest

package dspi_test

import (
	"errors"
	"fmt"

	"github.com/google/gousb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Hardware Round-trip", Ordered, func() {
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
				"Device %s (%s) firmware %d.%d",
				d.Serial(),
				d.Platform(),
				snap.Header.FWMajor,
				snap.Header.FWMinor,
			))
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

	restoreSnapshot := func() {
		for _, d := range devices {
			if d.snapshot != nil {
				Expect(d.SetAllParams(d.snapshot)).To(Succeed(),
					"restoring state on %s", d.Serial())
			}
		}
	}

	Describe("MasterVolume", func() {
		It("sets and reads back -10 dB", func() {
			for _, d := range devices {
				Expect(d.SetMasterVolume(dspi.NewGain(-10))).To(Succeed())
				vol, err := d.GetMasterVolume()
				Expect(err).ToNot(HaveOccurred())
				Expect(vol.DB()).To(BeNumerically("~", -10, 0.5))
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("PreampChannel", func() {
		It("sets and reads back -6 dB on channel 0", func() {
			for _, d := range devices {
				Expect(d.SetPreampChannel(0, dspi.NewGain(-6))).To(Succeed())
				gain, err := d.GetPreampChannel(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(gain.DB()).To(BeNumerically("~", -6, 0.5))
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("OutputGain", func() {
		It("sets and reads back on output 0", func() {
			for _, d := range devices {
				Expect(d.SetOutputGain(0, dspi.NewGain(-12))).To(Succeed())
				gain, err := d.GetOutputGain(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(gain.DB()).To(BeNumerically("~", -12, 0.5))
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("OutputMute", func() {
		It("toggles mute on output 0", func() {
			for _, d := range devices {
				Expect(d.SetOutputMute(0, true)).To(Succeed())
				muted, err := d.GetOutputMute(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(muted).To(BeTrue())
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("OutputEnable", func() {
		It("toggles enable on output 0", func() {
			for _, d := range devices {
				Expect(d.SetOutputEnable(0, false)).To(Succeed())
				enabled, err := d.GetOutputEnable(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(enabled).To(BeFalse())
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("OutputDelay", func() {
		It("sets and reads back on output 0", func() {
			for _, d := range devices {
				Expect(d.SetOutputDelay(0, 5.0)).To(Succeed())
				delay, err := d.GetOutputDelay(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(delay).To(BeNumerically("~", 5.0, 0.1))
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("ChannelName", func() {
		It("sets and reads back on channel 0", func() {
			for _, d := range devices {
				Expect(d.SetChannelName(0, "HWTest")).To(Succeed())
				name, err := d.ChannelName(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(name).To(Equal("HWTest"))
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("MatrixRoute", func() {
		It("toggles enabled and changes gain on route (0,0)", func() {
			for _, d := range devices {
				Expect(d.SetMatrixRoute(&dspi.MatrixRoute{
					Input:   0,
					Output:  0,
					Enabled: true,
					Gain:    dspi.NewGain(3),
				})).To(Succeed())

				route, err := d.GetMatrixRoute(0, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(route.Enabled).To(BeTrue())
				Expect(route.Gain.DB()).To(BeNumerically("~", 3, 0.5))
			}
		})
		AfterEach(restoreSnapshot)
	})

	Describe("CrossoverBand", func() {
		It("sets and reads back on output channel 2, band 20", func() {
			for _, d := range devices {
				Expect(d.SetCrossoverBand(&dspi.CrossoverBand{
					Channel: 2,
					Band:    20,
					Type:    dspi.CrossoverTypeLR4LP,
					Freq:    800,
					Bypass:  false,
				})).To(Succeed())

				b, err := d.GetCrossoverBand(2, 20)
				if err != nil {
					var usbErr gousb.Error
					if errors.As(err, &usbErr) && usbErr == gousb.ErrorPipe {
						By(fmt.Sprintf(
							"Skipping crossover read-back on %s (firmware does not support crossover)",
							d.Serial(),
						))
						continue
					}
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Type).To(Equal(dspi.CrossoverTypeLR4LP))
				Expect(b.Freq).To(BeNumerically("~", 800, 1))
				Expect(b.Bypass).To(BeFalse())
			}
		})
		AfterEach(restoreSnapshot)
	})
})
