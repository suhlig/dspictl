package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// SiggenType identifies one of the 15 built-in test signals.
type SiggenType int

const (
	SiggenTypeSine SiggenType = iota
	SiggenTypeSquare
	SiggenTypeWhite
	SiggenTypePink
	SiggenTypeSweepLog
	SiggenTypeSweepLin
	SiggenTypeSweepStep
	SiggenTypeImpulse
	SiggenTypeClicksAlt
	SiggenTypePolarity
	SiggenTypeToneBurst
	SiggenTypeTonePair
	SiggenTypeMultitone
	SiggenTypeISP
	SiggenTypeChannelID
)

// siggenTypeNames maps each type to the short name the firmware advertises in
// the signal-generator caps descriptor.
var siggenTypeNames = map[SiggenType]string{
	SiggenTypeSine:      "sine",
	SiggenTypeSquare:    "square",
	SiggenTypeWhite:     "white",
	SiggenTypePink:      "pink",
	SiggenTypeSweepLog:  "swp-log",
	SiggenTypeSweepLin:  "swp-lin",
	SiggenTypeSweepStep: "swp-stp",
	SiggenTypeImpulse:   "impulse",
	SiggenTypeClicksAlt: "clk-alt",
	SiggenTypePolarity:  "polarty",
	SiggenTypeToneBurst: "burst",
	SiggenTypeTonePair:  "tonpair",
	SiggenTypeMultitone: "multi",
	SiggenTypeISP:       "isp",
	SiggenTypeChannelID: "chan-id",
}

// siggenTypeAliases provides additional human-friendly names accepted by the CLI.
var siggenTypeAliases = map[string]SiggenType{
	"sine":       SiggenTypeSine,
	"square":     SiggenTypeSquare,
	"white":      SiggenTypeWhite,
	"pink":       SiggenTypePink,
	"sweep-log":  SiggenTypeSweepLog,
	"sweep-lin":  SiggenTypeSweepLin,
	"sweep-step": SiggenTypeSweepStep,
	"impulse":    SiggenTypeImpulse,
	"clicks-alt": SiggenTypeClicksAlt,
	"polarity":   SiggenTypePolarity,
	"tone-burst": SiggenTypeToneBurst,
	"tone-pair":  SiggenTypeTonePair,
	"multitone":  SiggenTypeMultitone,
	"isp":        SiggenTypeISP,
	"channel-id": SiggenTypeChannelID,
}

// SiggenTypeCount is the number of signal types supported by the firmware.
const SiggenTypeCount = 15

// String returns the short name for the signal type.
func (t SiggenType) String() string {
	if name, ok := siggenTypeNames[t]; ok {
		return name
	}

	return fmt.Sprintf("unknown(%d)", t)
}

// ParseSiggenType resolves a signal type by its short name, alias, or numeric ID.
func ParseSiggenType(s string) (SiggenType, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "" {
		return 0, fmt.Errorf("signal type is required")
	}

	// Try numeric ID first.
	var id int
	if _, err := fmt.Sscanf(s, "%d", &id); err == nil {
		if id >= 0 && id < SiggenTypeCount {
			return SiggenType(id), nil
		}

		return 0, fmt.Errorf("signal type ID %d out of range (0-%d)", id, SiggenTypeCount-1)
	}

	// Try short firmware name.
	for typ, name := range siggenTypeNames {
		if s == name {
			return typ, nil
		}
	}

	// Try aliases.
	if typ, ok := siggenTypeAliases[s]; ok {
		return typ, nil
	}

	return 0, fmt.Errorf("unknown signal type %q", s)
}

// SiggenState describes the run state of the signal generator.
type SiggenState int

