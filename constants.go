package dspi

const (
	dspiVID         = 0x2e8b // Raspberry Pi (Pico vendor)
	dspiPID         = 0xfeaa // DSPi product ID
	vendorInterface = 0
	maxChannels     = 11

	reqGetStatus       = 0x50
	reqGetPlatform     = 0x7F
	reqClearClips      = 0x83
	reqSetMasterVolume = 0xD2
	reqGetMasterVolume = 0xD3
	reqGetChannelName  = 0x9C

	reqSetPreampCh = 0xD0
	reqGetPreampCh = 0xD1

	reqSetOutputGain   = 0x74
	reqGetOutputGain   = 0x75
	reqSetOutputMute   = 0x76
	reqGetOutputMute   = 0x77
	reqSetOutputDelay  = 0x78
	reqGetOutputDelay  = 0x79
	reqSetOutputEnable = 0x72
	reqGetOutputEnable = 0x73

	reqPresetSave      = 0x90
	reqPresetLoad      = 0x91
	reqPresetDelete    = 0x92
	reqPresetGetName   = 0x93
	reqPresetSetName   = 0x94
	reqPresetGetDir    = 0x95
	reqPresetGetActive = 0x9A

	reqSetMatrixRoute = 0x70
	reqGetMatrixRoute = 0x71
)
