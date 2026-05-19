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
	if d.device == nil {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 32)
	copy(buf, []byte(name))
	_, err := d.device.Control(vendorInterfaceOutRequest, reqSetChannelName, uint16(channelIndex), vendorInterface, buf)

	if err != nil {
		return fmt.Errorf("REQ_SET_CHANNEL_NAME: %w", err)
	}

	return nil
}

// ChannelTable returns the static channel layout for the given platform.
// This includes default channel names and groups without querying a device.
func ChannelTable(platform Platform) []ChannelInfo {
	var count int
	if platform == PlatformRP2350 {
		count = 11
	} else {
		count = 7
	}

	channels := make([]ChannelInfo, count)
	for i := 0; i < count; i++ {
		channels[i] = ChannelInfo{
			Index: i,
			Name:  defaultChannelName(i, platform),
			Group: channelGroup(i, platform),
		}
	}

	return channels
}