func (s SiggenState) String() string {
	switch s {
	case SiggenStateIdle:
		return "idle"
	case SiggenStateFadeIn:
		return "fade-in"
	case SiggenStateRun:
		return "run"
	case SiggenStateGap:
		return "gap"
	case SiggenStateFadeOut:
		return "fade-out"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// SiggenStopReason describes why the generator last stopped.
type SiggenStopReason int

func (r SiggenStopReason) String() string {
	switch r {
	case SiggenStopReasonNone:
		return "none"
	case SiggenStopReasonHost:
		return "host"
	case SiggenStopReasonCompleted:
		return "completed"
	case SiggenStopReasonPreset:
		return "preset"
	case SiggenStopReasonReconfig:
		return "reconfig"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

// SiggenParamSemantic describes the meaning of a type parameter.
type SiggenParamSemantic int

func (s SiggenParamSemantic) String() string {
	switch s {
	case SiggenParamUnused:
		return "-"
	case SiggenParamFreqHz:
		return "Hz"
	case SiggenParamMs:
		return "ms"
	case SiggenParamCycles:
		return "cycles"
	case SiggenParamCount:
		return "count"
	case SiggenParamRatio:
		return "ratio"
	case SiggenParamPattern:
		return "pattern"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// SiggenTimingModel classifies how duration/repeat/gap are interpreted.
type SiggenTimingModel int

func (m SiggenTimingModel) String() string {
	switch m {
	case SiggenTimingContinuous:
		return "continuous"
	case SiggenTimingSweep:
		return "sweep"
	case SiggenTimingPattern:
		return "pattern"
	default:
		return fmt.Sprintf("unknown(%d)", m)
	}
}

// SiggenConfig is the 36-byte REQ_SIGGEN_SET_CONFIG payload.
type SiggenConfig struct {
	Version     uint8
	SignalType  SiggenType
	ChannelMask uint16
	InvertMask  uint16
	Flags       uint8
	LevelDB     float64
	DurationMs  uint32
	Repeat      uint16
	GapMs       uint16
	P1          float64
	P2          float64
	P3          float64
	P4          float64
}

// SetFlag is a convenience helper that builds the flags byte.
func (c *SiggenConfig) SetFlag(raw, decorr, walk bool) {
	var flags uint8

	if raw {
		flags |= SiggenFlagRaw
	}

	if decorr {
		flags |= SiggenFlagDecorr
	}

	if walk {
		flags |= SiggenFlagWalk
	}

	c.Flags = flags
}

// SiggenStatus is the 16-byte REQ_SIGGEN_GET_STATUS response.
type SiggenStatus struct {
	Version       uint8
	State         SiggenState
	SignalType    SiggenType
	ActiveChannel int // 0xFF means not walking
	ElapsedMs     uint32
	CyclesDone    uint16
	StopReason    SiggenStopReason
	CurrentFreq   float64
}

// SiggenCaps is the 8-byte REQ_SIGGEN_GET_CAPS header response.
type SiggenCaps struct {
	Version          uint8
	TypeCount        uint8
	OutputChannels   uint8
	MultitoneMax     uint8
	ValidChannelMask uint16
}

// SiggenParamDesc is a single parameter descriptor from the caps response.
type SiggenParamDesc struct {
	Semantic SiggenParamSemantic
	Min      float64
	Max      float64
	Default  float64
}

// SiggenTypeDesc is the 62-byte per-type caps response.
type SiggenTypeDesc struct {
	ID          SiggenType
	Name        string
	TimingModel SiggenTimingModel
	Params      [4]SiggenParamDesc
}

// encodeSiggenConfig packs a SiggenConfig into the 36-byte wire format.
func encodeSiggenConfig(cfg *SiggenConfig) []byte {
	buf := make([]byte, 36)

	buf[0] = 1 // SIGGEN_CFG_VERSION
	buf[1] = byte(cfg.SignalType)
	binary.LittleEndian.PutUint16(buf[2:4], cfg.ChannelMask)
	binary.LittleEndian.PutUint16(buf[4:6], cfg.InvertMask)
	buf[6] = cfg.Flags
	// buf[7] reserved, already zero
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(float32(cfg.LevelDB)))
	binary.LittleEndian.PutUint32(buf[12:16], cfg.DurationMs)
	binary.LittleEndian.PutUint16(buf[16:18], cfg.Repeat)
	binary.LittleEndian.PutUint16(buf[18:20], cfg.GapMs)
	binary.LittleEndian.PutUint32(buf[20:24], math.Float32bits(float32(cfg.P1)))
	binary.LittleEndian.PutUint32(buf[24:28], math.Float32bits(float32(cfg.P2)))
	binary.LittleEndian.PutUint32(buf[28:32], math.Float32bits(float32(cfg.P3)))
	binary.LittleEndian.PutUint32(buf[32:36], math.Float32bits(float32(cfg.P4)))

	return buf
}

// decodeSiggenConfig unpacks a 36-byte wire buffer into a SiggenConfig.
func decodeSiggenConfig(buf []byte) (*SiggenConfig, error) {
	if len(buf) < 36 {
		return nil, fmt.Errorf("SiggenConfig too short: got %d bytes, need 36", len(buf))
	}

	cfg := &SiggenConfig{
		Version:     buf[0],
		SignalType:  SiggenType(buf[1]),
		ChannelMask: binary.LittleEndian.Uint16(buf[2:4]),
		InvertMask:  binary.LittleEndian.Uint16(buf[4:6]),
		Flags:       buf[6],
		LevelDB:     float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[8:12]))),
		DurationMs:  binary.LittleEndian.Uint32(buf[12:16]),
		Repeat:      binary.LittleEndian.Uint16(buf[16:18]),
		GapMs:       binary.LittleEndian.Uint16(buf[18:20]),
		P1:          float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[20:24]))),
		P2:          float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[24:28]))),
		P3:          float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[28:32]))),
		P4:          float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[32:36]))),
	}

	return cfg, nil
}

