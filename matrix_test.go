package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Matrix", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetMatrixRoute), 0x0203, 2}: {0x02, 0x03, 0x01, 0x01, 0x00, 0x00, 0xC0, 0xC0},
				{uint16(dspi.ReqSetMatrixRoute), 0, 2}:      {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("SetMatrixRoute", func() {
		It("encodes the route as an 8-byte payload", func() {
			route := &dspi.MatrixRoute{
				Input:       1,
				Output:      2,
				Enabled:     true,
				PhaseInvert: false,
				Gain:        dspi.NewGain(-6.0),
			}
			err := dev.SetMatrixRoute(route)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{
				0x01, 0x02, 0x01, 0x00,
				0x00, 0x00, 0xC0, 0xC0,
			}))
		})

		It("sends the correct bRequest and wValue", func() {
			route := &dspi.MatrixRoute{
				Input:  0,
				Output: 0,
				Gain:   dspi.NewGain(0.0),
			}
			err := dev.SetMatrixRoute(route)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetMatrixRoute)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetMatrixRoute(&dspi.MatrixRoute{Gain: dspi.NewGain(0.0)})
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetMatrixRoute", func() {
		It("decodes the 8-byte response into a MatrixRoute", func() {
			route, err := dev.GetMatrixRoute(2, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(route.Input).To(Equal(2))
			Expect(route.Output).To(Equal(3))
			Expect(route.Enabled).To(BeTrue())
			Expect(route.PhaseInvert).To(BeTrue())
			Expect(route.Gain.DB()).To(BeNumerically("~", -6.0, 0.001))
		})

		It("sends the correct bRequest and wValue", func() {
			_, _ = dev.GetMatrixRoute(2, 3)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetMatrixRoute)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0x0203)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetMatrixRoute(0, 0)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
