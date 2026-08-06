package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Bulk", func() {
	Describe("GetAllParams", func() {
		It("returns the parsed header and raw payload", func() {
			// Build a V28 payload (5944 bytes)
			payload := make([]byte, 5944)
			payload[0] = 28 // format version
			payload[1] = byte(dspi.PlatformRP2350)
			payload[2] = 17                                   // num channels
			payload[3] = 9                                    // num output channels
			payload[4] = 8                                    // num input channels (V16)
			payload[5] = 12                                   // max bands
			binary.LittleEndian.PutUint16(payload[6:8], 5944) // payload length

			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					// First chunk: offset=0, request 16 bytes for header
					{uint16(dspi.ReqGetAllParamsChunk), 0, 2}: func() []byte {
						b := make([]byte, 16)
						copy(b, payload[:16])
						return b
					}(),
					// Second chunk: offset=16, up to 4096 bytes
					{uint16(dspi.ReqGetAllParamsChunk), 16, 2}: payload[16 : 16+4096],
					// Third chunk: offset=4112, remaining 1832 bytes
					{uint16(dspi.ReqGetAllParamsChunk), 4112, 2}: payload[4112:5944],
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			bp, err := dev.GetAllParams()
			Expect(err).ToNot(HaveOccurred())
			Expect(bp).ToNot(BeNil())
			Expect(bp.Header.FormatVersion).To(Equal(uint8(28)))
			Expect(bp.Header.Platform).To(Equal(dspi.PlatformRP2350))
			Expect(bp.Header.NumChannels).To(Equal(17))
			Expect(bp.Header.NumOutputs).To(Equal(9))
			Expect(bp.Header.NumInputChannels).To(Equal(8))
			Expect(bp.Header.MaxBands).To(Equal(12))
			Expect(bp.Header.PayloadLength).To(Equal(5944))
			Expect(bp.Raw).To(HaveLen(5944))
		})

		It("sends chunked requests starting from offset 0", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqGetAllParamsChunk), 0, 2}: func() []byte {
						b := make([]byte, 16)
						binary.LittleEndian.PutUint16(b[6:8], 5944)
						return b
					}(),
					{uint16(dspi.ReqGetAllParamsChunk), 16, 2}:   make([]byte, 4096),
					{uint16(dspi.ReqGetAllParamsChunk), 4112, 2}: make([]byte, 1832),
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			_, _ = dev.GetAllParams()

			Expect(mock.CapturedRequests).To(HaveLen(3))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetAllParamsChunk)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(16)))
			Expect(mock.CapturedRequests[2].WValue).To(Equal(uint16(4112)))
		})

		It("errors on a short response", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqGetAllParamsChunk), 0, 2}: make([]byte, 5),
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			_, err := dev.GetAllParams()
			Expect(err).To(HaveOccurred())
		})

		It("errors when the device is closed", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			dev.Close()
			_, err := dev.GetAllParams()
			Expect(err).To(MatchError(ContainSubstring("closed")))
		})
	})

	Describe("SetAllParams", func() {
		It("sends the raw payload via chunked transfers", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqSetAllParamsChunk), 0, 2}:    {},
					{uint16(dspi.ReqSetAllParamsChunk), 4096, 2}: {},
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2040)

			bp := &dspi.BulkParams{
				Raw: make([]byte, 5944),
			}
			err := dev.SetAllParams(bp)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(2))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAllParamsChunk)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].Data).To(HaveLen(4096))
			Expect(mock.CapturedRequests[1].WValue).To(Equal(uint16(4096)))
			Expect(mock.CapturedRequests[1].Data).To(HaveLen(1848))
		})

		It("errors on a payload that is not the current wire size", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			bp := &dspi.BulkParams{
				Header: dspi.BulkHeader{
					FormatVersion: 24,
					Platform:      dspi.PlatformRP2350,
				},
				Raw: make([]byte, 5900), // pre-V25 export
			}
			err := dev.SetAllParams(bp)
			Expect(err).To(MatchError(ContainSubstring("snapshot is 5900 bytes, device expects 5944")))
		})

		It("errors when the device is closed", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			dev.Close()

			bp := &dspi.BulkParams{Raw: make([]byte, 5944)}
			err := dev.SetAllParams(bp)
			Expect(err).To(MatchError(ContainSubstring("closed")))
		})

		It("errors when given nil params", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			err := dev.SetAllParams(nil)
			Expect(err).To(MatchError(ContainSubstring("no params")))
		})

		It("errors on platform mismatch", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2040)

			bp := &dspi.BulkParams{
				Header: dspi.BulkHeader{
					FormatVersion: 28,
					Platform:      dspi.PlatformRP2350,
				},
				Raw: make([]byte, 5944),
			}
			err := dev.SetAllParams(bp)
			Expect(err).To(MatchError(ContainSubstring("platform mismatch")))
			Expect(err).To(MatchError(ContainSubstring("RP2350")))
			Expect(err).To(MatchError(ContainSubstring("RP2040")))
		})

		It("errors when given empty raw data", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			err := dev.SetAllParams(&dspi.BulkParams{})
			Expect(err).To(MatchError(ContainSubstring("no params")))
		})
	})

	Describe("WireInputConfig accessors via field registry", func() {
		It("reads input source, RX pin, rate, and I2S channels", func() {
			raw := make([]byte, 5944)
			raw[4716] = 2 // input_config offset: input source = I2S
			raw[4718] = 4 // rx pin (pair 0)
			raw[4719] = 1 // 48000
			raw[4720] = 8 // I2S input channels

			bp := &dspi.BulkParams{Raw: raw}

			src, ok := bp.InputSource()
			Expect(ok).To(BeTrue())
			Expect(src).To(Equal(2))

			pin, ok := bp.I2SRxPin()
			Expect(ok).To(BeTrue())
			Expect(pin).To(Equal(4))

			rate, ok := bp.I2SInputRate()
			Expect(ok).To(BeTrue())
			Expect(rate).To(Equal(1))

			ch, ok := bp.I2SInputChannels()
			Expect(ok).To(BeTrue())
			Expect(ch).To(Equal(8))
		})

		It("returns false for short payloads", func() {
			bp := &dspi.BulkParams{Raw: make([]byte, 100)}
			_, ok := bp.InputSource()
			Expect(ok).To(BeFalse())
		})

		It("writes values into raw", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetInputSource(2)
			Expect(raw[4716]).To(Equal(byte(2)))

			bp.SetI2SRxPin(5)
			Expect(raw[4718]).To(Equal(byte(5)))

			bp.SetI2SInputRate(1)
			Expect(raw[4719]).To(Equal(byte(1)))

			bp.SetI2SInputChannels(4)
			Expect(raw[4720]).To(Equal(byte(4)))
		})

		It("reads and writes ADAT input fields at their V28 offsets", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetAdatInputPin(22)
			bp.SetAdatInputEnabledP1(2)
			bp.SetAdatInputClockModeP1(2)

			pin, ok := bp.AdatInputPin()
			Expect(ok).To(BeTrue())
			Expect(pin).To(Equal(uint8(22)))
			Expect(raw[4716+13]).To(Equal(byte(22))) // input_config +13 (V28: after spdif_rx_pin_ext[3])

			enP1, ok := bp.AdatInputEnabledP1()
			Expect(ok).To(BeTrue())
			Expect(enP1).To(Equal(uint8(2)))
			Expect(raw[4716+14]).To(Equal(byte(2)))

			modeP1, ok := bp.AdatInputClockModeP1()
			Expect(ok).To(BeTrue())
			Expect(modeP1).To(Equal(uint8(2)))
			Expect(raw[4716+15]).To(Equal(byte(2)))
		})

		It("reads and writes the V28 SPDIF ext and I2S clock fields", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetSpdifRxPinExt(1, 20)
			bp.SetSpdifRxPinExt(2, 21)
			bp.SetSpdifRxPinExt(3, 22)
			bp.SetSpdifRxEnabledExtP1(4) // all three optional inputs enabled
			bp.SetI2SClockMode(1)        // slave

			Expect(raw[4716+8]).To(Equal(byte(20)))
			Expect(raw[4716+9]).To(Equal(byte(21)))
			Expect(raw[4716+10]).To(Equal(byte(22)))
			Expect(raw[4716+11]).To(Equal(byte(4)))
			Expect(raw[4716+12]).To(Equal(byte(1)))

			pin, ok := bp.SpdifRxPinExt(2)
			Expect(ok).To(BeTrue())
			Expect(pin).To(Equal(uint8(21)))

			_, ok = bp.SpdifRxPinExt(4) // out of range
			Expect(ok).To(BeFalse())

			enP1, ok := bp.SpdifRxEnabledExtP1()
			Expect(ok).To(BeTrue())
			Expect(enP1).To(Equal(uint8(4)))

			mode, ok := bp.I2SClockMode()
			Expect(ok).To(BeTrue())
			Expect(mode).To(Equal(uint8(1)))
		})

		It("reads and writes the loudness reference and intensity in global", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetLoudnessRefSPL(75)
			bp.SetLoudnessIntensityPct(100)

			spl, ok := bp.LoudnessRefSPL()
			Expect(ok).To(BeTrue())
			Expect(spl).To(BeNumerically("~", float32(75), 0.001))
			Expect(raw[16+11]).To(Equal(byte(0x42))) // 75.0f LE high byte

			pct, ok := bp.LoudnessIntensityPct()
			Expect(ok).To(BeTrue())
			Expect(pct).To(BeNumerically("~", float32(100), 0.001))
		})
	})

	Describe("Field registry accessors", func() {
		It("reads and writes via GetU8/SetU8 on named fields", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			// Write via field registry
			bp.SetU8("input_config", 0, 2)
			bp.SetU8("input_config", 2, 7)
			bp.SetU8("input_config", 3, 0)

			// Read back
			v, ok := bp.GetU8("input_config", 0)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint8(2)))

			v, ok = bp.GetU8("input_config", 2)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint8(7)))

			// Unknown field returns false
			_, ok = bp.GetU8("nonexistent", 0)
			Expect(ok).To(BeFalse())
		})

		It("reads and writes via GetU32/SetU32", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetU32("master_volume", 0, 0xDEADBEEF)
			v, ok := bp.GetU32("master_volume", 0)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint32(0xDEADBEEF)))
		})

		It("reads and writes via GetFloat32/SetFloat32", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetFloat32("delays", 0, 1.5)
			v, ok := bp.GetFloat32("delays", 0)
			Expect(ok).To(BeTrue())
			Expect(v).To(BeNumerically("~", 1.5, 0.001))
		})

		It("reads and writes psychoacoustic bass fields", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetPsybassEnabled(true)
			bp.SetPsybassOutputMask(0x1234)
			bp.SetPsybassCutoff(80)
			bp.SetPsybassHarmonics(-12)
			bp.SetPsybassDrive(6)
			bp.SetPsybassCharacter(50)
			bp.SetPsybassOriginal(-30)

			enabled, ok := bp.PsybassEnabled()
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeTrue())

			mask, ok := bp.PsybassOutputMask()
			Expect(ok).To(BeTrue())
			Expect(mask).To(Equal(uint16(0x1234)))

			cutoff, ok := bp.PsybassCutoff()
			Expect(ok).To(BeTrue())
			Expect(cutoff).To(BeNumerically("~", float32(80), 0.001))

			harmonics, ok := bp.PsybassHarmonics()
			Expect(ok).To(BeTrue())
			Expect(harmonics).To(BeNumerically("~", float32(-12), 0.001))

			drive, ok := bp.PsybassDrive()
			Expect(ok).To(BeTrue())
			Expect(drive).To(BeNumerically("~", float32(6), 0.001))

			character, ok := bp.PsybassCharacter()
			Expect(ok).To(BeTrue())
			Expect(character).To(BeNumerically("~", float32(50), 0.001))

			original, ok := bp.PsybassOriginal()
			Expect(ok).To(BeTrue())
			Expect(original).To(BeNumerically("~", float32(-30), 0.001))
		})

		It("reads and writes upmixer fields at wire offset 5900", func() {
			raw := make([]byte, 5944)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetUpmixEnabled(true)
			bp.SetUpmixCenterMode(2) // OFF (V27+)
			bp.SetUpmixSurroundMode(2)
			bp.SetUpmixPresenceQ1(-8) // -4 dB
			bp.SetUpmixStrengthPct(50)
			bp.SetUpmixCenterWidthPct(25)
			bp.SetUpmixCorrThresholdPct(40)
			bp.SetUpmixAttackMs(10)
			bp.SetUpmixReleaseMs(200)
			bp.SetUpmixDetectorHpfHz(30)
			bp.SetUpmixSurroundDelayMs(12)
			bp.SetUpmixSurroundHpfHz(300)
			bp.SetUpmixSurroundLpfHz(8000)
			bp.SetUpmixDecorrPct(75)

			enabled, ok := bp.UpmixEnabled()
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeTrue())
			Expect(raw[5900]).To(Equal(byte(1)))

			center, ok := bp.UpmixCenterMode()
			Expect(ok).To(BeTrue())
			Expect(center).To(Equal(uint8(2)))

			surround, ok := bp.UpmixSurroundMode()
			Expect(ok).To(BeTrue())
			Expect(surround).To(Equal(uint8(2)))

			presence, ok := bp.UpmixPresenceQ1()
			Expect(ok).To(BeTrue())
			Expect(presence).To(Equal(int8(-8)))
			Expect(raw[5903]).To(Equal(byte(0xF8))) // -8 as uint8

			strength, ok := bp.UpmixStrengthPct()
			Expect(ok).To(BeTrue())
			Expect(strength).To(BeNumerically("~", float32(50), 0.001))

			width, ok := bp.UpmixCenterWidthPct()
			Expect(ok).To(BeTrue())
			Expect(width).To(BeNumerically("~", float32(25), 0.001))

			corr, ok := bp.UpmixCorrThresholdPct()
			Expect(ok).To(BeTrue())
			Expect(corr).To(BeNumerically("~", float32(40), 0.001))

			attack, ok := bp.UpmixAttackMs()
			Expect(ok).To(BeTrue())
			Expect(attack).To(BeNumerically("~", float32(10), 0.001))

			release, ok := bp.UpmixReleaseMs()
			Expect(ok).To(BeTrue())
			Expect(release).To(BeNumerically("~", float32(200), 0.001))

			detHpf, ok := bp.UpmixDetectorHpfHz()
			Expect(ok).To(BeTrue())
			Expect(detHpf).To(BeNumerically("~", float32(30), 0.001))

			surDelay, ok := bp.UpmixSurroundDelayMs()
			Expect(ok).To(BeTrue())
			Expect(surDelay).To(BeNumerically("~", float32(12), 0.001))

			surHpf, ok := bp.UpmixSurroundHpfHz()
			Expect(ok).To(BeTrue())
			Expect(surHpf).To(BeNumerically("~", float32(300), 0.001))

			surLpf, ok := bp.UpmixSurroundLpfHz()
			Expect(ok).To(BeTrue())
			Expect(surLpf).To(BeNumerically("~", float32(8000), 0.001))

			decorr, ok := bp.UpmixDecorrPct()
			Expect(ok).To(BeTrue())
			Expect(decorr).To(BeNumerically("~", float32(75), 0.001))
		})
	})
})
