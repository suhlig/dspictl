package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Platform", func() {
	Describe("String", func() {
		It("returns RP2040 for RP2040 platform", func() {
			Expect(dspi.PlatformRP2040.String()).To(Equal("RP2040"))
		})

		It("returns RP2350 for RP2350 platform", func() {
			Expect(dspi.PlatformRP2350.String()).To(Equal("RP2350"))
		})

		It("returns Unknown for unrecognized platform value", func() {
			Expect(dspi.Platform(2).String()).To(Equal("Unknown"))
		})

		It("returns Unknown for negative platform value", func() {
			Expect(dspi.Platform(-1).String()).To(Equal("Unknown"))
		})
	})
})
