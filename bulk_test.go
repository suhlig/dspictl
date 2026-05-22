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
			payload := make([]byte, 64)
			payload[0] = 1 // format version
			payload[1] = byte(dspi.PlatformRP2350)
			payload[2] = 11                                   // num channels
			payload[3] = 9                                    // num output channels
			payload[4] = 2                                    // num input channels
			payload[5] = 12                                   // max bands
			binary.LittleEndian.PutUint16(payload[6:8], 2832) // payload length
			binary.LittleEndian.PutUint16(payload[8:10], 2)   // fw major
			binary.LittleEndian.PutUint16(payload[10:12], 5)  // fw minor

			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqGetAllParams), 0, 0}: payload,
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			bp, err := dev.GetAllParams()
			Expect(err).ToNot(HaveOccurred())
			Expect(bp).ToNot(BeNil())
			Expect(bp.Header.FormatVersion).To(Equal(uint8(1)))
			Expect(bp.Header.Platform).To(Equal(dspi.PlatformRP2350))
			Expect(bp.Header.NumChannels).To(Equal(11))
			Expect(bp.Header.NumOutputs).To(Equal(9))
			Expect(bp.Header.PayloadLength).To(Equal(2832))
			Expect(bp.Header.FWMajor).To(Equal(uint16(2)))
			Expect(bp.Header.FWMinor).To(Equal(uint16(5)))
			Expect(bp.Raw).To(HaveLen(64))
		})

		It("sends the correct request", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqGetAllParams), 0, 0}: make([]byte, 16),
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2040)
			_, _ = dev.GetAllParams()

			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetAllParams)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(0)))
		})

		It("errors on a short response", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqGetAllParams), 0, 0}: make([]byte, 5),
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2040)
			_, err := dev.GetAllParams()
			Expect(err).To(HaveOccurred())
		})

		It("errors when the device is closed", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2040)
			dev.Close()
			_, err := dev.GetAllParams()
			Expect(err).To(MatchError(ContainSubstring("closed")))
		})
	})

	Describe("SetAllParams", func() {
		It("sends the raw payload back", func() {
			mock := &mockControlTransfer{
				ReturnData: map[[3]uint16][]byte{
					{uint16(dspi.ReqSetAllParams), 0, 0}: {},
				},
			}
			dev := newTestDevice(mock, dspi.PlatformRP2040)

			bp := &dspi.BulkParams{
				Raw: []byte{0x01, 0x02, 0x03, 0x04},
			}
			err := dev.SetAllParams(bp)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetAllParams)))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01, 0x02, 0x03, 0x04}))
		})

		It("errors when the device is closed", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2040)
			dev.Close()

			bp := &dspi.BulkParams{Raw: []byte{0x01}}
			err := dev.SetAllParams(bp)
			Expect(err).To(MatchError(ContainSubstring("closed")))
		})

		It("errors when given nil params", func() {
			mock := &mockControlTransfer{}
			dev := newTestDevice(mock, dspi.PlatformRP2040)
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
			dev := newTestDevice(mock, dspi.PlatformRP2040)
			err := dev.SetAllParams(&dspi.BulkParams{})
			Expect(err).To(MatchError(ContainSubstring("no params")))
		})
	})
})
