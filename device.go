package dspi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/gousb"
)

const (
	// vendorInterfaceInRequest is the USB bmRequestType for vendor-specific,
	// device-to-host (IN) control transfers addressed to an interface.
	vendorInterfaceInRequest = gousb.ControlIn | gousb.ControlVendor | gousb.ControlInterface

	// vendorInterfaceOutRequest is the USB bmRequestType for vendor-specific,
	// host-to-device (OUT) control transfers addressed to an interface.
	vendorInterfaceOutRequest = gousb.ControlOut | gousb.ControlVendor | gousb.ControlInterface
)

type gousbControlTransfer struct {
	device *gousb.Device
}

func (c *gousbControlTransfer) ControlTransfer(bmRequestType, bRequest uint8, wValue, wIndex uint16, data []byte) (int, error) {
	return c.device.Control(bmRequestType, bRequest, wValue, wIndex, data)
}

func (c *gousbControlTransfer) Close() error {
	return c.device.Close()
}

// Device wraps a USB connection to a DSPi.
type Device struct {
	usb              USBControlTransfer
	ctx              *gousb.Context
	platform         Platform
	fwVersion        FirmwareVersion
	serial           string
	bus              int
	address          int
	closed           bool
	maxBands         int // cached from bulk header; 0 means uninitialised
	numInputChannels int // cached from bulk header; 0 means uninitialised
}

// Open opens a specific DSPi device identified by info.
func Open(info DeviceInfo) (*Device, error) {
	ctx := gousb.NewContext()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(dspiVID) && desc.Product == gousb.ID(dspiPID)
	})

	if err != nil {
		_ = ctx.Close()

		return nil, fmt.Errorf("opening DSPi device: %w", err)
	}

	var target *gousb.Device

	for _, dev := range devs {
		serial, err := dev.SerialNumber()

		if err != nil {
			_ = dev.Close()

			continue
		}

		if serial == info.Serial {
			target = dev
		} else {
			_ = dev.Close()
		}
	}

	if target == nil {
		_ = ctx.Close()

		return nil, fmt.Errorf("DSPi device with serial %s not found", info.Serial)
	}

	d := &Device{
		ctx:     ctx,
		usb:     &gousbControlTransfer{device: target},
		serial:  info.Serial,
		bus:     info.Bus,
		address: info.Address,
	}

	plat, fw, err := d.detectPlatform()

	if err != nil {
		d.Close()

		return nil, fmt.Errorf("detect platform: %w", err)
	}

	d.platform = plat
	d.fwVersion = fw

	return d, nil
}

// OpenAll opens all connected DSPi devices.
func OpenAll() ([]*Device, error) {
	infos, err := List()

	if err != nil {
		return nil, err
	}

	if len(infos) == 0 {
		return nil, fmt.Errorf("no DSPi device found")
	}

	devices := make([]*Device, 0, len(infos))
	var errs []error

	for _, info := range infos {
		dev, err := Open(info)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", info.Serial, err))

			continue
		}

		devices = append(devices, dev)
	}

	if len(devices) == 0 {
		if len(errs) > 0 {
			return nil, fmt.Errorf("no DSPi device could be opened: %w", errors.Join(errs...))
		}

		return nil, fmt.Errorf("no DSPi device could be opened")
	}

	return devices, nil
}

// Close releases the USB device.
func (d *Device) Close() {
	if d.closed {
		return
	}
	if d.usb != nil {
		_ = d.usb.Close()
	}
	if d.ctx != nil {
		_ = d.ctx.Close()
	}
	d.closed = true
}

// setControlTimeout sets the USB control transfer timeout on the underlying
// gousb device. It is a no-op when running against a mock.
func (d *Device) setControlTimeout(dur time.Duration) {
	if gt, ok := d.usb.(*gousbControlTransfer); ok {
		gt.device.ControlTimeout = dur
	}
}

// Platform returns the detected hardware platform.
func (d *Device) Platform() Platform { return d.platform }

// FirmwareVersion returns the decoded firmware version from REQ_GET_PLATFORM.
func (d *Device) FirmwareVersion() FirmwareVersion { return d.fwVersion }

// Serial returns the unique serial number of the device.
func (d *Device) Serial() string { return d.serial }

// GetSerial reads the 16-byte firmware serial via REQ_GET_SERIAL (0x7E).
func (d *Device) GetSerial() (string, error) {
	if d.closed {
		return "", fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetSerial, 0, vendorInterface, buf)

	if err != nil {
		return "", fmt.Errorf("REQ_GET_SERIAL: %w", err)
	}

	end := bytes.IndexByte(buf, 0)
	if end == -1 {
		end = len(buf)
	}

	return string(bytes.TrimSpace(buf[:end])), nil
}

// Bus returns the USB bus number of the device.
func (d *Device) Bus() int { return d.bus }

// Address returns the USB device address on the bus.
func (d *Device) Address() int { return d.address }

func (d *Device) detectPlatform() (Platform, FirmwareVersion, error) {
	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetPlatform, 0, vendorInterface, buf)

	if err != nil {
		return 0, FirmwareVersion{}, fmt.Errorf("REQ_GET_PLATFORM: %w", err)
	}
	if len(buf) < 4 {
		return 0, FirmwareVersion{}, fmt.Errorf("REQ_GET_PLATFORM: response too short (%d bytes)", len(buf))
	}

	platform := Platform(buf[0])
	fw := FirmwareVersion{
		Major: buf[1],
		Minor: buf[2] >> 4,
		Patch: buf[2] & 0x0F,
	}

	return platform, fw, nil
}

