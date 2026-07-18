package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Input", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetInputSource), 0, 2}: {0x02},
				{uint16(dspi.ReqSetInputSource), 0, 2}: {},
				{uint16(dspi.ReqGetInputRate), 0, 2}: func() []byte {
					b := make([]byte, 8)
					binary.LittleEndian.PutUint32(b[0:4], 48000)
					binary.LittleEndian.PutUint32(b[4:8], 48000)

					return b
				}(),
				{uint16(dspi.ReqSetInputRate), 0, 2}: {},
				{uint16(dspi.ReqGetI2SRxPin), 0, 2}:  {0x04},
				{uint16(dspi.ReqSetI2SRxPin), 4, 2}:  {0x00},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("SetInputSource", func() {
		It("sends I2S source", func() {
			err := dev.SetInputSource(dspi.InputSourceI2S)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetInputSource)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x02}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetInputSource(dspi.InputSourceI2S)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetInputSource", func() {
		It("returns the active source", func() {
			src, err := dev.GetInputSource()
			Expect(err).ToNot(HaveOccurred())
			Expect(src).To(Equal(dspi.InputSourceI2S))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetInputSource()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetInputRate", func() {
		It("sends 48000 as little-endian uint32", func() {
			err := dev.SetInputRate(48000)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x80, 0xBB, 0x00, 0x00}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetInputRate(48000)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetInputRate", func() {
		It("returns both rates", func() {
			rate, err := dev.GetInputRate()
			Expect(err).ToNot(HaveOccurred())
			Expect(rate.PipelineHz).To(Equal(uint32(48000)))
			Expect(rate.SelectedHz).To(Equal(uint32(48000)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetInputRate()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetI2SRxPin", func() {
		It("sends the pin in wValue and checks status", func() {
			err := dev.SetI2SRxPin(4)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetI2SRxPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(4)))
		})

		It("returns an error on non-zero status", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqSetI2SRxPin), 4, 2}] = []byte{0x02}
			err := dev.SetI2SRxPin(4)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("status 0x02"))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetI2SRxPin(4)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetI2SRxPin", func() {
		It("returns the pin number", func() {
			pin, err := dev.GetI2SRxPin()
			Expect(err).ToNot(HaveOccurred())
			Expect(pin).To(Equal(4))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetI2SRxPin()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("InputSourceName", func() {
		It("returns names for known sources", func() {
			Expect(dspi.InputSourceName(0)).To(Equal("USB"))
			Expect(dspi.InputSourceName(1)).To(Equal("S/PDIF"))
			Expect(dspi.InputSourceName(2)).To(Equal("I2S"))
			Expect(dspi.InputSourceName(3)).To(Equal("ADAT"))
			Expect(dspi.InputSourceName(4)).To(Equal("S/PDIF 2"))
			Expect(dspi.InputSourceName(5)).To(Equal("S/PDIF 3"))
			Expect(dspi.InputSourceName(99)).To(Equal("Unknown(99)"))
		})
	})

	Describe("AdatClockModeName", func() {
		It("returns names for known modes", func() {
			Expect(dspi.AdatClockModeName(0)).To(Equal("master"))
			Expect(dspi.AdatClockModeName(1)).To(Equal("slave"))
			Expect(dspi.AdatClockModeName(99)).To(Equal("unknown(99)"))
		})
	})

	Describe("I2SInputRateHz", func() {
		It("converts enums correctly", func() {
			Expect(dspi.I2SInputRateHz(0)).To(Equal(uint32(44100)))
			Expect(dspi.I2SInputRateHz(1)).To(Equal(uint32(48000)))
			Expect(dspi.I2SInputRateHz(2)).To(Equal(uint32(96000)))
			Expect(dspi.I2SInputRateHz(99)).To(Equal(uint32(0)))
		})
	})

	Describe("I2SInputRateEnum", func() {
		It("converts Hz correctly", func() {
			Expect(dspi.I2SInputRateEnum(44100)).To(Equal(0))
			Expect(dspi.I2SInputRateEnum(48000)).To(Equal(1))
			Expect(dspi.I2SInputRateEnum(96000)).To(Equal(2))
			Expect(dspi.I2SInputRateEnum(88200)).To(Equal(-1))
		})
	})
})
