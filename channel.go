package dspi

// ChannelInfo describes one metered channel.
type ChannelInfo struct {
	Index int
	Name  string
	Group string
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
