# dspictl Codebase Map

**Purpose.** This is the "where is everything?" guide for `dspictl`. If you are new and want to find the code behind a feature, start here. Every entry points you at a real file and an approximate line.

**How to read the references.** Locations are written as `file.go:LINE`. Line numbers are a *snapshot* and drift as code changes; the **function/struct name is the durable anchor**. If a line is off by a few, search for the symbol.

The repository is a Go module (`github.com/suhlig/dspi`). The root package is the library; `cmd/dspictl` is the CLI binary.

---

## 1. Project Layout

| What | Where |
|---|---|
| Go module definition | `go.mod:1` |
| Library root package | all `*.go` files at repository root except `cmd/` and `*_test.go` |
| CLI entry point | `cmd/dspictl/main.go:20` (`main`), `cmd/dspictl/main.go:31` (`mainE`) |
| CLI command tree | `cmd/dspictl/main.go:107` (`newRootCmd`) |
| Device opening helper used by commands | `cmd/dspictl/main.go:432` (`openDevices`) |
| Interactive TUI mixer | `cmd/dspictl/mixer/` |
| Man page source (embedded at build time) | `man/dspictl.md` |
| Man page generator command | `cmd/dspictl/man.go:13` (`newManCmd`) |
| Release/changelog config | `cliff.toml` |
| Linux build container | `Dockerfile.linux` |

---

## 2. CLI Commands

All commands live under `cmd/dspictl/`. Each file is a thin wrapper around the library. A typical command opens devices via `openDevices()`, calls a library method, and prints the result.

| Command | File | Registration | Primary run function |
|---|---|---|---|
| Root / `--target` | `main.go` | `main.go:107` (`newRootCmd`) | `main.go:432` (`openDevices`) |
| `status` | `status.go` | `status.go:11` (`newStatusCmd`) | `status.go:19` (`runStatus`) |
| `volume` | `volume.go` | `volume.go:12` (`newVolumeCmd`) | `volume.go:75` (`runVolumeMode`), `volume.go:164` (`runVolumeGet`), `volume.go:188` (`runVolumeSet`) |
| `preamp` | `preamp.go` | `preamp.go:11` (`newPreampCmd`) | (per-channel and global wrappers) |
| `channel-name` | `channelname.go` | `channelname.go:11` (`newChannelNameCmd`) | `channelname.go:89` (`runChannelNameGet`), `channelname.go:139` (`runChannelNameSet`) |
| `channel` | `channel.go` | `channel.go:11` (`newChannelCmd`) | (gain/mute/delay for individual channels) |
| `eq` | `eq.go` | `eq.go:13` (`newEQCmd`) | `eq.go:26` (`newEQMasterCmd`), `eq.go:88` (`newEQOutputCmd`) |
| `crossover` | `crossover.go` | `crossover.go:11` (`newCrossoverCmd`) | `crossover.go:14` (`newCrossoverListCmd`) |
| `matrix` | `matrix.go` | `matrix.go:12` (`newMatrixCmd`) | `matrix.go:107` (`runMatrixList`), `matrix.go:157` (`runMatrixSet`) |
| `preset` | `preset.go` | `preset.go:11` (`newPresetCmd`) | `preset.go:206` (`runPresetList`), `preset.go:271` (`runPresetSave`) |
| `config` | `config.go` | `config.go:14` (`newConfigCmd`) | `config.go:139` (`runConfigExport`), `config.go:165` (`runConfigImport`), `config.go:713` (`runConfigSave`) |
| `input` | `input.go` | `input.go:11` (`newInputCmd`) | (input source, rate, I2S/ADAT settings) |
| `output` | `output.go` | `output.go:11` (`newOutputCmd`) | (output gain/mute/delay/enable) |
| `loudness` | `loudness.go` | `loudness.go:11` (`newLoudnessCmd`) | (loudness compensation settings) |
| `crossfeed` | `crossfeed.go` | `crossfeed.go:11` (`newCrossfeedCmd`) | (headphone crossfeed) |
| `leveller` | `leveller.go` | `leveller.go:11` (`newLevellerCmd`) | (dynamic range compressor) |
| `psybass` | `psybass.go` | `psybass.go:11` (`newPsybassCmd`) | (psychoacoustic bass) |
| `dac-mute` | `dacmute.go` | `dacmute.go:11` (`newDACMuteCmd`) | (DAC hardware mute) |
| `lg-sound-sync` | `lgsound.go` | `lgsound.go:11` (`newLGSoundSyncCmd`) | (LG TV volume sync) |
| `siggen` | `siggen.go` | `siggen.go:11` (`newSiggenCmd`) | (signal generator) |
| `upmix` | `upmix.go` | `upmix.go:11` (`newUpmixCmd`) | (stereo upmixer) |
| `adat` | `adatoutput.go` | `adatoutput.go:12` (`newAdatOutputCmd`) | (ADAT bulk output) |
| `ctrl` | `ctrl.go` | `ctrl.go:11` (`newCtrlCmd`) | (UART/I2C control interfaces) |
| `cs` | `cs.go` | `cs.go:11` (`newCsCmd`) | (Control Surfaces) |
| `diagnostics` | `diagnostics.go` | `diagnostics.go:11` (`newDiagnosticsCmd`) | `diagnostics.go:92` (`runDiagnosticsBufferStats`), `diagnostics.go:128` (`runDiagnosticsUSBErrors`) |
| `firmware` | `firmware.go` | `firmware.go:17` (`newFirmwareCmd`) | `firmware.go:41` (`newFirmwareVersionCmd`), `firmware.go:89` (`newFirmwareUpgradeCmd`) |
| `factory-reset` | `factoryreset.go` | `factoryreset.go:11` (`newFactoryResetCmd`) | `factoryreset.go:33` (`runFactoryReset`) |
| `mixer` | `mixer/cmd.go` | `mixer/cmd.go:10` (`NewCmd`) | `mixer/cmd.go:18` (`runMixer`) |
| `man` | `man.go` | `man.go:13` (`newManCmd`) | `man.go:23` (`runMan`) |

