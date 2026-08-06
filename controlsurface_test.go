package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Control Surfaces", func() {
	Describe("CsBinding wire layout", func() {
		It("encodes the 24-byte binding at the firmware offsets", func() {
			b := &dspi.CsBinding{
				Type:     dspi.CsTypeButton,
				Noun:     dspi.CsNounUserMute,
				Action:   dspi.CsActToggle,
				Flags:    dspi.CsFlagInvert,
				GPIO:     [2]uint8{26, dspi.CsGPIOUnused},
				Event:    dspi.CsEventPress,
				Target:   0,
				Index:    0,
				Value:    0,
				Step:     0,
				RangeMin: 0,
				RangeMax: 0,
			}

			raw := b.Encode()
			Expect(raw).To(HaveLen(24))
			Expect(raw[0]).To(Equal(byte(dspi.CsTypeButton)))
			Expect(raw[1]).To(Equal(byte(dspi.CsNounUserMute)))
			Expect(raw[2]).To(Equal(byte(dspi.CsActToggle)))
			Expect(raw[3]).To(Equal(byte(dspi.CsFlagInvert)))
			Expect(raw[4]).To(Equal(byte(26)))
			Expect(raw[5]).To(Equal(byte(dspi.CsGPIOUnused)))
			Expect(raw[6]).To(Equal(byte(dspi.CsEventPress)))
		})

		It("round-trips through DecodeCsBinding", func() {
			b := &dspi.CsBinding{
				Type: dspi.CsTypePot, Noun: dspi.CsNounMasterVolume, Action: dspi.CsActAdjust,
				GPIO: [2]uint8{27, dspi.CsGPIOUnused}, Value: 0, Step: 256,
				RangeMin: -1024, RangeMax: 1024,
			}
			decoded, err := dspi.DecodeCsBinding(b.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.Type).To(Equal(uint8(dspi.CsTypePot)))
			Expect(decoded.Noun).To(Equal(uint8(dspi.CsNounMasterVolume)))
			Expect(decoded.Step).To(Equal(int16(256)))
			Expect(decoded.RangeMax).To(Equal(int16(1024)))
		})

		It("rejects short payloads", func() {
			_, err := dspi.DecodeCsBinding(make([]byte, 16))
			Expect(err).To(MatchError(ContainSubstring("too short")))
		})
	})

	Describe("IrCommand wire layout", func() {
		It("encodes the 16-byte command", func() {
			c := &dspi.IrCommand{
				Noun: dspi.CsNounPreset, Action: dspi.CsActSet, Flags: dspi.CsFlagWrap,
				Target: 0, Index: 0, Protocol: dspi.CsIRProtoNEC,
				Value: 1, Step: 0, Code: 0x00FF10EF,
			}

			raw := c.Encode()
			Expect(raw).To(HaveLen(16))
			Expect(raw[0]).To(Equal(byte(dspi.CsNounPreset)))
			Expect(raw[5]).To(Equal(byte(dspi.CsIRProtoNEC)))
			Expect(binary.LittleEndian.Uint32(raw[12:16])).To(Equal(uint32(0x00FF10EF)))

			decoded, err := dspi.DecodeIrCommand(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.Protocol).To(Equal(uint8(dspi.CsIRProtoNEC)))
			Expect(decoded.Code).To(Equal(uint32(0x00FF10EF)))
			Expect(decoded.Value).To(Equal(int16(1)))
		})
	})

	Describe("CsStatusPacket decoding", func() {
		It("parses the 41-byte caps v6 layout", func() {
			raw := make([]byte, 41)
			raw[0] = dspi.CsStatusPending
			raw[1] = 3
			raw[2] = 16
			raw[3] = 1 // dirty
			binary.LittleEndian.PutUint16(raw[4:6], 0x0005)
			raw[21] = dspi.CsStatusInvalidValue // slot_status[15]
			binary.LittleEndian.PutUint16(raw[22:24], 0x0003)
			raw[24] = dspi.CsIrLearnArmed
			raw[25] = dspi.CsStatusInvalidSlot // IR sub-slot 0

			pkt, err := dspi.DecodeCsStatus(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pkt.LastStatus).To(Equal(uint8(dspi.CsStatusPending)))
			Expect(pkt.LastSlot).To(Equal(uint8(3)))
			Expect(pkt.Dirty).To(BeTrue())
			Expect(pkt.ActiveMask).To(Equal(uint16(0x0005)))
			Expect(pkt.SlotStatus[15]).To(Equal(uint8(dspi.CsStatusInvalidValue)))
			Expect(pkt.IRActiveMask).To(Equal(uint16(0x0003)))
			Expect(pkt.IRLearnState).To(Equal(uint8(dspi.CsIrLearnArmed)))
			Expect(pkt.IRCmdStatus[0]).To(Equal(uint8(dspi.CsStatusInvalidSlot)))
		})

		It("handles the pre-v6 32-byte short read", func() {
			raw := make([]byte, 32)
			raw[0] = dspi.CsStatusBusy
			pkt, err := dspi.DecodeCsStatus(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pkt.LastStatus).To(Equal(uint8(dspi.CsStatusBusy)))
			Expect(pkt.IRLearnState).To(Equal(uint8(0)))
		})

		It("rejects very short payloads", func() {
			_, err := dspi.DecodeCsStatus(make([]byte, 3))
			Expect(err).To(MatchError(ContainSubstring("too short")))
		})
	})

	Describe("Caps decoding", func() {
		It("parses the 40-byte caps header and type table", func() {
			raw := make([]byte, 40)
			raw[0] = 7 // caps version
			raw[1] = 16
			raw[2] = 8 // type count
			raw[3] = 51
			binary.LittleEndian.PutUint16(raw[8:10], 0x00FF) // button actions
			raw[10] = 1                                      // pin count
			raw[11] = 0                                      // pin class any
			raw[36] = 16                                     // max IR commands

			h, err := dspi.DecodeCsCaps(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(h.CapsVersion).To(Equal(uint8(7)))
			Expect(h.TypeCount).To(Equal(uint8(8)))
			Expect(h.NounCount).To(Equal(uint8(51)))
			Expect(h.MaxIRCommands).To(Equal(uint8(16)))
			Expect(h.Types[dspi.CsTypeButton].Actions).To(Equal(uint16(0x00FF)))
		})

		It("parses a 12-byte noun descriptor", func() {
			raw := make([]byte, 12)
			raw[0] = dspi.CsKindContinuous
			binary.LittleEndian.PutUint16(raw[2:4], 0x0001) // adjust only
			minQ := int16(-32768)
			binary.LittleEndian.PutUint16(raw[4:6], uint16(minQ))
			binary.LittleEndian.PutUint16(raw[6:8], 0)
			raw[8] = dspi.CsUnitDB
			raw[9] = dspi.CsTargetOutputCh
			raw[10] = 9

			desc, err := dspi.DecodeCsNounDesc(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(desc.Kind).To(Equal(uint8(dspi.CsKindContinuous)))
			Expect(desc.Actions).To(Equal(uint16(0x0001)))
			Expect(desc.MinQ).To(Equal(int16(-32768)))
			Expect(desc.Unit).To(Equal(uint8(dspi.CsUnitDB)))
			Expect(desc.TargetKind).To(Equal(uint8(dspi.CsTargetOutputCh)))
			Expect(desc.TargetCount).To(Equal(uint8(9)))
		})
	})

	Describe("Device methods", func() {
		It("sets a binding with the 24-byte payload and validates the slot", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetCsBinding), 3, 2}: {},
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.SetCsBinding(3, &dspi.CsBinding{Type: dspi.CsTypeButton, Noun: dspi.CsNounUserMute, Action: dspi.CsActToggle})
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetCsBinding)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(3)))
			Expect(mock.CapturedRequests[0].Data).To(HaveLen(24))

			err = dev.SetCsBinding(16, &dspi.CsBinding{})
			Expect(err).To(MatchError(ContainSubstring("out of range")))
		})

		It("gets a binding", func() {
			b := &dspi.CsBinding{Type: dspi.CsTypeEncoder, Noun: dspi.CsNounMasterVolume, Action: dspi.CsActStep, GPIO: [2]uint8{2, 3}}
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetCsBinding), 5, 2}: b.Encode(),
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			got, err := dev.GetCsBinding(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(got.Type).To(Equal(uint8(dspi.CsTypeEncoder)))
			Expect(got.GPIO).To(Equal([2]uint8{2, 3}))
		})

		It("reads caps, status, and noun descriptors", func() {
			raw := make([]byte, 40)
			raw[0] = 7
			raw[2] = 8
			status := make([]byte, 41)
			noun := make([]byte, 12)
			noun[0] = dspi.CsKindEnum

			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetCsCaps), dspi.CsCapsAll, 2}: raw,
				{uint16(dspi.ReqGetCsStatus), 0, 2}:            status,
				{uint16(dspi.ReqGetCsCaps), 23, 2}:             noun,
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			caps, err := dev.GetCsCaps()
			Expect(err).ToNot(HaveOccurred())
			Expect(caps.CapsVersion).To(Equal(uint8(7)))

			st, err := dev.GetCsStatus()
			Expect(err).ToNot(HaveOccurred())
			Expect(st.LastStatus).To(Equal(uint8(0)))

			desc, err := dev.GetCsNounDesc(23)
			Expect(err).ToNot(HaveOccurred())
			Expect(desc.Kind).To(Equal(uint8(dspi.CsKindEnum)))
		})

		It("sets and gets slot names", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetCsName), 2, 2}: {},
				{uint16(dspi.ReqGetCsName), 2, 2}: []byte("Sub Level\x00\x00\x00"),
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.SetCsName(2, "Sub Level")
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte("Sub Level")))

			name, err := dev.GetCsName(2)
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("Sub Level"))
		})

		It("rejects over-long names", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			err := dev.SetCsName(0, "this name is far too long for the 32-byte slot")
			Expect(err).To(MatchError(ContainSubstring("name too long")))
		})

		It("sets and gets IR commands", func() {
			c := &dspi.IrCommand{Noun: dspi.CsNounPreset, Action: dspi.CsActSet, Protocol: dspi.CsIRProtoNEC, Code: 0x00FF10EF}
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetCsIrCommand), 4, 2}: {},
				{uint16(dspi.ReqGetCsIrCommand), 4, 2}: c.Encode(),
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.SetCsIrCommand(4, c)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(4)))

			got, err := dev.GetCsIrCommand(4)
			Expect(err).ToNot(HaveOccurred())
			Expect(got.Code).To(Equal(uint32(0x00FF10EF)))

			err = dev.SetCsIrCommand(16, c)
			Expect(err).To(MatchError(ContainSubstring("out of range")))
		})

		It("arms, reads, and cancels IR learn", func() {
			raw := make([]byte, 8)
			raw[0] = dspi.CsIrLearnDone
			raw[1] = dspi.CsIRProtoRC5
			binary.LittleEndian.PutUint32(raw[4:8], 0x1C)

			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqCsIrLearn), 1, 2}: {1},
				{uint16(dspi.ReqCsIrLearn), 2, 2}: raw,
				{uint16(dspi.ReqCsIrLearn), 0, 2}: {1},
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.CsIrLearnArm()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(1)))

			res, err := dev.CsIrLearnRead()
			Expect(err).ToNot(HaveOccurred())
			Expect(res.State).To(Equal(uint8(dspi.CsIrLearnDone)))
			Expect(res.Protocol).To(Equal(uint8(dspi.CsIRProtoRC5)))
			Expect(res.Code).To(Equal(uint32(0x1C)))

			err = dev.CsIrLearnCancel()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[2].WValue).To(Equal(uint16(0)))
		})

		It("saves and reverts via the ack byte", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqCsSave), 0, 2}:   {1},
				{uint16(dspi.ReqCsRevert), 0, 2}: {1},
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.CsSave()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqCsSave)))

			err = dev.CsRevert()
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[1].BRequest).To(Equal(uint8(dspi.ReqCsRevert)))
		})
	})

	Describe("Name helpers", func() {
		It("names types, nouns, and statuses", func() {
			Expect(dspi.CsTypeName(dspi.CsTypeEncoder)).To(Equal("encoder"))
			Expect(dspi.CsTypeName(dspi.CsTypeIR)).To(Equal("ir"))
			Expect(dspi.CsNounName(dspi.CsNounLoudnessSPL)).To(Equal("loudness-spl"))
			Expect(dspi.CsNounName(dspi.CsNounLoudnessIntensity)).To(Equal("loudness-intensity"))
			Expect(dspi.CsStatusName(dspi.CsStatusPending)).To(Equal("pending"))
			Expect(dspi.CsStatusName(dspi.PinConfigSuccess)).To(Equal("ok"))
		})

		It("parses types, nouns, and actions", func() {
			t, err := dspi.ParseCsType("button")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(uint8(dspi.CsTypeButton)))

			n, err := dspi.ParseCsNoun("filter-freq")
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(uint8(dspi.CsNounFilterFreq)))

			a, err := dspi.ParseCsAction("toggle")
			Expect(err).ToNot(HaveOccurred())
			Expect(a).To(Equal(uint8(dspi.CsActToggle)))

			_, err = dspi.ParseCsNoun("bogus")
			Expect(err).To(MatchError(ContainSubstring("unknown CS noun")))
		})
	})
})