// ReadMeter polls the device for combined status (wValue=9).
// V16 response format: [peak pairs...] [CPU0(1)] [CPU1(1)] [ClipFlags(4)] [ActiveInputs(1)]
// That is channels*2 + 7 bytes total.
func (d *Device) ReadMeter() MeterSnapshot {
	var snap MeterSnapshot

	if d.closed {
		snap.err = fmt.Errorf("device is closed")

		return snap
	}

	buf := make([]byte, maxChannels*2+7)

	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetStatus, 9, vendorInterface, buf)

	if err != nil {
		snap.err = fmt.Errorf("REQ_GET_STATUS: %w", err)

		return snap
	}

	// Determine channel count from the response length.
	// V16 trailer: CPU0(1) + CPU1(1) + ClipFlags(4) + ActiveInputs(1) = 7 bytes.
	const v16TrailerLen = 7

	if n < v16TrailerLen || (n-v16TrailerLen)%2 != 0 {
		snap.err = fmt.Errorf("REQ_GET_STATUS: unexpected response length %d", n)

		return snap
	}

	snap.Channels = (n - v16TrailerLen) / 2
	actualCh := min(snap.Channels, maxChannels)

	needed := actualCh*2 + v16TrailerLen

	if n < needed {
		snap.err = fmt.Errorf("REQ_GET_STATUS: response too short (got %d, need %d)", n, needed)

		return snap
	}

	for i := range actualCh {
		raw := binary.LittleEndian.Uint16(buf[i*2:])
		snap.Peaks[i] = NewLevel(normalize(raw))
	}

	offset := actualCh * 2
	snap.CPU0 = int(buf[offset])
	snap.CPU1 = int(buf[offset+1])
	snap.ClipFlags = binary.LittleEndian.Uint32(buf[offset+2:])

	return snap
}

// ClearClips sends REQ_CLEAR_CLIPS (0x83) to reset the clip bitmask on the device.
// V16 returns a 4-byte uint32 (was 2 bytes).
func (d *Device) ClearClips() error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqClearClips, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("clearing clips: %w", err)
	}

	return nil
}

// SetMasterVolume sets the device-side master volume.
// Range: -128 (mute sentinel) to 0 dB. Typical range: -127 to 0 dB.
func (d *Device) SetMasterVolume(g Gain) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(g.DB())))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_MASTER_VOLUME: %w", err)
	}

	return nil
}

// GetMasterVolume reads the current device-side master volume.
func (d *Device) GetMasterVolume() (Gain, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetMasterVolume, 0, vendorInterface, buf)

	if err != nil {
		return 0, fmt.Errorf("REQ_GET_MASTER_VOLUME: %w", err)
	}

	return NewGain(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf)))), nil
}

// ChannelName fetches the name of a channel from the device hardware.
// The channelIndex should be 0-based.
func (d *Device) ChannelName(channelIndex int) (string, error) {
	if d.closed {
		return "", fmt.Errorf("device is closed")
	}

	buf := make([]byte, 64)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetChannelName, uint16(channelIndex), vendorInterface, buf)

	if err != nil {
		return "", fmt.Errorf("REQ_GET_CHANNEL_NAME: %w", err)
	}

	// Find null terminator or use full buffer
	end := bytes.IndexByte(buf[:n], 0)
	if end == -1 {
		end = n
	}

	return string(bytes.TrimSpace(buf[:end])), nil
}

// Channels queries the device for all channel names and returns a slice of ChannelInfo.
// The number of channels is determined by the platform (7 for RP2040, 17 for RP2350).
func (d *Device) Channels() ([]ChannelInfo, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	channelCount := 7
	if d.platform == PlatformRP2350 {
		channelCount = 17
	}

	numInputCh, err := d.NumInputChannels()
	if err != nil {
		numInputCh = 2 // fallback
	}

	inputSource, _ := d.GetInputSource()

	channels := make([]ChannelInfo, 0, channelCount)

	for i := 0; i < channelCount; i++ {
		name, err := d.ChannelName(i)
		if err != nil {
			return nil, fmt.Errorf("reading channel %d name: %w", i, err)
		}

		channels = append(channels, ChannelInfo{
			Index: i,
			Name:  name,
			Group: channelGroup(i, d.platform, inputSource, numInputCh),
		})
	}

	return channels, nil
}

// channelGroup returns the group name for a channel based on its index, platform,
// active input source, and the number of input channels from the bulk header.
func channelGroup(index int, platform Platform, inputSource int, numInputChannels int) string {
	totalCh := 7
	if platform == PlatformRP2350 {
		totalCh = 17
	}

	switch {
	case index < numInputChannels:
		switch inputSource {
		case InputSourceSPDIF, InputSourceSPDIF2, InputSourceSPDIF3, InputSourceSPDIF4:
			return "S/PDIF Input"
		case InputSourceI2S:
			return "I2S Input"
		case InputSourceADAT:
			return "ADAT Input"
		default:
			return "USB Input"
		}
	case index < totalCh-1:
		return "S/PDIF Output"
	default:
		return "PDM Sub"
	}
}

// NumInputChannels returns the number of input channels detected from the bulk header.
// Caches the value after first fetch (similar to MaxBands).
func (d *Device) NumInputChannels() (int, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}

	if d.numInputChannels > 0 {
		return d.numInputChannels, nil
	}

	bp, err := d.GetAllParams()
	if err != nil {
		return 0, fmt.Errorf("getting all params: %w", err)
	}

	numInput := bp.Header.NumInputChannels
	if numInput <= 0 {
		numInput = 2 // fallback for RP2040
	}

	d.numInputChannels = numInput

	return d.numInputChannels, nil
}

func normalize(raw uint16) float64 {
	// maxInt16 is the maximum value of a 16-bit signed integer,
	// used to normalize raw ADC readings to the 0.0–1.0 range.
	const maxInt16 = 32767

	return float64(raw) / maxInt16
}
