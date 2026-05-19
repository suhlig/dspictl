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
	reqSetChannelName  = 0x9B

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

	reqPresetSave           = 0x90
	reqPresetLoad           = 0x91
	reqPresetDelete         = 0x92
	reqPresetGetName        = 0x93
	reqPresetSetName        = 0x94
	reqPresetGetDir         = 0x95
	reqPresetSetStartup     = 0x96
	reqPresetGetStartup     = 0x97
	reqPresetSetIncludePins = 0x98
	reqPresetGetIncludePins = 0x99
	reqPresetGetActive      = 0x9A

	reqSetMatrixRoute = 0x70
	reqGetMatrixRoute = 0x71

	reqFactoryReset    = 0x53
	reqEnterBootloader = 0xF0

	reqGetCore1Mode     = 0x7A
	reqGetCore1Conflict = 0x7B

	reqSetOutputPin = 0x7C
	reqGetOutputPin = 0x7D

	reqSetOutputType    = 0xC0
	reqGetOutputType    = 0xC1
	reqSetI2SBckPin     = 0xC2
	reqGetI2SBckPin     = 0xC3
	reqSetMCKEnable     = 0xC4
	reqGetMCKEnable     = 0xC5
	reqSetMCKPin        = 0xC6
	reqGetMCKPin        = 0xC7
	reqSetMCKMultiplier = 0xC8
	reqGetMCKMultiplier = 0xC9

	reqSetMasterVolumeMode = 0xD4
	reqGetMasterVolumeMode = 0xD5
	reqSaveMasterVolume    = 0xD6

	reqGetBufferStats   = 0xB0
	reqGetUSBErrorStats = 0xB2
)
