package dspi_test

import (
	"encoding/binary"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Upmixer", func() {
	Describe("UpmixConfigPacket wire layout", func() {
		It("encodes the 44-byte packet at the firmware offsets", func() {
			cfg := &dspi.UpmixConfigPacket{
				Enabled:          true,
				CenterMode:       dspi.UpmixCenterModeOff,
				SurroundMode:     dspi.UpmixSurroundModeAdaptive,
				PresenceQ1:       -8, // -4 dB
				StrengthPct:      50,
				CenterWidthPct:   25,
				CorrThresholdPct: 30,
				AttackMs:         10,
				ReleaseMs:        100,
				DetectorHpfHz:    200,
				SurroundDelayMs:  12,
				SurroundHpfHz:    300,
				SurroundLpfHz:    7000,
				DecorrPct:        90,
			}

			raw := cfg.Encode()
			Expect(raw).To(HaveLen(44))
			Expect(raw[0]).To(Equal(byte(1)))
			Expect(raw[1]).To(Equal(byte(2))) // centre OFF
			Expect(raw[2]).To(Equal(byte(2))) // surround adaptive
			Expect(int8(raw[3])).To(Equal(int8(-8)))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(raw[4:8]))).To(BeNumerically("~", 50, 0.001))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(raw[8:12]))).To(BeNumerically("~", 25, 0.001))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(raw[40:44]))).To(BeNumerically("~", 90, 0.001))
		})

		It("round-trips through DecodeUpmixConfig", func() {
			cfg := &dspi.UpmixConfigPacket{
				Enabled:         true,
				CenterMode:      dspi.UpmixCenterModeAdaptive,
				SurroundMode:    dspi.UpmixSurroundModePassive,
				PresenceQ1:      6,
				StrengthPct:     75,
				CenterWidthPct:  10,
				AttackMs:        25,
				SurroundDelayMs: 5,
				DecorrPct:       100,
			}

			decoded, err := dspi.DecodeUpmixConfig(cfg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.Enabled).To(BeTrue())
			Expect(decoded.CenterMode).To(Equal(uint8(dspi.UpmixCenterModeAdaptive)))
			Expect(decoded.SurroundMode).To(Equal(uint8(dspi.UpmixSurroundModePassive)))
			Expect(decoded.PresenceQ1).To(Equal(int8(6)))
			Expect(decoded.StrengthPct).To(BeNumerically("~", 75, 0.001))
			Expect(decoded.DecorrPct).To(BeNumerically("~", 100, 0.001))
		})

		It("rejects short payloads", func() {
			_, err := dspi.DecodeUpmixConfig(make([]byte, 10))
			Expect(err).To(MatchError(ContainSubstring("too short")))
		})

		It("encodes and decodes the presence bell in dB", func() {
			cfg := &dspi.UpmixConfigPacket{}
			cfg.SetPresenceDB(-4.5)
			Expect(cfg.PresenceQ1).To(Equal(int8(-9)))
			Expect(cfg.PresenceDB()).To(BeNumerically("~", float32(-4.5), 0.001))

			cfg.SetPresenceDB(50) // clamped to +12
			Expect(cfg.PresenceQ1).To(Equal(int8(24)))
			Expect(cfg.PresenceDB()).To(BeNumerically("~", float32(12), 0.001))
		})
	})

	Describe("Device methods", func() {
		It("sets the whole config as an OUT transfer with the 44-byte payload", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqUpmixSetConfig), 0, 2}: {},
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			cfg := &dspi.UpmixConfigPacket{Enabled: true, StrengthPct: 60}
			err := dev.SetUpmixConfig(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqUpmixSetConfig)))
			Expect(mock.CapturedRequests[0].Data).To(HaveLen(44))
		})

		It("reads the config", func() {
			cfg := &dspi.UpmixConfigPacket{Enabled: true, CenterMode: 2, StrengthPct: 50}
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqUpmixGetConfig), 0, 2}: cfg.Encode(),
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			got, err := dev.GetUpmixConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(got.Enabled).To(BeTrue())
			Expect(got.CenterMode).To(Equal(uint8(2)))
			Expect(got.StrengthPct).To(BeNumerically("~", float32(50), 0.001))
		})

		It("sets a single parameter as a float payload with wValue = param id", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqUpmixSetParam), dspi.UpmixParamPresence, 2}: {},
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.SetUpmixParam(dspi.UpmixParamPresence, -6)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqUpmixSetParam)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(dspi.UpmixParamPresence)))
			Expect(math.Float32frombits(binary.LittleEndian.Uint32(mock.CapturedRequests[0].Data))).To(BeNumerically("~", -6, 0.001))
		})

		It("rejects out-of-range parameter ids", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			err := dev.SetUpmixParam(dspi.UpmixParamCount, 1)
			Expect(err).To(MatchError(ContainSubstring("out of range")))
		})

		It("gets a single parameter", func() {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, math.Float32bits(12))
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqUpmixGetParam), dspi.UpmixParamSurDelay, 2}: buf,
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			v, err := dev.GetUpmixParam(dspi.UpmixParamSurDelay)
			Expect(err).ToNot(HaveOccurred())
			Expect(v).To(BeNumerically("~", float32(12), 0.001))
		})

		It("decodes the status packet", func() {
			raw := make([]byte, 16)
			raw[0] = 1 // active
			raw[1] = 3 // parked: rate too high
			binary.LittleEndian.PutUint16(raw[2:4], 8192)
			binary.LittleEndian.PutUint16(raw[8:10], 16384)

			st, err := dspi.DecodeUpmixStatus(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(st.Active).To(BeTrue())
			Expect(st.ParkedReason).To(Equal(uint8(dspi.UpmixParkedRateTooHigh)))
			Expect(st.CorrQ14).To(Equal(int16(8192)))
			Expect(st.LsGainQ15).To(Equal(uint16(16384)))
		})

		It("names parked reasons and parameters", func() {
			Expect(dspi.ParkedReasonName(dspi.UpmixParkedNotStereo)).To(Equal("input not stereo"))
			Expect(dspi.UpmixCenterModeName(dspi.UpmixCenterModeOff)).To(Equal("off"))
			Expect(dspi.UpmixSurroundModeName(dspi.UpmixSurroundModeAdaptive)).To(Equal("adaptive"))
			Expect(dspi.UpmixParamName(dspi.UpmixParamCenterWidth)).To(Equal("center-width"))

			p, err := dspi.ParseUpmixParam("presence")
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(Equal(uint16(dspi.UpmixParamPresence)))

			_, err = dspi.ParseUpmixParam("bogus")
			Expect(err).To(MatchError(ContainSubstring("unknown upmix parameter")))
		})
	})
})
