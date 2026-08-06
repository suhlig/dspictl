package dspi_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("External control interfaces", func() {
	Describe("UartCtrlConfig wire layout", func() {
		It("encodes the 8-byte config", func() {
			cfg := &dspi.UartCtrlConfig{
				Enabled:      true,
				TxPin:        16,
				RxPin:        17,
				NotifyEnable: true,
				Baud:         115200,
			}

			raw := cfg.Encode()
			Expect(raw).To(HaveLen(8))
			Expect(raw[0]).To(Equal(byte(1)))
			Expect(raw[1]).To(Equal(byte(16)))
			Expect(raw[2]).To(Equal(byte(17)))
			Expect(raw[3]).To(Equal(byte(1)))
			Expect(binary.LittleEndian.Uint32(raw[4:8])).To(Equal(uint32(115200)))
		})

		It("round-trips through DecodeUartCtrlConfig", func() {
			cfg := &dspi.UartCtrlConfig{Enabled: true, TxPin: 16, RxPin: 17, Baud: 9600}
			decoded, err := dspi.DecodeUartCtrlConfig(cfg.Encode())
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.TxPin).To(Equal(uint8(16)))
			Expect(decoded.Baud).To(Equal(uint32(9600)))
		})
	})

	Describe("I2cCtrlConfig wire layout", func() {
		It("encodes the 8-byte config", func() {
			cfg := &dspi.I2cCtrlConfig{Enabled: true, SdaPin: 18, SclPin: 19, Address: 0x42}
			raw := cfg.Encode()
			Expect(raw).To(HaveLen(8))
			Expect(raw[0]).To(Equal(byte(1)))
			Expect(raw[1]).To(Equal(byte(18)))
			Expect(raw[2]).To(Equal(byte(19)))
			Expect(raw[3]).To(Equal(byte(0x42)))

			decoded, err := dspi.DecodeI2cCtrlConfig(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded.Address).To(Equal(uint8(0x42)))
		})
	})

	Describe("Device methods", func() {
		It("sets the UART config as an OUT transfer", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetUartConfig), 0, 2}: {},
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.SetUartConfig(&dspi.UartCtrlConfig{Enabled: true, TxPin: 16, RxPin: 17, Baud: 115200})
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetUartConfig)))
			Expect(mock.CapturedRequests[0].Data).To(HaveLen(8))
		})

		It("gets the UART config", func() {
			raw := (&dspi.UartCtrlConfig{Enabled: true, TxPin: 16, RxPin: 17, Baud: 230400}).Encode()
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetUartConfig), 0, 2}: raw,
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			cfg, err := dev.GetUartConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Enabled).To(BeTrue())
			Expect(cfg.Baud).To(Equal(uint32(230400)))
		})

		It("sets and gets the I2C config", func() {
			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqSetI2CConfig), 0, 2}: {},
				{uint16(dspi.ReqGetI2CConfig), 0, 2}: (&dspi.I2cCtrlConfig{Enabled: true, SdaPin: 18, SclPin: 19, Address: 0x42}).Encode(),
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			err := dev.SetI2CConfig(&dspi.I2cCtrlConfig{Enabled: true, SdaPin: 18, SclPin: 19, Address: 0x42})
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetI2CConfig)))

			cfg, err := dev.GetI2CConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.SdaPin).To(Equal(uint8(18)))
			Expect(cfg.Address).To(Equal(uint8(0x42)))
		})

		It("decodes the control interface status", func() {
			raw := make([]byte, 8)
			raw[0] = dspi.PinConfigSuccess
			raw[1] = 1 // uart live
			raw[2] = dspi.PinConfigPinInUse
			raw[3] = 0 // i2c down
			raw[4] = 1 // proto version

			mock := &mockControlTransfer{ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqGetCtrlIfaceStatus), 0, 2}: raw,
			}}
			dev := newTestDevice(mock, dspi.PlatformRP2350)

			st, err := dev.GetCtrlIfaceStatus()
			Expect(err).ToNot(HaveOccurred())
			Expect(st.UartLive).To(BeTrue())
			Expect(st.I2cLive).To(BeFalse())
			Expect(st.I2cLastStatus).To(Equal(uint8(dspi.PinConfigPinInUse)))
			Expect(st.ProtoVersion).To(Equal(uint8(1)))
		})
	})
})
