package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Preamp", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetPreampCh), 3, 0}: {0x00, 0x00, 0x20, 0x41}, // float32 10.0
				{uint16(dspi.ReqSetPreampCh), 3, 0}: {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("GetPreampChannel", func() {
		It("decodes the float32 LE response", func() {
			gain, err := dev.GetPreampChannel(3)
			Expect(err).ToNot(HaveOccurred())
			Expect(gain.DB()).To(BeNumerically("~", 10.0, 0.001))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetPreampChannel(3)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetPreampCh)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(3)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetPreampChannel(3)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetPreampChannel", func() {
		It("encodes -6 dB as little-endian float32", func() {
			err := dev.SetPreampChannel(3, dspi.NewGain(-6.0))
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00, 0x00, 0xc0, 0xc0}))
		})

		It("sends the correct bRequest and wValue", func() {
			_ = dev.SetPreampChannel(3, dspi.NewGain(0.0))
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetPreampCh)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(3)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetPreampChannel(3, dspi.NewGain(0.0))
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