---

## 3. Core Device & USB Layer

| Concern | File | Anchor |
|---|---|---|
| USB control-transfer abstraction | `usb.go:4` | `USBControlTransfer` interface |
| Device handle | `device.go:37` | `Device` struct |
| Open a specific device by serial | `device.go:51` | `Open` |
| Open all connected devices | `device.go:111` | `OpenAll` |
| USB VID/PID, interface, request constants | `constants.go:3` | `dspiVID`, `dspiPID`, `Req*` constants |
| Enumerate connected devices without opening | `discovery.go:10` | `DeviceInfo`, `discovery.go:17` `List` |
| Construct a device for tests | `usb.go:18` | `NewDevice` |
| Platform detection at open time | `device.go:206` | `detectPlatform` |
| Platform constants & names | `platform.go:6` | `Platform`, `PlatformRP2040`, `PlatformRP2350` |
| Firmware version type | `platform.go:25` | `FirmwareVersion` |
| Gain type | `gain.go:6` | `Gain`, `gain.go:9` `NewGain` |
| Signal level type | `snapshot.go:9` | `Level`, `snapshot.go:12` `NewLevel` |
| Meter snapshot | `snapshot.go:43` | `MeterSnapshot` |
| Read meters / CPU / clip flags | `device.go:228` | `ReadMeter` |
| Clear clip latches | `device.go:285` | `ClearClips` |
| Read serial number | `device.go:180` | `GetSerial` |
| Read channel names | `device.go:336` | `ChannelName`, `device.go:359` `Channels` |
| Input channel count from bulk header | `device.go:423` | `NumInputChannels` |
| Channel grouping logic | `device.go:396` | `channelGroup` |
| Close device | `device.go:149` | `Close` |

---

## 4. Audio Features

### 4.1 Volume

