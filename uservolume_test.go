package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("User mute", func() {
	It("sets the vendor mute flag", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetUserMute), 0, 2}: {},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetUserMute(true)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetUserMute)))
		Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{1}))

		err = dev.SetUserMute(false)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[1].Data).To(Equal([]byte{0}))
	})

	It("gets the vendor mute flag", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetUserMute), 0, 2}: {1},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		muted, err := dev.GetUserMute()
		Expect(err).ToNot(HaveOccurred())
		Expect(muted).To(BeTrue())
	})
})

var _ = Describe("I2S clock features", func() {
	It("sets and gets the I2S clock mode", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetI2SClockMode), 0, 2}: {},
			{uint16(dspi.ReqGetI2SClockMode), 0, 2}: {1},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetI2SClockMode(dspi.I2SClockModeSlave)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetI2SClockMode)))
		Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{1}))

		mode, err := dev.GetI2SClockMode()
		Expect(err).ToNot(HaveOccurred())
		Expect(mode).To(Equal(dspi.I2SClockModeSlave))
	})

	It("decodes the 16-byte I2S slave status packet", func() {
		raw := make([]byte, 16)
		raw[0] = 3 // locked
		raw[1] = 1 // slave mode
		raw[2] = 7 // lock count
		raw[3] = 1 // loss count
		binary.LittleEndian.PutUint32(raw[4:8], 48000)
		binary.LittleEndian.PutUint32(raw[8:12], 48013)
		raw[12] = 0 // slip count

		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqGetI2SSlaveStatus), 0, 2}: raw,
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		st, err := dev.GetI2sSlaveStatus()
		Expect(err).ToNot(HaveOccurred())
		Expect(st.State).To(Equal(dspi.I2sSlaveLocked))
		Expect(st.State.String()).To(Equal("locked"))
		Expect(st.ClockMode).To(Equal(1))
		Expect(st.LockCount).To(Equal(uint8(7)))
		Expect(st.LossCount).To(Equal(uint8(1)))
		Expect(st.DetectedRate).To(Equal(uint32(48000)))
		Expect(st.MeasuredHz).To(Equal(uint32(48013)))
	})

	It("sets and gets the I2S clock-pin mode", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetI2SClockPinMode), 1, 2}: {dspi.PinConfigSuccess},
			{uint16(dspi.ReqGetI2SClockPinMode), 0, 2}: {1},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetI2SClockPinMode(dspi.I2SClockPinModeSplit)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(1)))

		mode, err := dev.GetI2SClockPinMode()
		Expect(err).ToNot(HaveOccurred())
		Expect(mode).To(Equal(dspi.I2SClockPinModeSplit))
	})

	It("sets and gets the slave BCK pin via the role byte", func() {
		mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
			{uint16(dspi.ReqSetI2SBckPin), 0x011A, 2}: {dspi.PinConfigSuccess},
			{uint16(dspi.ReqGetI2SBckPin), 1, 2}:      {26},
		}}
		dev := newTestDevice(mock, dspi.PlatformRP2350)

		err := dev.SetI2SBckPinRole(dspi.I2SBckRoleSlave, 26)
		Expect(err).ToNot(HaveOccurred())
		Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0x011A)))

		pin, err := dev.GetI2SBckPinRole(dspi.I2SBckRoleSlave)
		Expect(err).ToNot(HaveOccurred())
		Expect(pin).To(Equal(26))
	})
})
