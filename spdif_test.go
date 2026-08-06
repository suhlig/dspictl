package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("S/PDIF multi-input", func() {
	It("sets an optional input pin with (index << 8) | pin in wValue", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetSpdifRxPin), 0x0114, 2}: {dspi.PinConfigSuccess},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetSpdifRxPinForIndex(1, 20)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0x0114)))
	})

	It("fails on a non-success status byte", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetSpdifRxPin), 0x0214, 2}: {dspi.PinConfigPinInUse},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetSpdifRxPinForIndex(2, 20)
		Expect(err).To(MatchError(ContainSubstring("status 0x02")))
	})

	It("gets an optional input pin by index", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetSpdifRxPin), 3, 2}: {22},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		pin, err := dev.GetSpdifRxPinForIndex(3)
		Expect(err).ToNot(HaveOccurred())
		Expect(pin).To(Equal(22))
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(3)))
	})

	It("keeps the legacy wrappers on input 1", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetSpdifRxPin), 5, 2}: {dspi.PinConfigSuccess},
			{uint16(dspi.ReqGetSpdifRxPin), 0, 2}: {5},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetSpdifRxPin(5)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))

		pin, err := dev.GetSpdifRxPin()
		Expect(err).ToNot(HaveOccurred())
		Expect(pin).To(Equal(5))
	})

	It("enables an optional input and surfaces the status byte", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetSpdifInputEnable), 0x0201, 2}: {dspi.PinConfigSuccess},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetSpdifInputEnable(2, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0x0201)))

		mock2 := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetSpdifInputEnable), 0x0300, 2}: {dspi.PinConfigPinInUse},
		}}
		dev2 := newTestDevice(mock2, dspi.PlatformRP2350)
		err = dev2.SetSpdifInputEnable(3, false)
		Expect(err).To(MatchError(ContainSubstring("status 0x02")))
	})

	It("reads the input inventory (count, enable mask, pins)", func() {
		raw := []byte{4, 0x03, 5, 20, 21, 22} // 4 inputs, inputs 1+2 enabled, GPIOs
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetSpdifInputConfig), 0, 2}: raw,
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		cfg, err := dev.GetSpdifInputConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Count).To(Equal(4))
		Expect(cfg.EnableMask).To(Equal(uint8(0x03)))
		Expect(cfg.Pins).To(Equal([]uint8{5, 20, 21, 22}))
	})

	It("handles short reads from older firmware", func() {
		raw := []byte{3, 0x01, 5, 20, 21} // 3-input firmware
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetSpdifInputConfig), 0, 2}: raw,
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		cfg, err := dev.GetSpdifInputConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Count).To(Equal(3))
		Expect(cfg.Pins).To(Equal([]uint8{5, 20, 21}))
	})
})
