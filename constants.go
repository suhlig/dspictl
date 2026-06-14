package dspi

const (
	dspiVID         = 0x2e8b // Raspberry Pi (Pico vendor)
	dspiPID         = 0xfeaa // DSPi product ID
	vendorInterface = 0
	maxChannels     = 11

	ReqGetStatus       = 0x50
	ReqGetPlatform     = 0x7F
	ReqClearClips      = 0x83
	ReqSetMasterVolume = 0xD2
	ReqGetMasterVolume = 0xD3
	ReqGetChannelName  = 0x9C
	ReqSetChannelName  = 0x9B

	ReqSetPreampCh = 0xD0
	ReqGetPreampCh = 0xD1

	ReqSetOutputGain   = 0x74
	ReqGetOutputGain   = 0x75
	ReqSetOutputMute   = 0x76
	ReqGetOutputMute   = 0x77
	ReqSetOutputDelay  = 0x78
	ReqGetOutputDelay  = 0x79
	ReqSetOutputEnable = 0x72
	ReqGetOutputEnable = 0x73

	ReqPresetSave          = 0x90
	ReqPresetLoad          = 0x91
	ReqPresetDelete        = 0x92
	ReqPresetGetName       = 0x93
	ReqPresetSetName       = 0x94
	ReqPresetGetDir        = 0x95
	ReqPresetSetStartup    = 0x96
	ReqPresetGetStartup    = 0x97
	ReqSetOutputConfigMode = 0x98
	ReqGetOutputConfigMode = 0x99
	ReqPresetGetActive     = 0x9A

	ReqSetMatrixRoute = 0x70
	ReqGetMatrixRoute = 0x71

	ReqSetEQParam = 0x42
	ReqGetEQParam = 0x43
	ReqSetBypass  = 0x46
	ReqGetBypass  = 0x47

	ReqFactoryReset    = 0x53
	ReqEnterBootloader = 0xF0

	ReqGetCore1Mode     = 0x7A
	ReqGetCore1Conflict = 0x7B

	ReqSetOutputPin = 0x7C
	ReqGetOutputPin = 0x7D

	ReqSetOutputType    = 0xC0
	ReqGetOutputType    = 0xC1
	ReqSetI2SBckPin     = 0xC2
	ReqGetI2SBckPin     = 0xC3
	ReqSetMCKEnable     = 0xC4
	ReqGetMCKEnable     = 0xC5
	ReqSetMCKPin        = 0xC6
	ReqGetMCKPin        = 0xC7
	ReqSetMCKMultiplier = 0xC8
	ReqGetMCKMultiplier = 0xC9

	ReqSetMasterVolumeMode = 0xD4
	ReqGetMasterVolumeMode = 0xD5
	ReqSaveMasterVolume    = 0xD6

	ReqGetBufferStats   = 0xB0
	ReqGetUSBErrorStats = 0xB2

	ReqGetAllParams = 0xA0
	ReqSetAllParams = 0xA1

	ReqSetLoudness          = 0x58
	ReqGetLoudness          = 0x59
	ReqSetLoudnessReference = 0x5A
	ReqGetLoudnessReference = 0x5B
	ReqSetLoudnessIntensity = 0x5C
	ReqGetLoudnessIntensity = 0x5D

	ReqSetBandBypass = 0xD8
	ReqGetBandBypass = 0xD9

	ReqGetSerial = 0x7E

	ReqSetInputSource = 0xE0
	ReqGetInputSource = 0xE1
	ReqSetInputRate   = 0xED
	ReqGetInputRate   = 0xEE
	ReqSetI2SRxPin    = 0xF1
	ReqGetI2SRxPin    = 0xF2

	InputSourceUSB   = 0
	InputSourceSPDIF = 1
	InputSourceI2S   = 2

	PinConfigSuccess      = 0x00
	PinConfigInvalidPin   = 0x01
	PinConfigPinInUse     = 0x02
	PinConfigOutputActive = 0x04
)
