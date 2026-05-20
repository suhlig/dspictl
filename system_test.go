package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("System", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqFactoryReset), 0, 0}:     {},
				{uint16(dspi.ReqEnterBootloader), 0, 0}:  {},
				{uint16(dspi.ReqGetCore1Mode), 0, 0}:     {0x02},
				{uint16(dspi.ReqGetCore1Conflict), 0, 0}: {0x01},
				{uint16(dspi.ReqGetBufferStats), 0, 0}:   {0xAB, 0xCD, 0xEF},
				{uint16(dspi.ReqGetUSBErrorStats), 0, 0}: {
					// 24 bytes: CRC=1, BitStuff=2, Timeout=3, Overflow=4, Sequence=5, Unknown=6 as LE uint32
					0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
					0x04, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00,
				},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("FactoryReset", func() {
		It("sends the correct bRequest and wValue", func() {
			err := dev.FactoryReset()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqFactoryReset)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.FactoryReset()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("EnterBootloader", func() {
		It("sends the correct bRequest and wValue", func() {
			err := dev.EnterBootloader()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqEnterBootloader)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.EnterBootloader()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetCore1Mode", func() {
		It("returns the mode byte", func() {
			mode, err := dev.GetCore1Mode()
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(2))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetCore1Mode()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetCore1Mode)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetCore1Mode()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetCore1Conflict", func() {
		It("returns true for a non-zero response byte", func() {
			conflict, err := dev.GetCore1Conflict()
			Expect(err).ToNot(HaveOccurred())
			Expect(conflict).To(BeTrue())
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetCore1Conflict()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetCore1Conflict)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetCore1Conflict()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetBufferStats", func() {
		It("returns the response data up to the received length", func() {
			stats, err := dev.GetBufferStats()
			Expect(err).ToNot(HaveOccurred())
			Expect(stats.Data).To(Equal([]byte{0xAB, 0xCD, 0xEF}))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetBufferStats()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetBufferStats)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetBufferStats()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetUSBErrorStats", func() {
		It("decodes the 24-byte response into six uint32 LE fields", func() {
			stats, err := dev.GetUSBErrorStats()
			Expect(err).ToNot(HaveOccurred())
			Expect(stats.CRC).To(Equal(uint32(1)))
			Expect(stats.BitStuff).To(Equal(uint32(2)))
			Expect(stats.Timeout).To(Equal(uint32(3)))
			Expect(stats.Overflow).To(Equal(uint32(4)))
			Expect(stats.Sequence).To(Equal(uint32(5)))
			Expect(stats.Unknown).To(Equal(uint32(6)))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetUSBErrorStats()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetUSBErrorStats)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetUSBErrorStats()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
