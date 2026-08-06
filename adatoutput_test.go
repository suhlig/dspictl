package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("ADAT bulk output", func() {
	It("enables with wValue 1 and checks the status byte", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetAdatOutputEnable), 1, 2}: {dspi.PinConfigSuccess},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetAdatOutputEnable(true)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAdatOutputEnable)))
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(1)))
	})

	It("returns an error on a non-success status byte", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetAdatOutputEnable), 1, 2}: {dspi.PinConfigInvalidOutput},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2040)

		err := dev.SetAdatOutputEnable(true)
		Expect(err).To(MatchError(ContainSubstring("status 0x03")))
	})

	It("gets the configured enable", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetAdatOutputEnable), 0, 2}: {1},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		enabled, err := dev.GetAdatOutputEnable()
		Expect(err).ToNot(HaveOccurred())
		Expect(enabled).To(BeTrue())
	})

	It("sets and gets the pin", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetAdatOutputPin), 12, 2}: {dspi.PinConfigSuccess},
			{uint16(dspi.ReqGetAdatOutputPin), 0, 2}:  {12},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetAdatOutputPin(12)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(12)))

		pin, err := dev.GetAdatOutputPin()
		Expect(err).ToNot(HaveOccurred())
		Expect(pin).To(Equal(12))
	})

	It("decodes the 8-byte status packet", func() {
		raw := make([]byte, 8)
		raw[0] = 1 // enabled
		raw[1] = 1 // active
		raw[2] = 12
		raw[3] = 1 // rate ok
		binary.LittleEndian.PutUint16(raw[4:6], 3)
		binary.LittleEndian.PutUint16(raw[6:8], 0)

		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetAdatOutputStatus), 0, 2}: raw,
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		st, err := dev.GetAdatOutputStatus()
		Expect(err).ToNot(HaveOccurred())
		Expect(st.Enabled).To(BeTrue())
		Expect(st.Active).To(BeTrue())
		Expect(st.Pin).To(Equal(uint8(12)))
		Expect(st.RateOK).To(BeTrue())
		Expect(st.ResyncCount).To(Equal(uint16(3)))
		Expect(st.SlipCount).To(Equal(uint16(0)))
	})

	It("rejects short status payloads", func() {
		_, err := dspi.DecodeAdatOutputStatus(make([]byte, 4))
		Expect(err).To(MatchError(ContainSubstring("too short")))
	})
})
