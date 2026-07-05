package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("ReadMeter", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: make(map[[3]uint16][]byte),
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("request parameters", func() {
		BeforeEach(func() {
			// V16: 7 channels * 2 + 7 trailer = 21 bytes
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = make([]byte, 21)
		})

		It("sends ReqGetStatus with wValue=9", func() {
			_ = dev.ReadMeter()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetStatus)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(9)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})
	})

	Describe("response length parsing", func() {
		Context("7 channels (21 bytes V16)", func() {
			BeforeEach(func() {
				// V16: 7 * 2 + 7 trailer = 21 bytes
				mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = make([]byte, 21)
			})

			It("determines 7 channels", func() {
				snap := dev.ReadMeter()
				Expect(snap.Err()).ToNot(HaveOccurred())
				Expect(snap.Channels).To(Equal(7))
			})
		})

		Context("17 channels (41 bytes V16)", func() {
			BeforeEach(func() {
				// V16: 17 * 2 + 7 trailer = 41 bytes
				mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = make([]byte, 41)
			})

			It("determines 17 channels", func() {
				snap := dev.ReadMeter()
				Expect(snap.Err()).ToNot(HaveOccurred())
				Expect(snap.Channels).To(Equal(17))
			})
		})
	})

	Describe("peak normalization", func() {
		BeforeEach(func() {
			data := make([]byte, 21)
			// peak 0: 0x7FFF (32767) -> 1.0
			data[0] = 0xFF
			data[1] = 0x7F
			// peak 1: 0x4000 (16384) -> ~0.5
			data[2] = 0x00
			data[3] = 0x40
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = data
		})

		It("normalizes 0x7FFF to Level 1.0", func() {
			snap := dev.ReadMeter()
			Expect(snap.Err()).ToNot(HaveOccurred())
			Expect(snap.Peaks[0].Linear()).To(BeNumerically("~", 1.0, 0.0001))
		})

		It("normalizes 0x4000 to Level ~0.5", func() {
			snap := dev.ReadMeter()
			Expect(snap.Err()).ToNot(HaveOccurred())
			Expect(snap.Peaks[1].Linear()).To(BeNumerically("~", 0.5, 0.0001))
		})
	})

	Describe("CPU0/CPU1 extraction", func() {
		BeforeEach(func() {
			data := make([]byte, 21)
			data[14] = 42
			data[15] = 99
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = data
		})

		It("extracts CPU0 and CPU1 from the correct offsets", func() {
			snap := dev.ReadMeter()
			Expect(snap.Err()).ToNot(HaveOccurred())
			Expect(snap.CPU0).To(Equal(42))
			Expect(snap.CPU1).To(Equal(99))
		})
	})

	Describe("ClipFlags decoding", func() {
		BeforeEach(func() {
			data := make([]byte, 21)
			// ClipFlags as little-endian uint32
			data[16] = 0xAB
			data[17] = 0xCD
			data[18] = 0x01
			data[19] = 0x00
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = data
		})

		It("decodes as little-endian uint32", func() {
			snap := dev.ReadMeter()
			Expect(snap.Err()).ToNot(HaveOccurred())
			Expect(snap.ClipFlags).To(Equal(uint32(0x0001CDAB)))
		})
	})

	Describe("when the device is closed", func() {
		It("returns an error via snap.Err()", func() {
			dev.Close()
			snap := dev.ReadMeter()
			Expect(snap.Err()).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("with an unexpected response length", func() {
		BeforeEach(func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetStatus), 9, 2}] = []byte{0x01, 0x02, 0x03}
		})

		It("returns an error", func() {
			snap := dev.ReadMeter()
			Expect(snap.Err()).To(MatchError(ContainSubstring("unexpected response length")))
		})
	})
})
