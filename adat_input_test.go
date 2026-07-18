package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("ADAT Input", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetAdatInputEnable), 0, 2}:    {0x01},
				{uint16(dspi.ReqSetAdatInputEnable), 1, 2}:    {0x00},
				{uint16(dspi.ReqGetAdatInputPin), 0, 2}:       {0x14},
				{uint16(dspi.ReqSetAdatInputPin), 0x14, 2}:    {0x00},
				{uint16(dspi.ReqGetAdatInputClockMode), 0, 2}: {0x01},
				{uint16(dspi.ReqSetAdatInputClockMode), 1, 2}: {0x00},
				{uint16(dspi.ReqGetAdatInputStatus), 0, 2}: func() []byte {
					b := make([]byte, 20)
					b[0] = 3  // locked
					b[1] = 1  // slave
					b[2] = 1  // enabled
					b[3] = 20 // pin
					b[4] = 1  // rate_ok
					binary.LittleEndian.PutUint32(b[12:16], 48000)
					binary.LittleEndian.PutUint32(b[16:20], 48000)
					return b
				}(),
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("GetAdatInputEnable", func() {
		It("returns the enable state", func() {
			enabled, err := dev.GetAdatInputEnable()
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(BeTrue())
		})
	})

	Describe("SetAdatInputEnable", func() {
		It("sends the enable value in wValue", func() {
			err := dev.SetAdatInputEnable(true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAdatInputEnable)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(1)))
		})
	})

	Describe("GetAdatInputPin", func() {
		It("returns the configured pin", func() {
			pin, err := dev.GetAdatInputPin()
			Expect(err).ToNot(HaveOccurred())
			Expect(pin).To(Equal(uint8(20)))
		})
	})

	Describe("SetAdatInputPin", func() {
		It("sends the pin in wValue", func() {
			err := dev.SetAdatInputPin(20)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAdatInputPin)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(20)))
		})
	})

	Describe("GetAdatInputClockMode", func() {
		It("returns the clock mode", func() {
			mode, err := dev.GetAdatInputClockMode()
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(dspi.AdatClockModeSlave))
		})
	})

	Describe("SetAdatInputClockMode", func() {
		It("sends the mode in wValue", func() {
			err := dev.SetAdatInputClockMode(dspi.AdatClockModeSlave)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAdatInputClockMode)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(1)))
		})
	})

	Describe("GetAdatInputStatus", func() {
		It("parses the 20-byte status packet", func() {
			status, err := dev.GetAdatInputStatus()
			Expect(err).ToNot(HaveOccurred())
			Expect(status.State).To(Equal(dspi.AdatInputLocked))
			Expect(status.ClockMode).To(Equal(dspi.AdatClockModeSlave))
			Expect(status.Enabled).To(BeTrue())
			Expect(status.Pin).To(Equal(uint8(20)))
			Expect(status.RateOK).To(BeTrue())
			Expect(status.DetectedRate).To(Equal(uint32(48000)))
			Expect(status.MeasuredHz).To(Equal(uint32(48000)))
		})
	})

	Describe("AdatInputState", func() {
		It("returns human-readable state names", func() {
			Expect(dspi.AdatInputLocked.String()).To(Equal("locked"))
			Expect(dspi.AdatInputState(99).String()).To(Equal("unknown(99)"))
		})
	})
})
