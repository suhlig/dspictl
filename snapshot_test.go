package dspi_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Level", func() {
	Describe("NewLevel", func() {
		It("returns a Level with the given linear value", func() {
			l := dspi.NewLevel(0.5)

			Expect(l.Linear()).To(Equal(0.5))
		})
	})

	Describe("DBFS", func() {
		Context("with zero input", func() {
			It("returns -Inf", func() {
				l := dspi.NewLevel(0)

				Expect(l.DBFS()).To(Equal(math.Inf(-1)))
			})
		})

		Context("with negative input", func() {
			It("returns -Inf", func() {
				l := dspi.NewLevel(-1)

				Expect(l.DBFS()).To(Equal(math.Inf(-1)))
			})
		})

		Context("with full-scale input (1.0)", func() {
			It("returns 0", func() {
				l := dspi.NewLevel(1)

				Expect(l.DBFS()).To(Equal(0.0))
			})
		})

		Context("with half-scale input (0.5)", func() {
			It("returns approximately -6.02", func() {
				l := dspi.NewLevel(0.5)

				Expect(l.DBFS()).To(BeNumerically("~", -6.0206, 0.001))
			})
		})

		Context("with 10%% input (0.1)", func() {
			It("returns approximately -20", func() {
				l := dspi.NewLevel(0.1)

				Expect(l.DBFS()).To(BeNumerically("~", -20.0, 0.001))
			})
		})
	})

	Describe("String", func() {
		Context("with -Inf dBFS", func() {
			It("returns -inf dBFS", func() {
				l := dspi.NewLevel(0)

				Expect(l.String()).To(Equal("-inf dBFS"))
			})
		})

		Context("with zero dBFS", func() {
			It("formats as 0.0 dBFS", func() {
				l := dspi.NewLevel(1)

				Expect(l.String()).To(Equal("0.0 dBFS"))
			})
		})

		Context("with -6.0 dBFS", func() {
			It("formats as -6.0 dBFS", func() {
				l := dspi.NewLevel(0.5)

				Expect(l.String()).To(Equal("-6.0 dBFS"))
			})
		})

		Context("with -20 dBFS", func() {
			It("formats as -20.0 dBFS", func() {
				l := dspi.NewLevel(0.1)

				Expect(l.String()).To(Equal("-20.0 dBFS"))
			})
		})

		Context("with positive 3.0 dBFS", func() {
			It("formats as 3.0 dBFS", func() {
				l := dspi.NewLevel(1.413)

				Expect(l.String()).To(Equal("3.0 dBFS"))
			})
		})
	})
})
