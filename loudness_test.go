package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Loudness", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				// GetLoudness: enabled
				{uint16(dspi.ReqGetLoudness), 0, 0}: {1},

				// SetLoudness: empty success
				{uint16(dspi.ReqSetLoudness), 0, 0}: {},

				// GetLoudnessReference: 83.0 dB SPL
				{uint16(dspi.ReqGetLoudnessReference), 0, 0}: {0x00, 0x00, 0xA6, 0x42}, // float32 83.0

				// SetLoudnessReference: empty success
				{uint16(dspi.ReqSetLoudnessReference), 0, 0}: {},

				// GetLoudnessIntensity: 100.0%
				{uint16(dspi.ReqGetLoudnessIntensity), 0, 0}: {0x00, 0x00, 0xC8, 0x42}, // float32 100.0

				// SetLoudnessIntensity: empty success
				{uint16(dspi.ReqSetLoudnessIntensity), 0, 0}: {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("GetLoudness", func() {
		It("returns true when enabled", func() {
			enabled, err := dev.GetLoudness()
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())
		})

		It("returns false when disabled", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetLoudness), 0, 0}] = []byte{0}
			enabled, err := dev.GetLoudness()
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeFalse())
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetLoudness()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetLoudness)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetLoudness()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetLoudness", func() {
		It("sends 1 when enabling", func() {
			err := dev.SetLoudness(true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{1}))
		})

		It("sends 0 when disabling", func() {
			err := dev.SetLoudness(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0}))
		})

		It("sends the correct bRequest", func() {
			_ = dev.SetLoudness(true)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetLoudness)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetLoudness(true)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetLoudnessReference", func() {
		It("decodes the float32 LE response", func() {
			spl, err := dev.GetLoudnessReference()
			Expect(err).ToNot(HaveOccurred())
			Expect(spl).To(BeNumerically("~", 83.0, 0.01))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetLoudnessReference()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetLoudnessReference)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetLoudnessReference()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetLoudnessReference", func() {
		It("encodes 75 dB as little-endian float32", func() {
			err := dev.SetLoudnessReference(75.0)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00, 0x00, 0x96, 0x42})) // float32 75.0
		})

		It("sends the correct bRequest", func() {
			_ = dev.SetLoudnessReference(75.0)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetLoudnessReference)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetLoudnessReference(75.0)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetLoudnessIntensity", func() {
		It("decodes the float32 LE response", func() {
			pct, err := dev.GetLoudnessIntensity()
			Expect(err).ToNot(HaveOccurred())
			Expect(pct).To(BeNumerically("~", 100.0, 0.01))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetLoudnessIntensity()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetLoudnessIntensity)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetLoudnessIntensity()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetLoudnessIntensity", func() {
		It("encodes 50% as little-endian float32", func() {
			err := dev.SetLoudnessIntensity(50.0)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00, 0x00, 0x48, 0x42})) // float32 50.0
		})

		It("sends the correct bRequest", func() {
			_ = dev.SetLoudnessIntensity(50.0)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetLoudnessIntensity)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetLoudnessIntensity(50.0)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
