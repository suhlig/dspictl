package dspi

const (
	dspiVID         = 0x2e8b // Raspberry Pi (Pico vendor)
	dspiPID         = 0xfeaa // DSPi product ID
	vendorInterface = 2
	maxChannels     = 17

	ReqSaveParams  = 0x51
	ReqGetStatus   = 0x50
	ReqGetPlatform = 0x7F
	ReqClearClips  = 0x83

	// Psychoacoustic Bass (missing-fundamental bass enhancement).
	ReqSetPsybass          = 0x30
	ReqGetPsybass          = 0x31
	ReqSetPsybassCutoff    = 0x32
	ReqGetPsybassCutoff    = 0x33
	ReqSetPsybassHarmonics = 0x34
	ReqGetPsybassHarmonics = 0x35
	ReqSetPsybassDrive     = 0x36
	ReqGetPsybassDrive     = 0x37
	ReqSetPsybassCharacter = 0x38
	ReqGetPsybassCharacter = 0x39
	ReqSetPsybassOriginal  = 0x3A
	ReqGetPsybassOriginal  = 0x3B
	ReqSetPsybassMask      = 0x3C
	ReqGetPsybassMask      = 0x3D
	ReqSetMasterVolume     = 0xD2
	ReqGetMasterVolume     = 0xD3
	ReqGetChannelName      = 0x9C
	ReqSetChannelName      = 0x9B

	ReqSetPreamp   = 0x44
	ReqGetPreamp   = 0x45
	ReqSetPreampCh = 0xD0
	ReqGetPreampCh = 0xD1

	ReqSetChannelGain = 0x54
	ReqGetChannelGain = 0x55
	ReqSetChannelMute = 0x56
	ReqGetChannelMute = 0x57

	ReqSetDelay = 0x48
	ReqGetDelay = 0x49

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

	ReqSetMasterVolumeMode  = 0xD4
	ReqGetMasterVolumeMode  = 0xD5
	ReqSaveMasterVolume     = 0xD6
	ReqGetSavedMasterVolume = 0xD7

	ReqSetUserVolume = 0xDA
	ReqGetUserVolume = 0xDB

	ReqGetBufferStats   = 0xB0
	ReqResetBufferStats = 0xB1
	ReqGetUSBErrorStats = 0xB2

	ReqSaveOutputConfig = 0x52

	ReqGetSpdifRxStatus   = 0xE2
	ReqGetSpdifRxChStatus = 0xE3
	ReqSetSpdifRxPin      = 0xE4
	ReqGetSpdifRxPin      = 0xE5

	ReqGetAllParams = 0xA0
	ReqSetAllParams = 0xA1

	ReqSetLoudness          = 0x58
	ReqGetLoudness          = 0x59
	ReqSetLoudnessReference = 0x5A
	ReqGetLoudnessReference = 0x5B
	ReqSetLoudnessIntensity = 0x5C
	ReqGetLoudnessIntensity = 0x5D
	ReqSetLoudnessMask      = 0xFA
	ReqGetLoudnessMask      = 0xFB

	ReqSetBandBypass = 0xD8
	ReqGetBandBypass = 0xD9

	// ADAT Input commands (RP2350 only; state round-trips on RP2040).
	ReqSetAdatInputEnable    = 0x68
	ReqGetAdatInputEnable    = 0x69
	ReqSetAdatInputPin       = 0x6A
	ReqGetAdatInputPin       = 0x6B
	ReqSetAdatInputClockMode = 0x6C
	ReqGetAdatInputClockMode = 0x6D
	ReqGetAdatInputStatus    = 0x6E

	ReqSetCrossfeed        = 0x5E
	ReqGetCrossfeed        = 0x5F
	ReqSetCrossfeedPreset  = 0x60
	ReqGetCrossfeedPreset  = 0x61
	ReqSetCrossfeedFreq    = 0x62
	ReqGetCrossfeedFreq    = 0x63
	ReqSetCrossfeedFeed    = 0x64
	ReqGetCrossfeedFeed    = 0x65
	ReqSetCrossfeedITD     = 0x66
	ReqGetCrossfeedITD     = 0x67
	ReqSetCrossfeedOutputs = 0xFC
	ReqGetCrossfeedOutputs = 0xFD

	ReqSetLeveller          = 0xB4
	ReqGetLeveller          = 0xB5
	ReqSetLevellerAmount    = 0xB6
	ReqGetLevellerAmount    = 0xB7
	ReqSetLevellerSpeed     = 0xB8
	ReqGetLevellerSpeed     = 0xB9
	ReqSetLevellerMaxGain   = 0xBA
	ReqGetLevellerMaxGain   = 0xBB
	ReqSetLevellerLookahead = 0xBC
	ReqGetLevellerLookahead = 0xBD
	ReqSetLevellerGate      = 0xBE
	ReqGetLevellerGate      = 0xBF
	ReqSetLevellerMasks     = 0xDE
	ReqGetLevellerMasks     = 0xDF

	ReqSetDACHwMuteConfig = 0xEA
	ReqGetDACHwMuteConfig = 0xEB
	ReqTestDACHwMute      = 0xEC

	ReqSetLGSoundSync       = 0xE6
	ReqGetLGSoundSync       = 0xE7
	ReqGetLGSoundSyncStatus = 0xE8

	ReqGetSerial = 0x7E

	ReqSetInputSource      = 0xE0
	ReqGetInputSource      = 0xE1
	ReqSetInputRate        = 0xED
	ReqGetInputRate        = 0xEE
	ReqSetI2SRxPin         = 0xF1
	ReqGetI2SRxPin         = 0xF2
	ReqSetI2SInputChannels = 0xF3
	ReqGetI2SInputChannels = 0xF4

	ReqGetAllParamsChunk = 0xA2
	ReqSetAllParamsChunk = 0xA3

	ReqSiggenSetConfig = 0xA4
	ReqSiggenGetConfig = 0xA5
	ReqSiggenControl   = 0xA6
	ReqSiggenGetStatus = 0xA7
	ReqSiggenGetCaps   = 0xA8

	SiggenCtlStop    SiggenControlAction = 0
	SiggenCtlStart   SiggenControlAction = 1
	SiggenCtlStopNow SiggenControlAction = 2

	SiggenFlagRaw    = 0x01
	SiggenFlagDecorr = 0x02
	SiggenFlagWalk   = 0x04

	SiggenStateIdle    SiggenState = 0
	SiggenStateFadeIn  SiggenState = 1
	SiggenStateRun     SiggenState = 2
	SiggenStateGap     SiggenState = 3
	SiggenStateFadeOut SiggenState = 4

	SiggenStopReasonNone      SiggenStopReason = 0
	SiggenStopReasonHost      SiggenStopReason = 1
	SiggenStopReasonCompleted SiggenStopReason = 2
	SiggenStopReasonPreset    SiggenStopReason = 3
	SiggenStopReasonReconfig  SiggenStopReason = 4

	SiggenParamUnused  SiggenParamSemantic = 0
	SiggenParamFreqHz  SiggenParamSemantic = 1
	SiggenParamMs      SiggenParamSemantic = 2
	SiggenParamCycles  SiggenParamSemantic = 3
	SiggenParamCount   SiggenParamSemantic = 4
	SiggenParamRatio   SiggenParamSemantic = 5
	SiggenParamPattern SiggenParamSemantic = 6

	SiggenTimingContinuous SiggenTimingModel = 0
	SiggenTimingSweep      SiggenTimingModel = 1
	SiggenTimingPattern    SiggenTimingModel = 2

	InputSourceUSB    = 0
	InputSourceSPDIF  = 1
	InputSourceI2S    = 2
	InputSourceADAT   = 3
	InputSourceSPDIF2 = 4
	InputSourceSPDIF3 = 5

	// ADAT input clock modes (see AdatClockModeMaster/Slave constants).
	AdatClockModeMaster = 0
	AdatClockModeSlave  = 1

	// ADAT input RX GPIO sentinel meaning "unset".
	AdatInputPinUnset = 0xFF

	PinConfigSuccess      = 0x00
	PinConfigInvalidPin   = 0x01
	PinConfigPinInUse     = 0x02
	PinConfigOutputActive = 0x04
)
