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
			// Build a V16-style payload (5864 bytes)
			payload := make([]byte, 5864)
			payload[0] = 1 // format version
			payload[1] = byte(dspi.PlatformRP2350)
			payload[2] = 17                                   // num channels
			payload[3] = 9                                    // num output channels
			payload[4] = 8                                    // num input channels (V16)
			payload[5] = 12                                   // max bands
			binary.LittleEndian.PutUint16(payload[6:8], 5864) // payload length

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
					// Third chunk: offset=4112, remaining 1752 bytes
					{uint16(dspi.ReqGetAllParamsChunk), 4112, 2}: payload[4112:],
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			bp, err := dev.GetAllParams()
			Expect(err).ToNot(HaveOccurred())
			Expect(bp).ToNot(BeNil())
			Expect(bp.Header.FormatVersion).To(Equal(uint8(1)))
			Expect(bp.Header.Platform).To(Equal(dspi.PlatformRP2350))
			Expect(bp.Header.NumChannels).To(Equal(17))
			Expect(bp.Header.NumOutputs).To(Equal(9))
			Expect(bp.Header.NumInputChannels).To(Equal(8))
			Expect(bp.Header.MaxBands).To(Equal(12))
			Expect(bp.Header.PayloadLength).To(Equal(5864))
			Expect(bp.Raw).To(HaveLen(5864))
		})

		It("sends chunked requests starting from offset 0", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqGetAllParamsChunk), 0, 2}: func() []byte {
						b := make([]byte, 16)
						binary.LittleEndian.PutUint16(b[6:8], 5864)
						return b
					}(),
					{uint16(dspi.ReqGetAllParamsChunk), 16, 2}:   make([]byte, 4096),
					{uint16(dspi.ReqGetAllParamsChunk), 4112, 2}: make([]byte, 1752),
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
					{uint16(dspi.ReqSetAllParamsChunk), 0, 2}: {},
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2040)

			bp := &dspi.BulkParams{
				Raw: []byte{0x01, 0x02, 0x03, 0x04},
			}
			err := dev.SetAllParams(bp)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAllParamsChunk)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01, 0x02, 0x03, 0x04}))
		})

		It("errors when the device is closed", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2350)
			dev.Close()

			bp := &dspi.BulkParams{Raw: []byte{0x01}}
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
					FormatVersion: 2,
					Platform:      dspi.PlatformRP2350,
				},
				Raw: make([]byte, 16),
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
			raw := make([]byte, 5864)
			raw[4712] = 2 // input_config offset: input source = I2S
			raw[4714] = 4 // rx pin (pair 0)
			raw[4715] = 1 // 48000
			raw[4716] = 8 // I2S input channels

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
			raw := make([]byte, 5864)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetInputSource(2)
			Expect(raw[4712]).To(Equal(byte(2)))

			bp.SetI2SRxPin(5)
			Expect(raw[4714]).To(Equal(byte(5)))

			bp.SetI2SInputRate(1)
			Expect(raw[4715]).To(Equal(byte(1)))

			bp.SetI2SInputChannels(4)
			Expect(raw[4716]).To(Equal(byte(4)))
		})
	})

	Describe("Field registry accessors", func() {
		It("reads and writes via GetU8/SetU8 on named fields", func() {
			raw := make([]byte, 5864)
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
			raw := make([]byte, 5864)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetU32("master_volume", 0, 0xDEADBEEF)
			v, ok := bp.GetU32("master_volume", 0)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint32(0xDEADBEEF)))
		})

		It("reads and writes via GetFloat32/SetFloat32", func() {
			raw := make([]byte, 5864)
			bp := &dspi.BulkParams{Raw: raw}

			bp.SetFloat32("delays", 0, 1.5)
			v, ok := bp.GetFloat32("delays", 0)
			Expect(ok).To(BeTrue())
			Expect(v).To(BeNumerically("~", 1.5, 0.001))
		})
	})
})
