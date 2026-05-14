package dspi_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("DBFS", func() {
	Context("with zero input", func() {
		It("returns -Inf", func() {
			Expect(dspi.DBFS(0)).To(Equal(math.Inf(-1)))
		})
	})

	Context("with negative input", func() {
		It("returns -Inf", func() {
			Expect(dspi.DBFS(-1)).To(Equal(math.Inf(-1)))
		})
	})

	Context("with full-scale input (1.0)", func() {
		It("returns 0", func() {
			Expect(dspi.DBFS(1)).To(Equal(0.0))
		})
	})

	Context("with half-scale input (0.5)", func() {
		It("returns approximately -6.02", func() {
			Expect(dspi.DBFS(0.5)).To(BeNumerically("~", -6.0206, 0.001))
		})
	})

	Context("with 10%% input (0.1)", func() {
		It("returns approximately -20", func() {
			Expect(dspi.DBFS(0.1)).To(BeNumerically("~", -20.0, 0.001))
		})
	})
})

var _ = Describe("FormatDBFS", func() {
	Context("with -Inf", func() {
		It("returns -inf", func() {
			Expect(dspi.FormatDBFS(math.Inf(-1))).To(Equal("-inf"))
		})
	})

	Context("with zero", func() {
		It("formats as 0.0", func() {
			Expect(dspi.FormatDBFS(0)).To(Equal("0.0"))
		})
	})

	Context("with -6.0", func() {
		It("formats as -6.0", func() {
			Expect(dspi.FormatDBFS(-6.0)).To(Equal("-6.0"))
		})
	})

	Context("with -20", func() {
		It("formats as -20.0", func() {
			Expect(dspi.FormatDBFS(-20)).To(Equal("-20.0"))
		})
	})

	Context("with positive 3.0", func() {
		It("formats as 3.0", func() {
			Expect(dspi.FormatDBFS(3.0)).To(Equal("3.0"))
		})
	})
})
