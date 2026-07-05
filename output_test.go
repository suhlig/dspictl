package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Output", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetOutputGain), 2, 2}:   {0x00, 0x00, 0x20, 0x41}, // float32 10.0
				{uint16(dspi.ReqSetOutputGain), 2, 2}:   {},
				{uint16(dspi.ReqGetOutputMute), 2, 2}:   {0x01},
				{uint16(dspi.ReqSetOutputMute), 2, 2}:   {},
				{uint16(dspi.ReqGetOutputDelay), 2, 2}:  {0x00, 0x00, 0xa0, 0x40}, // float32 5.0
				{uint16(dspi.ReqSetOutputDelay), 2, 2}:  {},
				{uint16(dspi.ReqGetOutputEnable), 2, 2}: {0x01},
				{uint16(dspi.ReqSetOutputEnable), 2, 2}: {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("GetOutputGain", func() {
		It("decodes the float32 LE response", func() {
			gain, err := dev.GetOutputGain(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(gain.DB()).To(BeNumerically("~", 10.0, 0.001))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetOutputGain(2)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputGain)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputGain(2)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetOutputGain", func() {
		It("encodes -6 dB as little-endian float32", func() {
			err := dev.SetOutputGain(2, dspi.NewGain(-6.0))
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00, 0x00, 0xc0, 0xc0}))
		})

		It("sends the correct bRequest and wValue", func() {
			_ = dev.SetOutputGain(2, dspi.NewGain(0.0))
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputGain)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputGain(2, dspi.NewGain(0.0))
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetOutputMute", func() {
		It("returns true when the byte is non-zero", func() {
			muted, err := dev.GetOutputMute(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(muted).To(BeTrue())
		})

		It("returns false when the byte is zero", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetOutputMute), 2, 2}] = []byte{0x00}
			muted, err := dev.GetOutputMute(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(muted).To(BeFalse())
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetOutputMute(2)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputMute)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputMute(2)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetOutputMute", func() {
		It("encodes true as 0x01", func() {
			err := dev.SetOutputMute(2, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("encodes false as 0x00", func() {
			err := dev.SetOutputMute(2, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00}))
		})

		It("sends the correct bRequest and wValue", func() {
			_ = dev.SetOutputMute(2, true)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputMute)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputMute(2, true)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetOutputDelay", func() {
		It("decodes the float32 LE response", func() {
			delay, err := dev.GetOutputDelay(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(delay).To(BeNumerically("~", 5.0, 0.001))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetOutputDelay(2)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputDelay)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputDelay(2)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetOutputDelay", func() {
		It("encodes 5.0 ms as little-endian float32", func() {
			err := dev.SetOutputDelay(2, 5.0)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00, 0x00, 0xa0, 0x40}))
		})

		It("sends the correct bRequest and wValue", func() {
			_ = dev.SetOutputDelay(2, 0.0)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputDelay)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputDelay(2, 0.0)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetOutputEnable", func() {
		It("returns true when the byte is non-zero", func() {
			enabled, err := dev.GetOutputEnable(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())
		})

		It("returns false when the byte is zero", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetOutputEnable), 2, 2}] = []byte{0x00}
			enabled, err := dev.GetOutputEnable(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeFalse())
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetOutputEnable(2)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputEnable)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputEnable(2)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetOutputEnable", func() {
		It("encodes true as 0x01", func() {
			err := dev.SetOutputEnable(2, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("encodes false as 0x00", func() {
			err := dev.SetOutputEnable(2, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00}))
		})

		It("sends the correct bRequest and wValue", func() {
			_ = dev.SetOutputEnable(2, true)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputEnable)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputEnable(2, true)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