// decodeSiggenStatus unpacks a 16-byte wire buffer into a SiggenStatus.
func decodeSiggenStatus(buf []byte) (*SiggenStatus, error) {
	if len(buf) < 16 {
		return nil, fmt.Errorf("SiggenStatus too short: got %d bytes, need 16", len(buf))
	}

	active := int(buf[3])
	if buf[3] == 0xFF {
		active = -1
	}

	return &SiggenStatus{
		Version:       buf[0],
		State:         SiggenState(buf[1]),
		SignalType:    SiggenType(buf[2]),
		ActiveChannel: active,
		ElapsedMs:     binary.LittleEndian.Uint32(buf[4:8]),
		CyclesDone:    binary.LittleEndian.Uint16(buf[8:10]),
		StopReason:    SiggenStopReason(buf[10]),
		CurrentFreq:   float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[12:16]))),
	}, nil
}

// decodeSiggenCaps unpacks the 8-byte caps header.
func decodeSiggenCaps(buf []byte) (*SiggenCaps, error) {
	if len(buf) < 8 {
		return nil, fmt.Errorf("SiggenCaps too short: got %d bytes, need 8", len(buf))
	}

	return &SiggenCaps{
		Version:          buf[0],
		TypeCount:        buf[1],
		OutputChannels:   buf[2],
		MultitoneMax:     buf[3],
		ValidChannelMask: binary.LittleEndian.Uint16(buf[4:6]),
	}, nil
}

// decodeSiggenTypeDesc unpacks a 62-byte per-type descriptor.
func decodeSiggenTypeDesc(buf []byte) (*SiggenTypeDesc, error) {
	if len(buf) < 62 {
		return nil, fmt.Errorf("SiggenTypeDesc too short: got %d bytes, need 62", len(buf))
	}

	desc := &SiggenTypeDesc{
		ID:          SiggenType(buf[0]),
		Name:        strings.TrimRight(string(buf[1:9]), "\x00"),
		TimingModel: SiggenTimingModel(buf[9]),
	}

	off := 10

	for i := range 4 {
		p := &desc.Params[i]
		p.Semantic = SiggenParamSemantic(buf[off])
		p.Min = float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[off+1 : off+5])))
		p.Max = float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[off+5 : off+9])))
		p.Default = float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[off+9 : off+13])))
		off += 13
	}

	return desc, nil
}

// SetSiggenConfig stages a 36-byte SiggenConfig on the device.
// The generator does not start automatically; use SiggenStart afterwards.
func (d *Device) SetSiggenConfig(cfg *SiggenConfig) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	data := encodeSiggenConfig(cfg)

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSiggenSetConfig, 0, vendorInterface, data)
	if err != nil {
		return fmt.Errorf("REQ_SIGGEN_SET_CONFIG: %w", err)
	}

	return nil
}

// GetSiggenConfig reads the currently applied signal-generator config.
func (d *Device) GetSiggenConfig() (*SiggenConfig, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 36)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSiggenGetConfig, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_SIGGEN_GET_CONFIG: %w", err)
	}

	return decodeSiggenConfig(buf)
}

// SiggenControlAction selects a generator control command.
type SiggenControlAction int

// SiggenControl sends a control action (start / stop / stop-now) to the device.
func (d *Device) SiggenControl(action SiggenControlAction) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSiggenControl, uint16(action), vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_SIGGEN_CONTROL: %w", err)
	}

	if len(buf) > 0 && buf[0] != 1 {
		return fmt.Errorf("REQ_SIGGEN_CONTROL: unexpected status 0x%02X", buf[0])
	}

	return nil
}

// SiggenStart starts or restarts the generator with the applied config.
func (d *Device) SiggenStart() error {
	return d.SiggenControl(SiggenControlAction(SiggenCtlStart))
}

// SiggenStop fades the generator out and stops it.
func (d *Device) SiggenStop() error {
	return d.SiggenControl(SiggenControlAction(SiggenCtlStop))
}

// SiggenStopNow performs an immediate hard stop.
func (d *Device) SiggenStopNow() error {
	return d.SiggenControl(SiggenControlAction(SiggenCtlStopNow))
}

// GetSiggenStatus reads the current generator state.
func (d *Device) GetSiggenStatus() (*SiggenStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSiggenGetStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_SIGGEN_GET_STATUS: %w", err)
	}

	return decodeSiggenStatus(buf)
}

// GetSiggenCaps reads the generator capability header.
func (d *Device) GetSiggenCaps() (*SiggenCaps, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSiggenGetCaps, 0xFFFF, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_SIGGEN_GET_CAPS: %w", err)
	}

	return decodeSiggenCaps(buf)
}

// GetSiggenTypeDesc reads the 62-byte descriptor for the given signal type index.
func (d *Device) GetSiggenTypeDesc(index int) (*SiggenTypeDesc, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	if index < 0 || index >= SiggenTypeCount {
		return nil, fmt.Errorf("signal type index %d out of range (0-%d)", index, SiggenTypeCount-1)
	}

	buf := make([]byte, 62)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqSiggenGetCaps, uint16(index), vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_SIGGEN_GET_CAPS(%d): %w", index, err)
	}

	return decodeSiggenTypeDesc(buf)
}