| Feature | File | Anchor |
|---|---|---|
| Master volume set/get | `device.go:302` | `SetMasterVolume`, `device.go:319` `GetMasterVolume` |
| Master volume persistence mode | `mastervolume.go:11` | `SetMasterVolumeMode`, `mastervolume.go:27` `GetMasterVolumeMode` |
| Save master volume as boot default | `mastervolume.go:44` | `SaveMasterVolume` |
| Read saved boot-default volume | `mastervolume.go:60` | `GetSavedMasterVolume` |
| UAC/user volume | `uservolume.go:12` | `SetUserVolume`, `uservolume.go:29` `GetUserVolume` |

### 4.2 Preamp

| Feature | File | Anchor |
|---|---|---|
| Global preamp gain | `preamp_global.go:11` | `SetPreamp`, `preamp_global.go:28` `GetPreamp` |
| Per-channel preamp gain | `preamp.go:9` | `SetPreampChannel`, `preamp.go:25` `GetPreampChannel` |

### 4.3 Channels (gain, mute, delay, name)

| Feature | File | Anchor |
|---|---|---|
| Channel metadata | `channel.go:6` | `ChannelInfo` |
| Set channel name | `channel.go:15` | `SetChannelName` |
| Per-channel gain | `channel_controls.go:10` | `SetChannelGain`, `channel_controls.go:27` `GetChannelGain` |
| Per-channel mute | `channel_controls.go:43` | `SetChannelMute`, `channel_controls.go:63` `GetChannelMute` |
| Per-channel delay | `channel_controls.go:79` | `SetChannelDelay`, `channel_controls.go:96` `GetChannelDelay` |

### 4.4 Outputs

| Feature | File | Anchor |
|---|---|---|
| Output gain | `output.go:9` | `SetOutputGain`, `output.go:25` `GetOutputGain` |
| Output mute | `output.go:40` | `SetOutputMute`, `output.go:60` `GetOutputMute` |
| Output delay | `output.go:75` | `SetOutputDelay`, `output.go:91` `GetOutputDelay` |
| Output enable | `output.go:106` | `SetOutputEnable`, `output.go:126` `GetOutputEnable` |

### 4.5 EQ

| Feature | File | Anchor |
|---|---|---|
| Filter types | `eq.go:22` | `FilterType` |
| EQ band payload | `eq.go:121` | `EQBand` |
| EQ validation | `eq.go:132` | `EQBand.Validate` |
| Set EQ band | `eq.go:176` | `SetEQBand` |
| Get EQ band | `eq.go:226` | `GetEQBand` |
| Master EQ bypass | `eq.go:317` | `SetMasterEQBypass`, `eq.go:338` `GetMasterEQBypass` |
| Per-band bypass | `eq.go:392` | `SetBandBypass`, `eq.go:423` `GetBandBypass` |
| Max EQ channel per platform | `eq.go:354` | `MaxEQChannel` |
| Max active PEQ bands | `eq.go:366` | `MaxBands` |
| Band index validation | `eq.go:159` | `validateBandIndex` |
| CLI: `eq master` | `cmd/dspictl/eq.go:26` | `newEQMasterCmd` |
| CLI: `eq output` | `cmd/dspictl/eq.go:88` | `newEQOutputCmd` |

### 4.6 Crossover

| Feature | File | Anchor |
|---|---|---|
| Crossover filter types | `crossover.go:11` | `CrossoverFilterType` |
| Crossover band payload | `crossover.go:193` | `CrossoverBand` |
| Set crossover band | `crossover.go:228` | `SetCrossoverBand` |
| Get crossover band | `crossover.go:261` | `GetCrossoverBand` |
| Max crossover bands | `crossover.go:316` | `MaxCrossoverBands` |
| CLI: `crossover` | `cmd/dspictl/crossover.go:11` | `newCrossoverCmd` |

### 4.7 Matrix Mixer

| Feature | File | Anchor |
|---|---|---|
| Matrix route payload | `matrix.go:10` | `MatrixRoute` |
| Set matrix route | `matrix.go:18` | `SetMatrixRoute` |
| Get matrix route | `matrix.go:51` | `GetMatrixRoute` |
| CLI: `matrix` | `cmd/dspictl/matrix.go:12` | `newMatrixCmd` |

### 4.8 Presets

