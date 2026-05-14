package dspi

const (
	dspiVID         = 0x2e8b // Raspberry Pi (Pico vendor)
	dspiPID         = 0xfeaa // DSPi product ID
	reqGetStatus    = 0x50
	reqGetPlatform  = 0x7F
	reqClearClips   = 0x83
	vendorInterface = 0
	maxChannels     = 11
)
