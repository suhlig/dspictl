package dspi_test

import (
	"encoding/binary"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dspi "github.com/suhlig/dspi"
)

var _ = Describe("EQBand", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetEQParam), 0, 2}: {},
				{uint16(dspi.ReqGetAllParamsChunk), 0, 2}: func() []byte {
					b := make([]byte, 16)
					b[0] = 1 // format version
					b[1] = byte(dspi.PlatformRP2040)
					b[4] = 2                                  // num_input_channels
					b[5] = 12                                 // max bands
					binary.LittleEndian.PutUint16(b[6:8], 16) // payload length
					return b
				}(),
				{uint16(dspi.ReqGetAllParamsChunk), 16, 2}: make([]byte, 0),
				{uint16(dspi.ReqGetEQParam), 0x0228, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, uint32(dspi.FilterTypePeaking))
					return b
				}(),
				{uint16(dspi.ReqGetEQParam), 0x0229, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, math.Float32bits(1000.0))
					return b
				}(),
				{uint16(dspi.ReqGetEQParam), 0x022A, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, math.Float32bits(1.5))
					return b
				}(),
				{uint16(dspi.ReqGetEQParam), 0x022B, 2}: func() []byte {
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, math.Float32bits(-3.5))
					return b
				}(),
				{uint16(dspi.ReqSetBypass), 0, 2}:          {},
				{uint16(dspi.ReqGetBypass), 0, 2}:          {1},
				{uint16(dspi.ReqSetBandBypass), 0x0203, 2}: {},
				{uint16(dspi.ReqGetBandBypass), 0x0203, 2}: {1},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2040)
	})

	Describe("SetEQBand", func() {
		It("sends the correct 16-byte packet", func() {
			band := &dspi.EQBand{
				Channel:       2,
				Band:          3,
				Type:          dspi.FilterTypePeaking,
				Freq:          1000,
				QualityFactor: 1.5,
				Gain:          -3.5,
			}

			err := dev.SetEQBand(band)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetAllParamsChunk)))
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqSetEQParam)))

			data := mock.CapturedRequests[1].Data
			Expect(data).To(HaveLen(16))
			Expect(data[0]).To(Equal(byte(2)))
			Expect(data[1]).To(Equal(byte(3)))
			Expect(data[2]).To(Equal(byte(1)))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))).To(BeNumerically("~", 1000.0, 0.01))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[8:12]))).To(BeNumerically("~", 1.5, 0.01))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[12:16]))).To(BeNumerically("~", -3.5, 0.01))
		})

		It("rejects an invalid channel", func() {
			band := &dspi.EQBand{
				Channel:       99,
				Band:          0,
				Type:          dspi.FilterTypePeaking,
				Freq:          1000,
				QualityFactor: 1.0,
				Gain:          0,
			}

			err := dev.SetEQBand(band)
			Expect(err).To(MatchError(ContainSubstring("channel 99 out of range")))
		})

		It("rejects a band above the active range", func() {
			band := &dspi.EQBand{
				Channel:       0,
				Band:          10,
				Type:          dspi.FilterTypePeaking,
				Freq:          1000,
				QualityFactor: 1.0,
				Gain:          0,
			}

			err := dev.SetEQBand(band)
			Expect(err).To(MatchError(ContainSubstring("band 10 out of range (0-9)")))
		})

		It("rejects non-positive frequency", func() {
			band := &dspi.EQBand{
				Channel:       0,
				Band:          0,
				Type:          dspi.FilterTypePeaking,
				Freq:          0,
				QualityFactor: 1.0,
				Gain:          0,
			}

			err := dev.SetEQBand(band)
			Expect(err).To(MatchError(ContainSubstring("frequency must be > 0")))
		})

		It("allows first-order filters without a Q value", func() {
			for _, typ := range []dspi.FilterType{
				dspi.FilterTypeAllPass1,
				dspi.FilterTypeLowShelf1,
				dspi.FilterTypeHighShelf1,
				dspi.FilterTypeLowPass1,
				dspi.FilterTypeHighPass1,
			} {
				band := &dspi.EQBand{
					Channel:       0,
					Band:          0,
					Type:          typ,
					Freq:          100,
					QualityFactor: 0, // ignored by the firmware for first-order types
					Gain:          0,
				}

				err := dev.SetEQBand(band)
				Expect(err).ToNot(HaveOccurred(), "type %s", typ)
			}
		})

		It("sends the first-order low/high pass wire type values", func() {
			band := &dspi.EQBand{
				Channel: 0,
				Band:    0,
				Type:    dspi.FilterTypeLowPass1,
				Freq:    80,
			}

			err := dev.SetEQBand(band)
			Expect(err).ToNot(HaveOccurred())
			data := mock.CapturedRequests[1].Data
			Expect(data[2]).To(Equal(byte(12))) // FILTER_LOWPASS1
		})

		It("rejects non-positive Q", func() {
			band := &dspi.EQBand{
				Channel:       0,
				Band:          0,
				Type:          dspi.FilterTypePeaking,
				Freq:          1000,
				QualityFactor: 0,
				Gain:          0,
			}

			err := dev.SetEQBand(band)
			Expect(err).To(MatchError(ContainSubstring("quality factor must be > 0")))
		})

		It("allows flat filters without frequency or Q validation", func() {
			band := &dspi.EQBand{
				Channel:       0,
				Band:          0,
				Type:          dspi.FilterTypeFlat,
				Freq:          0,
				QualityFactor: 0,
				Gain:          0,
			}

			err := dev.SetEQBand(band)
			Expect(err).ToNot(HaveOccurred())
		})

		It("sends an 18-byte packet for Linkwitz Transform with Qp sidecar", func() {
			band := &dspi.EQBand{
				Channel:       0,
				Band:          0,
				Type:          dspi.FilterTypeLinkwitzTransform,
				Freq:          50,
				QualityFactor: 0.5,
				Gain:          25,
				Qp:            0.707,
			}

			err := dev.SetEQBand(band)
			Expect(err).ToNot(HaveOccurred())

			data := mock.CapturedRequests[1].Data
			Expect(data).To(HaveLen(18))
			Expect(data[2]).To(Equal(byte(11)))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))).To(BeNumerically("~", 50.0, 0.01))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[8:12]))).To(BeNumerically("~", 0.5, 0.01))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(data[12:16]))).To(BeNumerically("~", 25.0, 0.01))
			Expect(binary.LittleEndian.Uint16(data[16:18])).To(Equal(uint16(362))) // round(0.707 * 512)
		})
	})

	Describe("GetEQBand", func() {
		It("parses the response correctly", func() {
			band, err := dev.GetEQBand(2, 5)
			Expect(err).ToNot(HaveOccurred())
			Expect(band.Channel).To(Equal(2))
			Expect(band.Band).To(Equal(5))
			Expect(band.Type).To(Equal(dspi.FilterTypePeaking))
			Expect(band.Freq).To(BeNumerically("~", 1000.0, 0.01))
			Expect(band.QualityFactor).To(BeNumerically("~", 1.5, 0.01))
			Expect(band.Gain).To(BeNumerically("~", -3.5, 0.01))
		})

		It("sends the correct requests", func() {
			_, _ = dev.GetEQBand(2, 5)
			Expect(mock.CapturedRequests).To(HaveLen(5))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetAllParamsChunk)))
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqGetEQParam)))
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(0x0228))) // band 5 << 3
			Expect(mock.CapturedRequests[2].WValue).To(Equal(uint16(0x0229)))
			Expect(mock.CapturedRequests[3].WValue).To(Equal(uint16(0x022A)))
			Expect(mock.CapturedRequests[4].WValue).To(Equal(uint16(0x022B)))
		})

		It("reads Qp for Linkwitz Transform bands", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetEQParam), 0x0228, 2}] = func() []byte {
				b := make([]byte, 4)
				binary.LittleEndian.PutUint32(b, uint32(dspi.FilterTypeLinkwitzTransform))
				return b
			}()
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetEQParam), 0x022D, 2}] = func() []byte {
				b := make([]byte, 4)
				binary.LittleEndian.PutUint32(b, 362) // 0.707 * 512
				return b
			}()

			band, err := dev.GetEQBand(2, 5)
			Expect(err).ToNot(HaveOccurred())
			Expect(band.Type).To(Equal(dspi.FilterTypeLinkwitzTransform))
			Expect(band.Qp).To(BeNumerically("~", 0.707, 0.001))
		})
	})

	Describe("SetMasterEQBypass", func() {
		It("sends the correct value", func() {
			err := dev.SetMasterEQBypass(true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetBypass)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{1}))
		})
	})

	Describe("GetMasterEQBypass", func() {
		It("returns true when bypass is on", func() {
			bypass, err := dev.GetMasterEQBypass()
			Expect(err).ToNot(HaveOccurred())
			Expect(bypass).To(BeTrue())
		})
	})

	Describe("SetBandBypass", func() {
		It("sends the correct wValue for channel 2, band 3", func() {
			err := dev.SetBandBypass(2, 3, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetAllParamsChunk)))
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqSetBandBypass)))
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(0x0203)))
			Expect(mock.CapturedRequests[1].Data).To(Equal([]byte{1}))
		})

		It("sends 0 for bypass disabled", func() {
			err := dev.SetBandBypass(2, 3, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[1].Data).To(Equal([]byte{0}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetBandBypass(2, 3, true)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetBandBypass", func() {
		It("returns true when bypass is on", func() {
			bypass, err := dev.GetBandBypass(2, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(bypass).To(BeTrue())
		})

		It("sends the correct wValue", func() {
			_, _ = dev.GetBandBypass(2, 3)
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetAllParamsChunk)))
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqGetBandBypass)))
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(0x0203)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetBandBypass(2, 3)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("FilterType", func() {
		It("formats known types", func() {
			Expect(dspi.FilterTypeFlat.String()).To(Equal("flat"))
			Expect(dspi.FilterTypePeaking.String()).To(Equal("peak"))
			Expect(dspi.FilterTypeLowShelf.String()).To(Equal("lowshelf"))
			Expect(dspi.FilterTypeHighShelf.String()).To(Equal("highshelf"))
			Expect(dspi.FilterTypeLowPass.String()).To(Equal("lowpass"))
			Expect(dspi.FilterTypeHighPass.String()).To(Equal("highpass"))
			Expect(dspi.FilterTypeNotch.String()).To(Equal("notch"))
			Expect(dspi.FilterTypeAllPass.String()).To(Equal("allpass"))
			Expect(dspi.FilterTypeAllPass1.String()).To(Equal("allpass1"))
			Expect(dspi.FilterTypeLowShelf1.String()).To(Equal("lowshelf1"))
			Expect(dspi.FilterTypeHighShelf1.String()).To(Equal("highshelf1"))
			Expect(dspi.FilterTypeLowPass1.String()).To(Equal("lowpass1"))
			Expect(dspi.FilterTypeHighPass1.String()).To(Equal("highpass1"))
		})

		It("parses known type strings", func() {
			t, err := dspi.ParseFilterType("peak")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypePeaking))

			t, err = dspi.ParseFilterType("lowshelf")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeLowShelf))

			t, err = dspi.ParseFilterType("highpass")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeHighPass))

			t, err = dspi.ParseFilterType("notch")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeNotch))

			t, err = dspi.ParseFilterType("allpass")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeAllPass))

			t, err = dspi.ParseFilterType("lowshelf1")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeLowShelf1))

			t, err = dspi.ParseFilterType("lowpass1")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeLowPass1))

			t, err = dspi.ParseFilterType("highpass1")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeHighPass1))

			t, err = dspi.ParseFilterType("linkwitz")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.FilterTypeLinkwitzTransform))
		})

		It("rejects unknown type strings", func() {
			_, err := dspi.ParseFilterType("bogus")
			Expect(err).To(MatchError(ContainSubstring("unknown filter type")))
		})
	})
})
