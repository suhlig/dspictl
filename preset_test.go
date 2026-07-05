package dspi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/suhlig/dspi"
)

var _ = Describe("Preset", func() {
	var (
		mock *mockControlTransfer
		dev  *dspi.Device
	)

	BeforeEach(func() {
		mock = &mockControlTransfer{
			ReturnData: map[[3]uint16][]byte{
				{uint16(dspi.ReqPresetSave), 5, 2}:          {0x00},
				{uint16(dspi.ReqPresetLoad), 5, 2}:          {0x00},
				{uint16(dspi.ReqPresetDelete), 5, 2}:        {0x00},
				{uint16(dspi.ReqPresetGetName), 5, 2}:       append([]byte("Test Preset"), make([]byte, 21)...),
				{uint16(dspi.ReqPresetSetName), 5, 2}:       {},
				{uint16(dspi.ReqPresetGetDir), 0, 2}:        {0x1F, 0x00, 0x01, 0x03, 0x05, 0x01, 0x02},
				{uint16(dspi.ReqPresetGetActive), 0, 2}:     {0x04},
				{uint16(dspi.ReqPresetSetStartup), 0, 2}:    {},
				{uint16(dspi.ReqPresetGetStartup), 0, 2}:    {0x01, 0x02},
				{uint16(dspi.ReqSetOutputConfigMode), 0, 2}: {},
				{uint16(dspi.ReqGetOutputConfigMode), 0, 2}: {0x01},
			},
		}
		dev = newTestDevice(mock, dspi.PlatformRP2350)
	})

	Describe("PresetSave", func() {
		It("sends the correct request", func() {
			err := dev.PresetSave(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetSave)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error on non-zero status", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqPresetSave), 5, 2}] = []byte{0x01}
			err := dev.PresetSave(5)
			Expect(err).To(MatchError(ContainSubstring("status 0x01")))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.PresetSave(5)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("PresetLoad", func() {
		It("sends the correct request", func() {
			err := dev.PresetLoad(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetLoad)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error on non-zero status", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqPresetLoad), 5, 2}] = []byte{0x02}
			err := dev.PresetLoad(5)
			Expect(err).To(MatchError(ContainSubstring("status 0x02")))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.PresetLoad(5)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("PresetDelete", func() {
		It("sends the correct request", func() {
			err := dev.PresetDelete(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetDelete)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error on non-zero status", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqPresetDelete), 5, 2}] = []byte{0x03}
			err := dev.PresetDelete(5)
			Expect(err).To(MatchError(ContainSubstring("status 0x03")))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.PresetDelete(5)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetPresetName", func() {
		It("returns the preset name", func() {
			name, err := dev.GetPresetName(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("Test Preset"))
		})

		It("trims trailing spaces", func() {
			buf := make([]byte, 32)
			copy(buf, []byte("Padded"))
			for i := 6; i < 32; i++ {
				buf[i] = ' '
			}
			mock.ReturnData[[3]uint16{uint16(dspi.ReqPresetGetName), 5, 2}] = buf
			name, err := dev.GetPresetName(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("Padded"))
		})

		It("handles null-terminated strings", func() {
			buf := make([]byte, 32)
			copy(buf, []byte("Short"))
			buf[5] = 0
			mock.ReturnData[[3]uint16{uint16(dspi.ReqPresetGetName), 5, 2}] = buf
			name, err := dev.GetPresetName(5)
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("Short"))
		})

		It("sends the correct request", func() {
			_, _ = dev.GetPresetName(5)
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetGetName)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetPresetName(5)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetPresetName", func() {
		It("sends the correct request", func() {
			err := dev.SetPresetName(5, "My Preset")
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetSetName)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(5)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("pads the name to 32 bytes with zeros", func() {
			err := dev.SetPresetName(5, "My Preset")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(mock.CapturedRequests[0].Data)).To(Equal(32))
			Expect(mock.CapturedRequests[0].Data[:9]).To(Equal([]byte("My Preset")))
			Expect(mock.CapturedRequests[0].Data[9:]).To(Equal(make([]byte, 23)))
		})

		It("truncates names longer than 32 bytes", func() {
			longName := "This is a very long preset name that exceeds thirty-two characters"
			err := dev.SetPresetName(5, longName)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(mock.CapturedRequests[0].Data)).To(Equal(32))
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte(longName[:32])))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetPresetName(5, "My Preset")
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetPresetDirectory", func() {
		It("parses the 7-byte response correctly", func() {
			dir, err := dev.GetPresetDirectory()
			Expect(err).ToNot(HaveOccurred())
			Expect(dir.SlotOccupied).To(Equal(uint16(0x001F)))
			Expect(dir.StartupMode).To(Equal(1))
			Expect(dir.DefaultSlot).To(Equal(3))
			Expect(dir.LastActive).To(Equal(5))
			Expect(dir.OutputConfigMode).To(Equal(1))
			Expect(dir.MasterVolMode).To(Equal(2))
		})

		It("sends the correct request", func() {
			_, _ = dev.GetPresetDirectory()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetGetDir)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetPresetDirectory()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetActivePreset", func() {
		It("returns the active preset slot", func() {
			slot, err := dev.GetActivePreset()
			Expect(err).ToNot(HaveOccurred())
			Expect(slot).To(Equal(4))
		})

		It("sends the correct request", func() {
			_, _ = dev.GetActivePreset()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetGetActive)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetActivePreset()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetPresetStartup", func() {
		It("sends the correct request", func() {
			err := dev.SetPresetStartup(1, 2)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetSetStartup)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("encodes mode and defaultSlot as 2 bytes", func() {
			err := dev.SetPresetStartup(1, 2)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01, 0x02}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetPresetStartup(1, 2)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetPresetStartup", func() {
		It("returns mode and defaultSlot", func() {
			mode, defaultSlot, err := dev.GetPresetStartup()
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(1))
			Expect(defaultSlot).To(Equal(2))
		})

		It("sends the correct request", func() {
			_, _, _ = dev.GetPresetStartup()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqPresetGetStartup)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, _, err := dev.GetPresetStartup()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("SetOutputConfigMode", func() {
		It("sends the correct request", func() {
			err := dev.SetOutputConfigMode(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqSetOutputConfigMode)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("encodes mode 1 as 0x01", func() {
			err := dev.SetOutputConfigMode(1)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x01}))
		})

		It("encodes mode 0 as 0x00", func() {
			err := dev.SetOutputConfigMode(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(mock.CapturedRequests[0].Data).To(Equal([]byte{0x00}))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			err := dev.SetOutputConfigMode(1)
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})

	Describe("GetOutputConfigMode", func() {
		It("returns 1 when the byte is non-zero", func() {
			mode, err := dev.GetOutputConfigMode()
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(1))
		})

		It("returns 0 when the byte is zero", func() {
			mock.ReturnData[[3]uint16{uint16(dspi.ReqGetOutputConfigMode), 0, 2}] = []byte{0x00}
			mode, err := dev.GetOutputConfigMode()
			Expect(err).ToNot(HaveOccurred())
			Expect(mode).To(Equal(0))
		})

		It("sends the correct request", func() {
			_, _ = dev.GetOutputConfigMode()
			Expect(mock.CapturedRequests).To(HaveLen(1))
			Expect(mock.CapturedRequests[0].BRequest).To(Equal(uint8(dspi.ReqGetOutputConfigMode)))
			Expect(mock.CapturedRequests[0].WValue).To(Equal(uint16(0)))
			Expect(mock.CapturedRequests[0].WIndex).To(Equal(uint16(2)))
		})

		It("returns an error when the device is closed", func() {
			dev.Close()
			_, err := dev.GetOutputConfigMode()
			Expect(err).To(MatchError(ContainSubstring("device is closed")))
		})
	})
})
