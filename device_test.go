package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Device", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqClearClips), 0, 2}: {0x00, 0x00, 0x00, 0x00},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("NewDevice constructor", func() {
		It("exposes the platform via Platform()", func() {
			Expect(dev.Platform()).To(Equal(dspi.PlatformRP2350))
		})

		It("exposes the serial via Serial()", func() {
			Expect(dev.Serial()).To(Equal("test-serial"))
		})

		It("exposes the bus via Bus()", func() {
			Expect(dev.Bus()).To(Equal(1))
		})

		It("exposes the address via Address()", func() {
			Expect(dev.Address()).To(Equal(2))
		})
	})

	Describe("Platform", func() {
		It("returns the value passed to the constructor", func() {
			mock2 := &mockControlTransfer{ReturnData: make(map[[3]uint16][]byte)}
			dev2 := newTestDevice(mock2, dspi.PlatformRP2040)
			Expect(dev2.Platform()).To(Equal(dspi.PlatformRP2040))
		})
	})

	Describe("Close", func() {
		It("causes subsequent calls to return an error", func() {
			dev.Close()
			err := dev.ClearClips()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})

		It("is safe to call twice (does not panic)", func() {
			dev.Close()
			Expect(func() { dev.Close() }).ToNot(Panic())
		})
	})

	Describe("ClearClips", func() {
		It("sends ReqClearClips", func() {
			err := dev.ClearClips()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqClearClips)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("accepts a 4-byte response", func() {
			err := dev.ClearClips()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.ClearClips()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
