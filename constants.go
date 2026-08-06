package dspi

const (
	dspiVID         = 0x2e8b // Raspberry Pi (Pico vendor)
	dspiPID         = 0xfeaa // DSPi product ID
	vendorInterface = 2
	maxChannels     = 17

	// Stereo upmixer (RP2350 only; see upmix.go).  SET_CONFIG applies a whole
	// 44-byte UpmixConfigPacket atomically; SET/GET_PARAM use wValue = param id.
	ReqUpmixSetConfig = 0x4A
	ReqUpmixGetConfig = 0x4B
	ReqUpmixSetParam  = 0x4C
	ReqUpmixGetParam  = 0x4D
	ReqUpmixGetStatus = 0x4E

	// Upmixer parameter ids (wValue of REQ_UPMIX_SET/GET_PARAM).
	UpmixParamEnabled      = 0
	UpmixParamCenterMode   = 1
	UpmixParamSurroundMode = 2
	UpmixParamStrength     = 3
	UpmixParamCenterWidth  = 4
	UpmixParamThreshold    = 5
	UpmixParamAttack       = 6
	UpmixParamRelease      = 7
	UpmixParamDetHpf       = 8
	UpmixParamSurDelay     = 9
	UpmixParamSurHpf       = 10
	UpmixParamSurLpf       = 11
	UpmixParamDecorr       = 12
	UpmixParamPresence     = 13
	UpmixParamCount        = 14

	// Upmixer centre engine modes.
	UpmixCenterModePassive  = 0
	UpmixCenterModeAdaptive = 1
	UpmixCenterModeOff      = 2

	// Upmixer surround engine modes.
	UpmixSurroundModeOff      = 0
	UpmixSurroundModePassive  = 1
	UpmixSurroundModeAdaptive = 2

	// UpmixStatus.parked_reason values.
	UpmixParkedActive      = 0
	UpmixParkedDisabled    = 1
	UpmixParkedNotStereo   = 2
	UpmixParkedRateTooHigh = 3

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
	ReqSetUserMute   = 0xDC
	ReqGetUserMute   = 0xDD

	ReqGetBufferStats     = 0xB0
	ReqResetBufferStats   = 0xB1
	ReqGetUSBErrorStats   = 0xB2
	ReqResetUSBErrorStats = 0xB3

	// ADAT bulk output (RP2350 only; see adatoutput.go).
	ReqSetAdatOutputEnable = 0xCA
	ReqGetAdatOutputEnable = 0xCB
	ReqSetAdatOutputPin    = 0xCC
	ReqGetAdatOutputPin    = 0xCD
	ReqGetAdatOutputStatus = 0xCE

	ReqSaveOutputConfig = 0x52

	ReqGetSpdifRxStatus    = 0xE2
	ReqGetSpdifRxChStatus  = 0xE3
	ReqSetSpdifRxPin       = 0xE4
	ReqGetSpdifRxPin       = 0xE5
	ReqSetSpdifInputEnable = 0xE9
	ReqGetSpdifInputConfig = 0xEF

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

	// I2S clock mode / slave input status (see input.go).
	ReqSetI2SClockMode   = 0x88
	ReqGetI2SClockMode   = 0x89
	ReqGetI2SSlaveStatus = 0x8A

	// I2S clock-pin mode: unified vs split master/slave BCK pairs (see config.go).
	ReqSetI2SClockPinMode = 0xFE
	ReqGetI2SClockPinMode = 0xFF

	// I2S BCK pin roles carried in the REQ_SET/GET_I2S_BCK_PIN wValue high byte.
	I2SBckRoleMaster = 0
	I2SBckRoleSlave  = 1

	// I2S clock-pin modes (REQ_SET/GET_I2S_CLOCK_PIN_MODE).
	I2SClockPinModeUnified = 0
	I2SClockPinModeSplit   = 1

	// I2S clock modes (REQ_SET/GET_I2S_CLOCK_MODE).
	I2SClockModeMaster = 0
	I2SClockModeSlave  = 1

	// External control interfaces (UART / I2C target; see ctrlinterface.go).
	ReqSetUartConfig      = 0xF5
	ReqGetUartConfig      = 0xF6
	ReqSetI2CConfig       = 0xF7
	ReqGetI2CConfig       = 0xF8
	ReqGetCtrlIfaceStatus = 0xF9

	// Control Surfaces request codes (see controlsurface.go).
	ReqSetCsBinding   = 0x84
	ReqGetCsBinding   = 0x85
	ReqGetCsCaps      = 0x86
	ReqGetCsStatus    = 0x87
	ReqSetCsName      = 0x8B
	ReqGetCsName      = 0x8C
	ReqSetCsIrCommand = 0x8D
	ReqGetCsIrCommand = 0x8E
	ReqCsIrLearn      = 0x8F
	ReqCsSave         = 0x9D
	ReqCsRevert       = 0x9E

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
	InputSourceSPDIF4 = 6

	// ADAT input clock modes (see AdatClockModeMaster/Slave constants).
	AdatClockModeMaster = 0
	AdatClockModeSlave  = 1

	// ADAT input RX GPIO sentinel meaning "unset".
	AdatInputPinUnset = 0xFF

	PinConfigSuccess       = 0x00
	PinConfigInvalidPin    = 0x01
	PinConfigPinInUse      = 0x02
	PinConfigInvalidOutput = 0x03
	PinConfigOutputActive  = 0x04
	PinConfigInvalidParam  = 0x05
)