| Feature | File | Anchor |
|---|---|---|
| Preset directory metadata | `preset.go:10` | `PresetDirectory` |
| Save preset | `preset.go:19` | `PresetSave` |
| Load preset | `preset.go:38` | `PresetLoad` |
| Delete preset | `preset.go:57` | `PresetDelete` |
| Preset name get/set | `preset.go:76` | `GetPresetName`, `preset.go:97` `SetPresetName` |
| Preset directory / occupancy | `preset.go:113` | `GetPresetDirectory` |
| Active preset | `preset.go:137` | `GetActivePreset` |
| Startup mode / default slot | `preset.go:153` | `SetPresetStartup`, `preset.go:171` `GetPresetStartup` |
| Output config persistence mode | `preset.go:188` | `SetOutputConfigMode`, `preset.go:204` `GetOutputConfigMode` |
| Temporarily operate on a preset slot | `preset.go:223` | `WithPresetSlot`, `preset.go:230` `WithPresetSlotReadOnly` |
| CLI: `preset` | `cmd/dspictl/preset.go:11` | `newPresetCmd` |
| CLI: `preset eq` | `cmd/dspictl/preseteq.go:11` | `newPresetEQCmd` |
| CLI: `preset copy` | `cmd/dspictl/presetcopy.go:11` | `newPresetCopyCmd` |

### 4.9 Loudness Compensation

| Feature | File | Anchor |
|---|---|---|
| Enable / status | `loudness.go:10` | `SetLoudness`, `loudness.go:29` `GetLoudness` |
| Reference SPL | `loudness.go:44` | `SetLoudnessReference`, `loudness.go:60` `GetLoudnessReference` |
| Intensity | `loudness.go:75` | `SetLoudnessIntensity`, `loudness.go:91` `GetLoudnessIntensity` |
| Output mask | `loudness.go:109` | `SetLoudnessOutputMask`, `loudness.go:127` `GetLoudnessOutputMask` |

### 4.10 Crossfeed

| Feature | File | Anchor |
|---|---|---|
| Crossfeed status struct | `crossfeed.go:10` | `CrossfeedStatus` |
| Enable / status | `crossfeed.go:19` | `SetCrossfeed`, `crossfeed.go:38` `GetCrossfeed` |
| Preset | `crossfeed.go:54` | `SetCrossfeedPreset`, `crossfeed.go:68` `GetCrossfeedPreset` |
| Frequency / feed / ITD | `crossfeed.go:83` | `SetCrossfeedFreq`, `crossfeed.go:99` `GetCrossfeedFreq` |
| Output pair mask | `crossfeed.go:181` | `SetCrossfeedOutputPairMask`, `crossfeed.go:195` `GetCrossfeedOutputPairMask` |

### 4.11 Leveller (Compressor)

| Feature | File | Anchor |
|---|---|---|
| Leveller status | `leveller.go:9` | `LevellerStatus` |
| Enable / amount / speed / max gain | `leveller.go:22` | `SetLeveller`, `leveller.go:56` `SetLevellerAmount`, etc. |
| Lookahead / gate | `leveller.go:147` | `SetLevellerLookahead`, `leveller.go:180` `SetLevellerGate` |
| Detector / apply masks | `leveller.go:215` | `SetLevellerMasks`, `leveller.go:229` `GetLevellerMasks` |

### 4.12 Psychoacoustic Bass

| Feature | File | Anchor |
|---|---|---|
| Params struct | `psybass.go:10` | `PsybassParams` |
| Enable / mask | `psybass.go:21` | `GetPsybass`, `psybass.go:42` `SetPsybass` |
| Cutoff / harmonics / drive / character / original | `psybass.go:91` | `GetPsybassCutoff`, `SetPsybassCutoff`, etc. |

### 4.13 DAC Hardware Mute

| Feature | File | Anchor |
|---|---|---|
| Config struct | `dacmute.go:9` | `DACHwMuteConfig` |
| Set / get / test | `dacmute.go:25` | `SetDACHwMute`, `dacmute.go:51` `GetDACHwMute`, `dacmute.go:75` `TestDACHwMute` |

### 4.14 LG Sound Sync

