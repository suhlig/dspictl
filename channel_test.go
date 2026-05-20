package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Channel", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetChannelName), 0, 0}:  {'C', 'h', '0', 0},
				{uint16(dspi.ReqGetChannelName), 1, 0}:  {'C', 'h', '1', 0},
				{uint16(dspi.ReqGetChannelName), 2, 0}:  {'C', 'h', '2', 0},
				{uint16(dspi.ReqGetChannelName), 3, 0}:  {'C', 'h', '3', 0},
				{uint16(dspi.ReqGetChannelName), 4, 0}:  {'C', 'h', '4', 0},
				{uint16(dspi.ReqGetChannelName), 5, 0}:  {'C', 'h', '5', 0},
				{uint16(dspi.ReqGetChannelName), 6, 0}:  {'C', 'h', '6', 0},
				{uint16(dspi.ReqGetChannelName), 7, 0}:  {'C', 'h', '7', 0},
				{uint16(dspi.ReqGetChannelName), 8, 0}:  {'C', 'h', '8', 0},
				{uint16(dspi.ReqGetChannelName), 9, 0}:  {'C', 'h', '9', 0},
				{uint16(dspi.ReqGetChannelName), 10, 0}: {'C', 'h', '1', '0', 0},
				{uint16(dspi.ReqSetChannelName), 0, 0}:  {},
				{uint16(dspi.ReqSetChannelName), 3, 0}:  {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("SetChannelName", func() {
		It("encodes the name in a 32-byte payload", func() {
			err := dev.SetChannelName(0, "Test")
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			expected := make([]byte, 32)
			copy(expected, "Test")
			Expect(mock.CapturedRequests[0].Data).To(Equal(expected))
		})

		It("sends the correct bRequest and wValue", func() {
			err := dev.SetChannelName(3, "Name")
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetChannelName)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(3)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetChannelName(0, "Name")
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("ChannelName", func() {
		It("decodes a null-terminated string response", func() {
			name, err := dev.ChannelName(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("Ch0"))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.ChannelName(5)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetChannelName)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.ChannelName(0)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("Channels", func() {
		It("returns 7 channels for RP2040", func() {
			dev2040 := newTestDevice(mock, dspi.PlatformRP2040)
			channels, err := dev2040.Channels()
			Expect(err).ToNot(HaveOccurred())
			Expect(channels).To(HaveLen(7))
			for i, ch := range channels {
				Expect(ch.Index).To(Equal(i))
			}
			Expect(channels[0].Name).To(Equal("Ch0"))
			Expect(channels[6].Name).To(Equal("Ch6"))
		})

		It("returns 11 channels for RP2350", func() {
			channels, err := dev.Channels()
			Expect(err).ToNot(HaveOccurred())
			Expect(channels).To(HaveLen(11))
			for i, ch := range channels {
				Expect(ch.Index).To(Equal(i))
			}
			Expect(channels[0].Name).To(Equal("Ch0"))
			Expect(channels[10].Name).To(Equal("Ch10"))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.Channels()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
