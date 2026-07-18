package dspi_test

import (
	"encoding/binary"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Siggen", func() {
	var mock *mockControlTransfer
	var dev *dspi.Device

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSiggenSetConfig), 0, 2}: {},
				{uint16(dspi.ReqSiggenGetCaps), 0xFFFF, 2}: {
					0x01, 0x0F, 0x09, 0x10, 0xFF, 0x01, 0x00, 0x00,
				},
				{uint16(dspi.ReqSiggenGetCaps), 0, 2}: sineTypeDesc(),
				{uint16(dspi.ReqSiggenGetConfig), 0, 2}: {
					0x01, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0xA0, 0xC1, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7A, 0x44,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
				},
				{uint16(dspi.ReqSiggenGetStatus), 0, 2}: {
					0x01, 0x02, 0x00, 0xFF, 0xE8, 0x03, 0x00, 0x00,
					0x05, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x7A, 0x44,
				},
				{uint16(dspi.ReqSiggenControl), uint16(dspi.SiggenCtlStart), 2}:   {0x01},
				{uint16(dspi.ReqSiggenControl), uint16(dspi.SiggenCtlStop), 2}:    {0x01},
				{uint16(dspi.ReqSiggenControl), uint16(dspi.SiggenCtlStopNow), 2}: {0x01},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("ParseSiggenType", func() {
		It("parses numeric IDs", func() {
			t, err := dspi.ParseSiggenType("14")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.SiggenTypeChannelID))
		})

		It("parses firmware short names", func() {
			t, err := dspi.ParseSiggenType("swp-log")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.SiggenTypeSweepLog))
		})

		It("parses friendly aliases", func() {
			t, err := dspi.ParseSiggenType("tone-pair")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).To(Equal(dspi.SiggenTypeTonePair))
		})

		It("rejects invalid names", func() {
			_, err := dspi.ParseSiggenType("not-a-type")
			Expect(err).To(HaveOccurred())
		})

		It("rejects out-of-range IDs", func() {
			_, err := dspi.ParseSiggenType("99")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SetSiggenConfig", func() {
		It("sends a 36-byte payload", func() {
			cfg := &dspi.SiggenConfig{
				SignalType:  dspi.SiggenTypeSine,
				ChannelMask: 0x03,
				LevelDB:     -20,
				P1:          1000,
			}

			err := dev.SetSiggenConfig(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			req := mock.CapturedRequests[0]
			Expect(req.BmRequestType).To(Equal(uint8(0x41)))
			Expect(req.BRequest).To(Equal(uint8(dspi.ReqSiggenSetConfig)))
			Expect(req.WValue).To(Equal(uint16(0)))
			Expect(req.WIndex).To(Equal(uint16(2)))
			Expect(req.Data).To(HaveLen(36))
			Expect(req.Data[0]).To(Equal(uint8(1)))
			Expect(req.Data[1]).To(Equal(uint8(dspi.SiggenTypeSine)))
			Expect(binary.LittleEndian.Uint16(req.Data[2:4])).To(Equal(uint16(0x03)))
		})
	})

	Describe("GetSiggenConfig", func() {
		It("decodes the 36-byte response", func() {
			cfg, err := dev.GetSiggenConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.SignalType).To(Equal(dspi.SiggenTypeSine))
			Expect(cfg.ChannelMask).To(Equal(uint16(0x03)))
			Expect(cfg.LevelDB).To(BeNumerically("~", -20.0, 0.001))
			Expect(cfg.P1).To(BeNumerically("~", 1000.0, 0.001))
		})
	})

	Describe("SiggenStart / Stop / StopNow", func() {
		It("sends a control request with the correct action", func() {
			Expect(dev.SiggenStart()).To(Succeed())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSiggenControl)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(dspi.SiggenCtlStart)))
		})
	})

	Describe("GetSiggenStatus", func() {
		It("decodes the 16-byte response", func() {
			status, err := dev.GetSiggenStatus()
			Expect(err).ToNot(HaveOccurred())
			Expect(status.State).To(Equal(dspi.SiggenStateRun))
			Expect(status.SignalType).To(Equal(dspi.SiggenTypeSine))
			Expect(status.ElapsedMs).To(Equal(uint32(1000)))
			Expect(status.CyclesDone).To(Equal(uint16(5)))
			Expect(status.StopReason).To(Equal(dspi.SiggenStopReasonHost))
			Expect(status.ActiveChannel).To(Equal(-1))
			Expect(status.CurrentFreq).To(BeNumerically("~", 1000.0, 0.001))
		})
	})

	Describe("GetSiggenCaps", func() {
		It("decodes the 8-byte header", func() {
			caps, err := dev.GetSiggenCaps()
			Expect(err).ToNot(HaveOccurred())
			Expect(caps.TypeCount).To(Equal(uint8(15)))
			Expect(caps.OutputChannels).To(Equal(uint8(9)))
			Expect(caps.MultitoneMax).To(Equal(uint8(16)))
			Expect(caps.ValidChannelMask).To(Equal(uint16(0x01FF)))
		})
	})

	Describe("GetSiggenTypeDesc", func() {
		It("decodes the 62-byte descriptor", func() {
			desc, err := dev.GetSiggenTypeDesc(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(desc.ID).To(Equal(dspi.SiggenTypeSine))
			Expect(desc.Name).To(Equal("sine"))
			Expect(desc.TimingModel).To(Equal(dspi.SiggenTimingContinuous))
			Expect(desc.Params[0].Semantic).To(Equal(dspi.SiggenParamFreqHz))
			Expect(desc.Params[0].Min).To(BeNumerically("~", 1.0, 0.001))
			Expect(desc.Params[0].Max).To(BeNumerically("~", 40000.0, 0.001))
			Expect(desc.Params[0].Default).To(BeNumerically("~", 1000.0, 0.001))
		})
	})

	Describe("round-trip encoding", func() {
		It("preserves a full config", func() {
			original := &dspi.SiggenConfig{
				SignalType:  dspi.SiggenTypeSweepStep,
				ChannelMask: 0x05,
				InvertMask:  0x04,
				Flags:       dspi.SiggenFlagRaw | dspi.SiggenFlagWalk,
				LevelDB:     -30.5,
				DurationMs:  5000,
				Repeat:      3,
				GapMs:       100,
				P1:          20,
				P2:          20000,
				P3:          6,
				P4:          250,
			}

			mock.ReturnData[[3]uint16{uint16(dspi.ReqSiggenSetConfig), 0, 2}] = []byte{}
			Expect(dev.SetSiggenConfig(original)).To(Succeed())
			Expect(mock.CapturedRequests).To(HaveLen(1))

			wire := append([]byte{}, mock.CapturedRequests[0].Data...)
			Expect(wire).To(HaveLen(36))

			// Feed the captured wire back as the GET_CONFIG response to verify
			// the full encode/decode path.
			mock.ReturnData[[3]uint16{uint16(dspi.ReqSiggenGetConfig), 0, 2}] = wire

			decoded, err := dev.GetSiggenConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.SignalType).To(Equal(original.SignalType))
			Expect(decoded.ChannelMask).To(Equal(original.ChannelMask))
			Expect(decoded.InvertMask).To(Equal(original.InvertMask))
			Expect(decoded.Flags).To(Equal(original.Flags))
			Expect(decoded.LevelDB).To(BeNumerically("~", original.LevelDB, 0.001))
			Expect(decoded.DurationMs).To(Equal(original.DurationMs))
			Expect(decoded.Repeat).To(Equal(original.Repeat))
			Expect(decoded.GapMs).To(Equal(original.GapMs))
			Expect(decoded.P1).To(BeNumerically("~", original.P1, 0.001))
			Expect(decoded.P2).To(BeNumerically("~", original.P2, 0.001))
			Expect(decoded.P3).To(BeNumerically("~", original.P3, 0.001))
			Expect(decoded.P4).To(BeNumerically("~", original.P4, 0.001))
		})
	})
})

// sineTypeDesc returns a 62-byte descriptor for the sine type as the firmware
// advertises it.
func sineTypeDesc() []byte {
	buf := make([]byte, 62)
	buf[0] = 0x00 // id = SIGGEN_SINE
	copy(buf[1:9], "sine")
	buf[9] = byte(dspi.SiggenTimingContinuous)

	param := func(off int, semantic uint8, min, max, def float64) {
		buf[off] = semantic
		binary.LittleEndian.PutUint32(buf[off+1:off+5], math.Float32bits(float32(min)))
		binary.LittleEndian.PutUint32(buf[off+5:off+9], math.Float32bits(float32(max)))
		binary.LittleEndian.PutUint32(buf[off+9:off+13], math.Float32bits(float32(def)))
	}

	param(10, byte(dspi.SiggenParamFreqHz), 1, 40000, 1000)
	param(23, byte(dspi.SiggenParamUnused), 0, 0, 0)
	param(36, byte(dspi.SiggenParamUnused), 0, 0, 0)
	param(49, byte(dspi.SiggenParamUnused), 0, 0, 0)

	return buf
}