| Feature | File | Anchor |
|---|---|---|
| Status struct | `lgsound.go:8` | `LGSoundSyncStatus` |
| Enable / status | `lgsound.go:16` | `SetLGSoundSync`, `lgsound.go:35` `GetLGSoundSync`, `lgsound.go:58` `GetLGSoundSyncStatus` |

### 4.15 Signal Generator

| Feature | File | Anchor |
|---|---|---|
| Signal types | `siggen.go:11` | `SiggenType` |
| Config payload | `siggen.go:195` | `SiggenConfig` |
| Encode / decode config | `siggen.go:269` | `encodeSiggenConfig`, `siggen.go:291` `decodeSiggenConfig` |
| Set / get config | `siggen.go:381` | `SetSiggenConfig`, `siggen.go:397` `GetSiggenConfig` |
| Start / stop / stop-now | `siggen.go:434` | `SiggenStart`, `siggen.go:438` `SiggenStop`, `siggen.go:444` `SiggenStopNow` |
| Status / caps / type descriptors | `siggen.go:449` | `GetSiggenStatus`, `siggen.go:464` `GetSiggenCaps`, `siggen.go:479` `GetSiggenTypeDesc` |
| Status types | `siggen.go:115` | `SiggenState`, `SiggenStopReason`, `SiggenParamSemantic`, `SiggenTimingModel` |

### 4.16 Stereo Upmixer

| Feature | File | Anchor |
|---|---|---|
| Wire config packet | `upmix.go:12` | `UpmixConfigPacket` |
| Encode / decode config | `upmix.go:32` | `UpmixConfigPacket.Encode`, `upmix.go:57` `DecodeUpmixConfig` |
| Status packet | `upmix.go:120` | `UpmixStatus`, `DecodeUpmixStatus` |
| Set / get whole config | `upmix.go:215` | `SetUpmixConfig`, `upmix.go:226` `GetUpmixConfig` |
| Single-parameter set / get | `upmix.go:237` | `SetUpmixParam`, `upmix.go:264` `GetUpmixParam` |
| Live telemetry | `upmix.go:279` | `GetUpmixStatus` |
| Param / mode name helpers | `upmix.go:150` | `UpmixParamName`, `UpmixCenterModeName`, `UpmixSurroundModeName` |

### 4.17 ADAT Bulk Output

| Feature | File | Anchor |
|---|---|---|
| Status packet | `adatoutput.go:11` | `AdatOutputStatus`, `DecodeAdatOutputStatus` |
| Enable / pin / status | `adatoutput.go:32` | `SetAdatOutputEnable`, `adatoutput.go:63` `GetAdatOutputEnable`, `adatoutput.go:87` `SetAdatOutputPin`, `adatoutput.go:119` `GetAdatOutputStatus` |

### 4.18 Control Surfaces

| Feature | File | Anchor |
|---|---|---|
| Binding wire struct | `controlsurface.go:220` | `CsBinding` |
| IR command wire struct | `controlsurface.go:285` | `IrCommand` |
| Status / caps / noun descriptors | `controlsurface.go:330` | `CsStatusPacket`, `CsCapsHeader`, `CsNounDesc`, decode helpers |
| Binding set / get | `controlsurface.go:438` | `SetCsBinding`, `controlsurface.go:452` `GetCsBinding` |
| Caps / status reads | `controlsurface.go:468` | `GetCsCaps`, `controlsurface.go:487` `GetCsNounDesc`, `controlsurface.go:500` `GetCsStatus` |
| Slot names | `controlsurface.go:516` | `SetCsName`, `controlsurface.go:537` `GetCsName` |
| IR commands + learn | `controlsurface.go:556` | `SetCsIrCommand`, `controlsurface.go:580` `GetCsIrCommand`, `controlsurface.go:600` `CsIrLearnArm`/`Cancel`/`Read` |
| Persist / revert | `controlsurface.go:633` | `CsSave`, `CsRevert` |
| Name / parse helpers | `controlsurface.go:370` | `CsStatusName`, `CsTypeName`, `CsNounName`, `ParseCsType`, `ParseCsNoun`, `ParseCsAction` |

### 4.19 External Control Interfaces (UART / I2C)

