package dspi_test

import (
	"github.com/google/gousb"
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
				{uint16(dspi.ReqFactoryReset), 0, 2}:     {},
				{uint16(dspi.ReqEnterBootloader), 0, 2}:  {},
				{uint16(dspi.ReqGetCore1Mode), 0, 2}:     {0x02},
				{uint16(dspi.ReqGetCore1Conflict), 0, 2}: {0x01},
				{uint16(dspi.ReqGetBufferStats), 0, 2}:   {0xAB, 0xCD, 0xEF},
				{uint16(dspi.ReqGetUSBErrorStats), 0, 2}: {
					// 24 bytes: CRC=1, BitStuff=2, Timeout=3, Overflow=4, Sequence=5, Unknown=6 as LE uint32
					0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
					0x04, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00,
				},
				{uint16(dspi.ReqGetSerial), 0, 2}: {
					'E', '6', '6', '1', '4', '1', '0', '3', 'E', '3', '2', 'C', '3', 'B', '2', 'D',
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns nil when the device disconnects", func() {
			mock.ReturnErrors = map[[3]uint16]error{
				{uint16(dspi.ReqEnterBootloader), 0, 2}: gousb.ErrorNoDevice,
			}

			err := dev.EnterBootloader()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error for non-libusb errors", func() {
			delete(mock.ReturnData, [3]uint16{uint16(dspi.ReqEnterBootloader), 0, 2})

			err := dev.EnterBootloader()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("REQ_ENTER_BOOTLOADER"))
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
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
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetUSBErrorStats()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetSerial", func() {
		It("returns the 16-byte serial string", func() {
			serial, err := dev.GetSerial()
			Expect(err).ToNot(HaveOccurred())
			Expect(serial).To(Equal("E6614103E32C3B2D"))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetSerial()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetSerial)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetSerial()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
