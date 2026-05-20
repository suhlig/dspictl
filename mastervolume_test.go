package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("MasterVolume", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetMasterVolume), 0, 0}:     {0x00, 0x00, 0x20, 0x41}, // float32 10.0
				{uint16(dspi.ReqGetMasterVolumeMode), 0, 0}: {0x01},
				{uint16(dspi.ReqSaveMasterVolume), 0, 0}:    {},
				{uint16(dspi.ReqSetMasterVolume), 0, 0}:     {},
				{uint16(dspi.ReqSetMasterVolumeMode), 0, 0}: {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("GetMasterVolume", func() {
		It("decodes the float32 LE response", func() {
			gain, err := dev.GetMasterVolume()
			Expect(err).ToNot(HaveOccurred())
			Expect(gain.DB()).To(BeNumerically("~", 10.0, 0.001))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetMasterVolume()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetMasterVolume)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetMasterVolume()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetMasterVolume", func() {
		It("encodes -6 dB as little-endian float32", func() {
			err := dev.SetMasterVolume(dspi.NewGain(-6.0))
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00, 0x00, 0xc0, 0xc0}))
		})

		It("sends the correct bRequest", func() {
			_ = dev.SetMasterVolume(dspi.NewGain(0.0))
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetMasterVolume)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetMasterVolume(dspi.NewGain(0.0))
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetMasterVolumeMode", func() {
		It("returns the mode byte", func() {
			mode, err := dev.GetMasterVolumeMode()
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(1))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetMasterVolumeMode()
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetMasterVolumeMode)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetMasterVolumeMode()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetMasterVolumeMode", func() {
		It("sends the mode byte", func() {
			err := dev.SetMasterVolumeMode(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("sends the correct bRequest", func() {
			_ = dev.SetMasterVolumeMode(0)
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetMasterVolumeMode)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetMasterVolumeMode(0)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SaveMasterVolume", func() {
		It("sends the correct bRequest", func() {
			err := dev.SaveMasterVolume()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSaveMasterVolume)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SaveMasterVolume()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