| Feature | File | Anchor |
|---|---|---|
| UART config wire struct | `ctrlinterface.go:11` | `UartCtrlConfig` |
| I2C config wire struct | `ctrlinterface.go:48` | `I2cCtrlConfig` |
| Interface status | `ctrlinterface.go:85` | `CtrlIfaceStatus`, `DecodeCtrlIfaceStatus` |
| UART set / get | `ctrlinterface.go:110` | `SetUartConfig`, `ctrlinterface.go:128` `GetUartConfig` |
| I2C set / get | `ctrlinterface.go:146` | `SetI2CConfig`, `ctrlinterface.go:164` `GetI2CConfig` |
| Live status | `ctrlinterface.go:182` | `GetCtrlIfaceStatus` |

---

## 5. Hardware Configuration

| Feature | File | Anchor |
|---|---|---|
| Output type (S/PDIF / I2S) | `config.go:7` | `SetOutputType`, `config.go:23` `GetOutputType` |
| Output GPIO pin | `config.go:40` | `SetOutputPin`, `config.go:61` `GetOutputPin` |
| I2S BCK pin | `config.go:77` | `SetI2SBckPin`, `config.go:98` `GetI2SBckPin` |
| MCK enable / pin / multiplier | `config.go:114` | `SetMCKEnable`, `config.go:140` `GetMCKEnable`, `config.go:156` `SetMCKPin`, etc. |
| Save output config to flash | `config.go:215` | `SaveOutputConfig` |
| I2S RX pin (pair-aware) | `input.go:189` | `SetI2SRxPin`, `input.go:195` `SetI2SRxPinPair`, `input.go:218` `GetI2SRxPin` |
| I2S input channel count | `input.go:239` | `SetI2SInputChannels`, `input.go:255` `GetI2SInputChannels` |
| S/PDIF RX pin | `spdif.go:23` | `SetSpdifRxPin`, `spdif.go:43` `GetSpdifRxPin` |
| S/PDIF RX pin by input index | `spdif.go:46` | `SetSpdifRxPinForIndex`, `spdif.go:68` `GetSpdifRxPinForIndex` |
| S/PDIF optional-input enable | `spdif.go:89` | `SetSpdifInputEnable` |
| S/PDIF input inventory | `spdif.go:128` | `GetSpdifInputConfig` |
| S/PDIF RX status | `spdif.go:58` | `GetSpdifRxStatus` |
| S/PDIF RX channel status bytes | `spdif.go:98` | `GetSpdifRxChannelStatus` |
| ADAT input enable / pin / clock mode / status | `input.go:299` | `SetAdatInputEnable`, `input.go:325` `GetAdatInputEnable`, `input.go:342` `SetAdatInputPin`, etc. |
| I2S clock mode + slave status | `input.go:414` | `SetI2SClockMode`, `input.go:489` `GetI2SClockMode`, `input.go:448` `GetI2sSlaveStatus` |
| I2S clock-pin mode + slave BCK role | `config.go:98` | `SetI2SClockPinMode`, `config.go:124` `GetI2SClockPinMode`, `config.go:47` `SetI2SBckPinRole` |
| User mute (vendor channel) | `uservolume.go:44` | `SetUserMute`, `uservolume.go:68` `GetUserMute` |
| Input source | `input.go:119` | `SetInputSource`, `input.go:134` `GetInputSource` |
| Input rate (I2S) | `input.go:151` | `SetInputRate`, `input.go:169` `GetInputRate` |
| Sample rate (raw status) | `system.go:147` | `GetSampleRate` |
| Output config persistence mode | `preset.go:188` | `SetOutputConfigMode`, `preset.go:204` `GetOutputConfigMode` |
| CLI: `config` | `cmd/dspictl/config.go:14` | `newConfigCmd` |
| CLI: `input` | `cmd/dspictl/input.go:11` | `newInputCmd` |
| CLI: `output` | `cmd/dspictl/output.go:11` | `newOutputCmd` |

---

## 6. State Snapshots & Bulk Transfer

