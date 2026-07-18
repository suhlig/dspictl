package dspi_test

import (
	"encoding/binary"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Psybass", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetPsybass), 0, 2}:          {0x01},
				{uint16(dspi.ReqGetPsybassMask), 0, 2}:      {0xFF, 0xFF},
				{uint16(dspi.ReqSetPsybass), 0, 2}:          {},
				{uint16(dspi.ReqSetPsybassMask), 0, 2}:      {},
				{uint16(dspi.ReqGetPsybassCutoff), 0, 2}:    float32Bytes(80.0),
				{uint16(dspi.ReqSetPsybassCutoff), 0, 2}:    {},
				{uint16(dspi.ReqGetPsybassHarmonics), 0, 2}: float32Bytes(-12.0),
				{uint16(dspi.ReqSetPsybassHarmonics), 0, 2}: {},
				{uint16(dspi.ReqGetPsybassDrive), 0, 2}:     float32Bytes(6.0),
				{uint16(dspi.ReqSetPsybassDrive), 0, 2}:     {},
				{uint16(dspi.ReqGetPsybassCharacter), 0, 2}: float32Bytes(50.0),
				{uint16(dspi.ReqSetPsybassCharacter), 0, 2}: {},
				{uint16(dspi.ReqGetPsybassOriginal), 0, 2}:  float32Bytes(-30.0),
				{uint16(dspi.ReqSetPsybassOriginal), 0, 2}:  {},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("GetPsybass", func() {
		It("returns enabled state and mask", func() {
			enabled, mask, err := dev.GetPsybass()
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())
			Expect(mask).To(Equal(uint16(0xFFFF)))
		})
	})

	Describe("SetPsybass", func() {
		It("sends enable and mask", func() {
			err := dev.SetPsybass(true, 0x1234)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetPsybass)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqSetPsybassMask)))
			Expect(binary.LittleEndian.Uint16(mock.CapturedRequests[1].Data)).To(Equal(uint16(0x1234)))
		})
	})

	Describe("GetPsybassCutoff", func() {
		It("returns the cutoff frequency", func() {
			v, err := dev.GetPsybassCutoff()
			Expect(err).ToNot(HaveOccurred())
			Expect(v).To(BeNumerically("~", float32(80.0), 0.001))
		})
	})

	Describe("SetPsybassCutoff", func() {
		It("sends a float32 value", func() {
			err := dev.SetPsybassCutoff(80)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(mock.CapturedRequests[0].Data))).To(BeNumerically("~", float32(80.0), 0.001))
		})
	})
})

func float32Bytes(v float32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(v))
	return b
}
