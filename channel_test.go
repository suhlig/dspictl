package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("ChannelTable", func() {
	var (
		rp2040 []dspi.ChannelInfo
		rp2350 []dspi.ChannelInfo
	)

	BeforeEach(func() {
		rp2040 = dspi.ChannelTable(dspi.PlatformRP2040)
		rp2350 = dspi.ChannelTable(dspi.PlatformRP2350)
	})

	Context("RP2040 platform", func() {
		It("has 7 channels", func() {
			Expect(rp2040).To(HaveLen(7))
		})

		It("indexes start at 0", func() {
			Expect(rp2040[0].Index).To(Equal(0))
		})

		It("first channel is named USB L", func() {
			Expect(rp2040[0].Name).To(Equal("USB L"))
		})

		It("first channel is in USB Input group", func() {
			Expect(rp2040[0].Group).To(Equal("USB Input"))
		})

		It("second channel is named USB R", func() {
			Expect(rp2040[1].Name).To(Equal("USB R"))
		})

		It("second channel is in USB Input group", func() {
			Expect(rp2040[1].Group).To(Equal("USB Input"))
		})

		It("last channel is named PDM Sub", func() {
			Expect(rp2040[6].Name).To(Equal("PDM Sub"))
		})

		It("last channel is in PDM Sub group", func() {
			Expect(rp2040[6].Group).To(Equal("PDM Sub"))
		})
	})

	Context("RP2350 platform", func() {
		It("has 11 channels", func() {
			Expect(rp2350).To(HaveLen(11))
		})

		It("first channel is named USB L", func() {
			Expect(rp2350[0].Name).To(Equal("USB L"))
		})

		It("last channel is named PDM Sub", func() {
			Expect(rp2350[10].Name).To(Equal("PDM Sub"))
		})
	})
})
