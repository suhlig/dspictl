package dspi

// ChannelInfo describes one metered channel.
type ChannelInfo struct {
	Index int
	Name  string
	Group string
}

// ChannelTable returns the channel mapping for the given platform.
func ChannelTable(platform Platform) []ChannelInfo {
	if platform == PlatformRP2350 {
		return []ChannelInfo{
			{0, "USB L", "USB Input"},
			{1, "USB R", "USB Input"},
			{2, "SPDIF 1 L", "S/PDIF Output"},
			{3, "SPDIF 1 R", "S/PDIF Output"},
			{4, "SPDIF 2 L", "S/PDIF Output"},
			{5, "SPDIF 2 R", "S/PDIF Output"},
			{6, "SPDIF 3 L", "S/PDIF Output"},
			{7, "SPDIF 3 R", "S/PDIF Output"},
			{8, "SPDIF 4 L", "S/PDIF Output"},
			{9, "SPDIF 4 R", "S/PDIF Output"},
			{10, "PDM Sub", "PDM Sub"},
		}
	}

	return []ChannelInfo{
		{0, "USB L", "USB Input"},
		{1, "USB R", "USB Input"},
		{2, "SPDIF 1 L", "S/PDIF Output"},
		{3, "SPDIF 1 R", "S/PDIF Output"},
		{4, "SPDIF 2 L", "S/PDIF Output"},
		{5, "SPDIF 2 R", "S/PDIF Output"},
		{6, "PDM Sub", "PDM Sub"},
	}
}
