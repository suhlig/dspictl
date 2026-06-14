package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Config", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetOutputType), 2, 0}:     {0x01},
				{uint16(dspi.ReqSetOutputType), 2, 0}:     {},
				{uint16(dspi.ReqGetOutputPin), 3, 0}:      {0x05},
				{uint16(dspi.ReqSetOutputPin), 0x0503, 0}: {0x00},
				{uint16(dspi.ReqGetI2SBckPin), 0, 0}:      {0x07},
				{uint16(dspi.ReqSetI2SBckPin), 7, 0}:      {0x00},
				{uint16(dspi.ReqGetMCKEnable), 0, 0}:      {0x01},
				{uint16(dspi.ReqSetMCKEnable), 0, 0}:      {},
				{uint16(dspi.ReqGetMCKPin), 0, 0}:         {0x09},
				{uint16(dspi.ReqSetMCKPin), 0, 0}:         {},
				{uint16(dspi.ReqGetMCKMultiplier), 0, 0}:  {0x01},
				{uint16(dspi.ReqSetMCKMultiplier), 0, 0}:  {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("SetOutputType", func() {
		It("sends the correct bRequest, wValue, and payload", func() {
			err := dev.SetOutputType(2, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputType)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputType(2, 1)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetOutputType", func() {
		It("decodes the response byte as I2S", func() {
			outputType, err := dev.GetOutputType(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(outputType).To(Equal(1))
		})

		It("decodes the response byte as S/PDIF", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetOutputType), 2, 0}] = []byte{0x00}
			outputType, err := dev.GetOutputType(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(outputType).To(Equal(0))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetOutputType(2)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputType)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(2)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputType(2)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetOutputPin", func() {
		It("sends the correct bRequest and wValue=(pin<<8)|output", func() {
			err := dev.SetOutputPin(3, 5)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0x0503)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error on non-zero status", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqSetOutputPin), 0x0503, 0}] = []byte{0x02}
			err := dev.SetOutputPin(3, 5)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("status 0x02"))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputPin(3, 5)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetOutputPin", func() {
		It("decodes the response byte as the pin number", func() {
			pin, err := dev.GetOutputPin(3)
			Expect(err).ToNot(HaveOccurred())
			Expect(pin).To(Equal(5))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetOutputPin(3)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(3)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputPin(3)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetI2SBckPin", func() {
		It("sends the correct bRequest and wValue", func() {
			err := dev.SetI2SBckPin(7)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetI2SBckPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(7)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error on non-zero status", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqSetI2SBckPin), 7, 0}] = []byte{0x04}
			err := dev.SetI2SBckPin(7)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("status 0x04"))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetI2SBckPin(7)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetI2SBckPin", func() {
		It("decodes the response byte as the pin number", func() {
			pin, err := dev.GetI2SBckPin()
			Expect(err).ToNot(HaveOccurred())
			Expect(pin).To(Equal(7))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetI2SBckPin()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetI2SBckPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetI2SBckPin()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetMCKEnable", func() {
		It("encodes true as 0x01", func() {
			err := dev.SetMCKEnable(true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("encodes false as 0x00", func() {
			err := dev.SetMCKEnable(false)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00}))
		})

		It("sends the correct bRequest", func() {
			_ = dev.SetMCKEnable(true)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetMCKEnable)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetMCKEnable(true)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetMCKEnable", func() {
		It("returns true when the byte is non-zero", func() {
			enabled, err := dev.GetMCKEnable()
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())
		})

		It("returns false when the byte is zero", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetMCKEnable), 0, 0}] = []byte{0x00}
			enabled, err := dev.GetMCKEnable()
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeFalse())
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetMCKEnable()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetMCKEnable)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetMCKEnable()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetMCKPin", func() {
		It("sends the correct bRequest and payload", func() {
			err := dev.SetMCKPin(9)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetMCKPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x09}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetMCKPin(9)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetMCKPin", func() {
		It("decodes the response byte as the pin number", func() {
			pin, err := dev.GetMCKPin()
			Expect(err).ToNot(HaveOccurred())
			Expect(pin).To(Equal(9))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetMCKPin()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetMCKPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetMCKPin()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetMCKMultiplier", func() {
		It("sends the correct bRequest and payload", func() {
			err := dev.SetMCKMultiplier(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetMCKMultiplier)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetMCKMultiplier(1)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetMCKMultiplier", func() {
		It("decodes the response byte as the multiplier", func() {
			multiplier, err := dev.GetMCKMultiplier()
			Expect(err).ToNot(HaveOccurred())
			Expect(multiplier).To(Equal(1))
		})

		It("decodes 0x00 as 128x", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetMCKMultiplier), 0, 0}] = []byte{0x00}
			multiplier, err := dev.GetMCKMultiplier()
			Expect(err).ToNot(HaveOccurred())
			Expect(multiplier).To(Equal(0))
		})

		It("sends the correct bRequest", func() {
			_, _ = dev.GetMCKMultiplier()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetMCKMultiplier)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetMCKMultiplier()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
