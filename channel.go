package dspi

import "fmt"

// ChannelInfo describes one metered channel.
type ChannelInfo struct {
	Index int
	Name  string
	Group string
}

// SetChannelName writes the user-configurable name for an audio channel.
// The channelIndex should be 0-based. Names live in RAM and are persisted
// via preset save.
func (d *Device) SetChannelName(channelIndex int, name string) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 32)
	copy(buf, []byte(name))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetChannelName, uint16(channelIndex), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_CHANNEL_NAME: %w", err)
	}

	return nil
}
