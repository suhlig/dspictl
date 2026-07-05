package dspi_test

import (
	"encoding/binary"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dspi "github.com/suhlig/dspi"
)

var _ = Describe("CrossoverBand", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetEQParam), 0, 2}: {},
				{uint16(dspi.ReqGetEQParam), 0x02A0, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, uint32(dspi.CrossoverTypeLR4LP))
					return b
				}(),
				{uint16(dspi.ReqGetEQParam), 0x02A1, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, math.Float32bits(800.0))
					return b
				}(),
				{uint16(dspi.ReqGetEQParam), 0x02A4, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, 1)
					return b
				}(),
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2040)
	})

	Describe("CrossoverFilterType", func() {
		It("formats known types", func() {
			Expect(dspi.CrossoverTypeLR2LP.String()).To(Equal("lr2-lp"))
			Expect(dspi.CrossoverTypeLR8HP.String()).To(Equal("lr8-hp"))
			Expect(dspi.CrossoverTypeBW4LP.String()).To(Equal("bw4-lp"))
			Expect(dspi.CrossoverTypeBES8HP.String()).To(Equal("bes8-hp"))
		})

		It("parses known type strings", func() {
			t, err := dspi.ParseCrossoverFilterType("lr4-lp")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.CrossoverTypeLR4LP))

			t, err = dspi.ParseCrossoverFilterType("bw2-hp")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.CrossoverTypeBW2HP))

			t, err = dspi.ParseCrossoverFilterType("bes6-lp")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.CrossoverTypeBES6LP))
		})

		It("rejects unknown type strings", func() {
			_, err := dspi.ParseCrossoverFilterType("lr3-lp")
			Expect(err).To(MatchError(ContainSubstring("unknown crossover filter type")))
		})

		It("rejects PEQ type strings", func() {
			_, err := dspi.ParseCrossoverFilterType("peak")
			Expect(err).To(MatchError(ContainSubstring("unknown crossover filter type")))
		})
	})

	Describe("CrossoverBand.Validate", func() {
		It("accepts a valid band", func() {
			band := &dspi.CrossoverBand{
				Channel: 2,
				Band:    20,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
			}
			Expect(band.Validate(6)).To(Succeed())
		})

		It("rejects master channels", func() {
			band := &dspi.CrossoverBand{
				Channel: 0,
				Band:    20,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
			}
			Expect(band.Validate(6)).To(MatchError(ContainSubstring("not supported on master channels")))
		})

		It("rejects a band below the crossover range", func() {
			band := &dspi.CrossoverBand{
				Channel: 2,
				Band:    10,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
			}
			Expect(band.Validate(6)).To(MatchError(ContainSubstring("not a crossover band")))
		})

		It("rejects a band above the crossover range", func() {
			band := &dspi.CrossoverBand{
				Channel: 2,
				Band:    24,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
			}
			Expect(band.Validate(6)).To(MatchError(ContainSubstring("not a crossover band")))
		})

		It("rejects a non-positive frequency", func() {
			band := &dspi.CrossoverBand{
				Channel: 2,
				Band:    20,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    0,
			}
			Expect(band.Validate(6)).To(MatchError(ContainSubstring("frequency must be > 0")))
		})

		It("rejects an invalid type", func() {
			band := &dspi.CrossoverBand{
				Channel: 2,
				Band:    20,
				Type:    10, // PEQ type
				Freq:    800,
			}
			Expect(band.Validate(6)).To(MatchError(ContainSubstring("invalid crossover type")))
		})

		It("rejects an out-of-range channel", func() {
			band := &dspi.CrossoverBand{
				Channel: 99,
				Band:    20,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
			}
			Expect(band.Validate(6)).To(MatchError(ContainSubstring("channel 99 out of range")))
		})
	})

	Describe("SetCrossoverBand", func() {
		It("sends the correct 16-byte packet", func() {
			band := &dspi.CrossoverBand{
				Channel: 2,
				Band:    20,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
				Bypass:  true,
			}

			err := dev.SetCrossoverBand(band)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetEQParam)))

			data := mock.CapturedRequests[0].Data
			Expect(data).To(HaveLen(16))
			Expect(data[0]).To(Equal(byte(2)))
			Expect(data[1]).To(Equal(byte(20)))
			Expect(data[2]).To(Equal(byte(dspi.CrossoverTypeLR4LP)))
			Expect(data[3]).To(Equal(byte(1))) // bypass
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))).To(BeNumerically("~", 800.0, 0.01))
		})

		It("rejects master channels", func() {
			band := &dspi.CrossoverBand{
				Channel: 0,
				Band:    20,
				Type:    dspi.CrossoverTypeLR4LP,
				Freq:    800,
			}
			err := dev.SetCrossoverBand(band)
			Expect(err).To(MatchError(ContainSubstring("not supported on master channels")))
		})
	})

	Describe("GetCrossoverBand", func() {
		It("parses the response correctly", func() {
			band, err := dev.GetCrossoverBand(2, 20)
			Expect(err).ToNot(HaveOccurred())
			Expect(band.Channel).To(Equal(2))
			Expect(band.Band).To(Equal(20))
			Expect(band.Type).To(Equal(dspi.CrossoverTypeLR4LP))
			Expect(band.Freq).To(BeNumerically("~", 800.0, 0.01))
			Expect(band.Bypass).To(BeTrue())
		})

		It("sends the correct requests", func() {
			_, _ = dev.GetCrossoverBand(2, 20)
			Expect(mock.CapturedRequests).To(HaveLen(3))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetEQParam)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0x02A0))) // band 20 << 3
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(0x02A1)))
			Expect(mock.CapturedRequests[2].WValue).To(Equal(uint16(0x02A4)))
		})

		It("rejects master channels", func() {
			_, err := dev.GetCrossoverBand(0, 20)
			Expect(err).To(MatchError(ContainSubstring("not supported on master channels")))
		})

		It("rejects a band below the crossover range", func() {
			_, err := dev.GetCrossoverBand(2, 10)
			Expect(err).To(MatchError(ContainSubstring("not a crossover band")))
		})
	})

	Describe("MaxCrossoverBands", func() {
		It("returns 4", func() {
			Expect(dev.MaxCrossoverBands()).To(Equal(4))
		})
	})

	Describe("BandBypass for crossover bands", func() {
		It("allows setting bypass on a crossover band", func() {
			key1 := [3]uint16{uint16(dspi.ReqSetBandBypass), 0x0214, 2}
			mock.ReturnData[key1] = []byte{}
			key2 := [3]uint16{uint16(dspi.ReqGetAllParamsChunk), 0, 2}
			mock.ReturnData[key2] = func() []byte {
				b := make([]byte, 16)
				b[0] = 1
				b[1] = byte(dspi.PlatformRP2040)
				b[4] = 2
				b[5] = 12
				binary.LittleEndian.PutUint16(b[6:8], 16)
				return b
			}()
			key3 := [3]uint16{uint16(dspi.ReqGetAllParamsChunk), 16, 2}
			mock.ReturnData[key3] = make([]byte, 0)

			err := dev.SetBandBypass(2, 20, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqSetBandBypass)))
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(0x0214)))
		})
	})
})