| Feature | File | Anchor |
|---|---|---|
| Bulk header layout | `bulk.go:78` | `BulkHeader` |
| Full state snapshot | `bulk.go:92` | `BulkParams` |
| Read full state from device | `bulk.go:100` | `GetAllParams` |
| Write full state to device | `bulk.go:149` | `SetAllParams` |
| Chunked transfer helper | `bulk.go:180` | `chunkedTransfer` |
| Parse header from raw bytes | `bulk.go:195` | `ParseBulkHeader` |
| Field registry (offset/size map) | `bulk.go:54` | `fieldRegistry` |
| Field offsets constants | `bulk.go:24` | `fieldHeader`, `fieldGlobal`, `fieldEQ`, etc. |
| Bulk wire payload size | `bulk.go:19` | `wireBulkSize` |
| Header / global / input config accessors | `bulk.go:229` | `GetU8`, `SetU8`, `bulk.go:247` `GetU16`, etc. |
| Input source accessors | `bulk.go:304` | `InputSource`, `bulk.go:310` `SetInputSource` |
| I2S config accessors | `bulk.go:315` | `I2SRxPin`, `bulk.go:327` `I2SInputRate`, `bulk.go:339` `I2SInputChannels` |
| Loudness / crossfeed / leveller masks | `bulk.go:351` | `LoudnessOutputMask`, `bulk.go:362` `CrossfeedOutputPairMask`, `bulk.go:373` `LevellerDetectorMask` |
| ADAT input accessors | `bulk.go:395` | `AdatInputPin`, `bulk.go:406` `AdatInputEnabledP1`, `bulk.go:417` `AdatInputClockModeP1` |
| Psybass accessors | `bulk.go:427` | `PsybassEnabled`, `bulk.go:433` `SetPsybassEnabled`, `bulk.go:443` `PsybassOutputMask`, etc. |
| Commit live state to flash | `system.go:164` | `SaveParams` |
| Factory reset (live state only) | `system.go:29` | `FactoryReset` |
| CLI: `config export` | `cmd/dspictl/config.go:139` | `runConfigExport` |
| CLI: `config import` | `cmd/dspictl/config.go:165` | `runConfigImport` |

---

## 7. Diagnostics & System

| Feature | File | Anchor |
|---|---|---|
| USB PHY error counters | `system.go:13` | `USBErrorStats`, `system.go:184` `GetUSBErrorStats` |
| Buffer fill statistics | `system.go:23` | `BufferStats`, `system.go:108` `GetBufferStats`, `system.go:125` `ResetBufferStats` |
| Core 1 mode / conflict | `system.go:74` | `GetCore1Mode`, `system.go:92` `GetCore1Conflict` |
| Enter bootloader for firmware update | `system.go:51` | `EnterBootloader` |
| CLI: `diagnostics` | `cmd/dspictl/diagnostics.go:11` | `newDiagnosticsCmd` |

---

## 8. Firmware Update

| Feature | File | Anchor |
|---|---|---|
| UF2 block size / magic | `uf2.go:11` | `UF2BlockSize`, `uf2MagicStart` |
| UF2 board family IDs | `uf2.go:17` | `UF2FamilyRP2040`, `UF2FamilyRP2350`, etc. |
| Parse UF2 file metadata | `uf2.go:30` | `ParseUF2` |
| Map UF2 family to platform | `uf2.go:73` | `PlatformForFamily` |
| Map platform to UF2 family | `uf2.go:85` | `FamilyForPlatform` |
| Full upgrade flow | `cmd/dspictl/firmware.go:103` | `runFirmwareUpgrade` |
| Wait for bootloader volume | `cmd/dspictl/firmware.go:228` | `waitForBootloaderVolume` |
| Copy UF2 to volume | `cmd/dspictl/firmware.go:324` | `copyUF2` |
| Wait for device to reappear | `cmd/dspictl/firmware.go:260` | `waitForDevice` |
| CLI: `firmware upgrade` | `cmd/dspictl/firmware.go:89` | `newFirmwareUpgradeCmd` |
| CLI: `firmware enter-bootloader` | `cmd/dspictl/firmware.go:49` | `runFirmwareEnterBootloader` |

