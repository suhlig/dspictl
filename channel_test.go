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
				{uint16(dspi.ReqGetChannelName), 0, 2}:  {'C', 'h', '0', 0},
				{uint16(dspi.ReqGetChannelName), 1, 2}:  {'C', 'h', '1', 0},
				{uint16(dspi.ReqGetChannelName), 2, 2}:  {'C', 'h', '2', 0},
				{uint16(dspi.ReqGetChannelName), 3, 2}:  {'C', 'h', '3', 0},
				{uint16(dspi.ReqGetChannelName), 4, 2}:  {'C', 'h', '4', 0},
				{uint16(dspi.ReqGetChannelName), 5, 2}:  {'C', 'h', '5', 0},
				{uint16(dspi.ReqGetChannelName), 6, 2}:  {'C', 'h', '6', 0},
				{uint16(dspi.ReqGetChannelName), 7, 2}:  {'C', 'h', '7', 0},
				{uint16(dspi.ReqGetChannelName), 8, 2}:  {'C', 'h', '8', 0},
				{uint16(dspi.ReqGetChannelName), 9, 2}:  {'C', 'h', '9', 0},
				{uint16(dspi.ReqGetChannelName), 10, 2}: {'C', 'h', '1', '0', 0},
				{uint16(dspi.ReqGetChannelName), 11, 2}: {'C', 'h', '1', '1', 0},
				{uint16(dspi.ReqGetChannelName), 12, 2}: {'C', 'h', '1', '2', 0},
				{uint16(dspi.ReqGetChannelName), 13, 2}: {'C', 'h', '1', '3', 0},
				{uint16(dspi.ReqGetChannelName), 14, 2}: {'C', 'h', '1', '4', 0},
				{uint16(dspi.ReqGetChannelName), 15, 2}: {'C', 'h', '1', '5', 0},
				{uint16(dspi.ReqGetChannelName), 16, 2}: {'C', 'h', '1', '6', 0},
				{uint16(dspi.ReqSetChannelName), 0, 2}:  {},
				{uint16(dspi.ReqSetChannelName), 3, 2}:  {},
				{uint16(dspi.ReqGetInputSource), 0, 2}:  {0x00},
				// Mock GetAllParamsChunk for NumInputChannels
				{uint16(dspi.ReqGetAllParamsChunk), 0, 2}: func() []byte {
					b := make([]byte, 16)
					b[4] = 8 // num_input_channels for RP2350
					b[6] = 0x10
					return b
				}(),
				{uint16(dspi.ReqGetAllParamsChunk), 16, 2}: make([]byte, 0),
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
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

		It("returns 17 channels for RP2350", func() {
			channels, err := dev.Channels()
			Expect(err).ToNot(HaveOccurred())
			Expect(channels).To(HaveLen(17))
			for i, ch := range channels {
				Expect(ch.Index).To(Equal(i))
			}
			Expect(channels[0].Name).To(Equal("Ch0"))
			Expect(channels[16].Name).To(Equal("Ch16"))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.Channels()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