---

## 9. Interactive Mixer TUI

Built with Bubble Tea v2 (`charm.land/bubbletea/v2`) and Lipgloss v2.

| Concern | File | Anchor |
|---|---|---|
| Mixer command entry | `cmd/dspictl/mixer/cmd.go:10` | `NewCmd` |
| TEA model | `cmd/dspictl/mixer/model.go:15` | `model` |
| Device manager | `cmd/dspictl/mixer/model.go:35` | `deviceManager` |
| Initialize / rescan devices | `cmd/dspictl/mixer/model.go:49` | `Initialize`, `model.go:71` `Resync` |
| Meter polling / clip processing | `cmd/dspictl/mixer/model.go:172` | `ReadMeter`, `model.go:177` `ProcessClips`, `model.go:199` `TickClipTimer` |
| Mute handling | `cmd/dspictl/mixer/model.go:224` | `ToggleMute`, `model.go:234` `ToggleMuteAll` |
| Master volume refresh | `cmd/dspictl/mixer/model.go:212` | `RefreshMasterVolume` |
| TEA update / key handling | `cmd/dspictl/mixer/update.go:12` | `Update`, `update.go:43` `handleKeyPress` |
| TEA view rendering | `cmd/dspictl/mixer/view.go:13` | `View` |
| UI helpers (groups, layout, sliders, colors) | `cmd/dspictl/mixer/ui/` | `ui/*.go` |
| Connect / tick / rescan commands | `cmd/dspictl/mixer/commands.go` | (assumed; see `cmd/dspictl/mixer/commands.go`) |

---

## 10. Tests

| Concern | File | Anchor |
|---|---|---|
| Library Ginkgo suite | `dspi_suite_test.go` | `TestDspi` |
| CLI Ginkgo suite | `cmd/dspictl/dspictl_suite_test.go` | `TestDspictl` |
| Mock USB control transfer | `mock_usb_test.go` | `mockControlTransfer` |
| Unit tests for each feature | `*_test.go` | `var _ = Describe(...)` blocks |
| Hardware tests (require real device) | `hardware_test.go`, `hardware_roundtrip_test.go`, `meter_test.go` | gated by `-tags=hwtest` |
| Platform-specific tests | `platform_test.go` | `Platform` string/parse tests |
| Negative-number CLI protection | `cmd/dspictl/protect_negative_args_test.go` | `protectNegativeArgs` tests |

---

## 11. Man Page & Build

| Concern | File | Anchor |
|---|---|---|
| Man page Markdown source | `man/dspictl.md` | (embedded via `go:embed` in `man/man.go`) |
| Man package embedding | `man/man.go` | `Content` |
| Render to troff | `cmd/dspictl/man.go:34` | `md2man.Render` |
| Hidden `man` command | `cmd/dspictl/man.go:13` | `newManCmd` |

---

## 12. Common Patterns

- **Adding a new device command**: add a request code in `constants.go`, add library methods in a new or existing root file, add CLI wiring in `cmd/dspictl/main.go:107` (`newRootCmd`), and add a command file under `cmd/dspictl/`.
- **Device methods all follow the same shape**: check `d.closed`, build a byte buffer, call `d.usb.ControlTransfer(...)` with `vendorInterfaceInRequest` / `vendorInterfaceOutRequest` from `device.go:17`.
- **CLI commands are thin**: they open devices via `openDevices` (`cmd/dspictl/main.go:432`), call one or more library methods, print results, and `defer closeDevices(devices)`.
- **State persistence**: live state is read/written in bulk via `bulk.go:100` (`GetAllParams`) / `bulk.go:149` (`SetAllParams`); committing to flash is `system.go:164` (`SaveParams`).
- **Negative dB values**: `cmd/dspictl/main.go:78` (`protectNegativeArgs`) inserts `--` before negative numbers so `pflag` does not treat them as shorthand flags.

## Maintaining This Document

- Update this document whenever major changes were made to the code base.
- **Line numbers drift.** If you do a large edit to a mapped file, re-anchor the affected rows (search by symbol, not line). Treat function/struct names as the source of truth.
