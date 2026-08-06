package dspi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Control Surfaces: user-wired physical controls and indicators (buttons,
// switches, pots, rotary encoders, LEDs, IR receivers) on spare GPIOs, each
// bound to one firmware parameter.  Config is device-global, configured over
// USB; SETs are apply-live-only previews whose outcome is polled back via
// GetCsStatus, then persisted with CsSave (see control_surfaces_spec.md).
// Capability format v7 (v1.1.5): 16 IR sub-slots (caps v6) and the loudness
// reference SPL / intensity nouns (v7).

// CS component types (CsBinding.type).
const (
	CsTypeNone    = 0
	CsTypeButton  = 1
	CsTypeSwitch  = 2
	CsTypePot     = 3
	CsTypeEncoder = 4
	CsTypeLED     = 5
	CsTypeLEDPWM  = 6 // hardware-PWM-dimmed LED (IND_LEVEL meter)
	CsTypeIR      = 7 // IR remote receiver (container: one pin + command sub-slots)
	CsTypeCount   = 8
)

// CS nouns (CsBinding.noun / IrCommand.noun).  The live count and per-noun
// descriptors come from the caps; these mirror the firmware's CsNoun enum.
const (
	CsNounUserVolume        = 0
	CsNounMasterVolume      = 1
	CsNounUserMute          = 2
	CsNounLoudness          = 3
	CsNounCrossfeed         = 4
	CsNounLeveller          = 5
	CsNounPreset            = 6
	CsNounInputSource       = 7
	CsNounClip              = 8
	CsNounEQBypass          = 9
	CsNounLgSync            = 10
	CsNounCrossfeedPreset   = 11
	CsNounCrossfeedITD      = 12
	CsNounLevellerAmount    = 13
	CsNounLevellerSpeed     = 14
	CsNounLevellerLookahead = 15
	CsNounPreamp            = 16
	CsNounOutputGain        = 17
	CsNounOutputMute        = 18
	CsNounOutputEnable      = 19
	CsNounFilterFreq        = 20
	CsNounFilterGain        = 21
	CsNounFilterQ           = 22
	CsNounFilterType        = 23
	CsNounFilterBypass      = 24
	CsNounSiggen            = 25
	CsNounDacMuteTest       = 26
	CsNounClipCh            = 27
	CsNounLevel             = 28
	CsNounSpdifLock         = 29
	CsNounSampleRate        = 30
	CsNounUsbStreaming      = 31
	CsNounAdatActive        = 32
	CsNounLgPresent         = 33
	CsNounLgMuted           = 34
	CsNounUpmix             = 35
	CsNounUpmixCenterMode   = 36
	CsNounUpmixSurroundMode = 37
	CsNounUpmixStrength     = 38
	CsNounUpmixWidth        = 39
	CsNounUpmixPresence     = 40
	CsNounPsybass           = 41
	CsNounPsybassCutoff     = 42
	CsNounPsybassHarmonics  = 43
	CsNounPsybassDrive      = 44
	CsNounPsybassCharacter  = 45
	CsNounPsybassOriginal   = 46
	CsNounOutputDelay       = 47
	CsNounPresetReload      = 48
	CsNounLoudnessSPL       = 49 // caps v7
	CsNounLoudnessIntensity = 50 // caps v7
	CsNounCount             = 51
)

// CS actions (CsBinding.action / IrCommand.action).
const (
	CsActAdjust    = 0  // pot: absolute position maps onto a value range
	CsActStep      = 1  // encoder: +/- step per detent
	CsActInc       = 2  // button: + step per press
	CsActDec       = 3  // button: - step per press
	CsActToggle    = 4  // button: invert a bool per press
	CsActSet       = 5  // button: set the noun to `value` per press
	CsActFollow    = 6  // switch: bool tracks the switch position
	CsActTrigger   = 7  // button: fire the noun's command (e.g. clip clear)
	CsActIndEquals = 8  // LED: lit while noun value == `value`
	CsActMomentary = 9  // button (press): set to `value` while held
	CsActIndAbove  = 10 // LED: lit while a continuous noun value >= `value`
	CsActIndLevel  = 11 // PWM LED: brightness follows the noun across its range
	CsActCount     = 12
)

// CS binding flags (CsBinding.flags).
const (
	CsFlagInvert  = 0x01 // input active-high w/ pull-down; LED active-low
	CsFlagReverse = 0x02 // pot/encoder: invert direction
	CsFlagWrap    = 0x04 // enum STEP/INC/DEC wraps around the ends
	CsFlagAccel   = 0x08 // encoder: fast rotation multiplies the step
	CsFlagRepeat  = 0x10 // button INC/DEC: auto-repeat while held
)

// CS binding events (buttons only; MUST be 0 for other types).
const (
	CsEventPress  = 0 // short press
	CsEventLong   = 1 // held >= 500 ms
	CsEventDouble = 2 // two taps within 350 ms
)

// CS status codes.  0x00..0x05 reuse the shared PIN_CONFIG_* namespace;
// Control Surfaces extends it from 0x10.
const (
	CsStatusInvalidSlot   = 0x10
	CsStatusInvalidType   = 0x11
	CsStatusInvalidNoun   = 0x12
	CsStatusInvalidAction = 0x13
	CsStatusInvalidValue  = 0x14
	CsStatusPinNotADC     = 0x15
	CsStatusPending       = 0x16
	CsStatusInvalidTarget = 0x17
	CsStatusInvalidEvent  = 0x18
	CsStatusPWMConflict   = 0x19
	CsStatusEventInUse    = 0x1A
	CsStatusBusy          = 0x1B
	CsStatusFlashError    = 0x1C
	CsStatusIRInUse       = 0x1D
	CsStatusNoIR          = 0x1E
)

// CS noun descriptor kinds (CsNounDesc.kind).
const (
	CsKindContinuous = 0
	CsKindBool       = 1
	CsKindEnum       = 2
)

// CS noun descriptor units (CsNounDesc.unit).
const (
	CsUnitNone    = 0
	CsUnitDB      = 1 // signed 8.8 dB
	CsUnitHz      = 2 // plain integer Hz, log stepping
	CsUnitQ       = 3 // 8.8 Q, log stepping
	CsUnitPercent = 4 // 8.8 percent
	CsUnitMs      = 5 // 8.8 ms
)

// CS noun descriptor target kinds (CsNounDesc.target_kind).
const (
	CsTargetNone     = 0
	CsTargetInputCh  = 1
	CsTargetOutputCh = 2
	CsTargetDspCh    = 3
	CsTargetDspBand  = 4
)

// CS IR code protocols (IrCommand.protocol).
const (
	CsIRProtoNone = 0
	CsIRProtoNEC  = 1
	CsIRProtoRC5  = 2
	CsIRProtoRC6  = 3
	CsIRProtoHash = 4
)

// CS IR learn states (CsStatusPacket.ir_learn_state / learn result).
const (
	CsIrLearnIdle    = 0
	CsIrLearnArmed   = 1
	CsIrLearnDone    = 2
	CsIrLearnTimeout = 3
)

const (
	CsMaxBindings    = 16
	CsMaxIRCommands  = 16
	CsNameLen        = 32
	CsGPIOUnused     = 0xFF
	CsStatusLen      = 41 // caps v6+: 41-byte status packet
	CsCapsAll        = 0xFFFF
	CsLastSlotSave   = 0xFF // lastSlot sentinel for a save/revert outcome
	CsLastSlotIRFlag = 0x80 // lastSlot high bit marks an IR sub-slot outcome
)

// CsBinding is one control-surface binding; 24 bytes on the wire
// (REQ_SET/GET_CS_BINDING).  value/step/range encoding follows the noun's
// unit; a CsTypeIR binding is a container whose commands live in the
// IrCommand table.
type CsBinding struct {
	Type     uint8 // CsType*
	Noun     uint8 // CsNoun*
	Action   uint8 // CsAct*
	Flags    uint8 // CsFlag*
	GPIO     [2]uint8
	Event    uint8 // CsEvent* (buttons; 0 otherwise)
	Target   uint8 // channel index for targeted nouns (else 0)
	Index    uint8 // filter band for CsTargetDspBand nouns (else 0)
	Value    int16
	Step     int16
	RangeMin int16
	RangeMax int16
}

// Encode serialises the binding into the 24-byte wire layout.
func (b *CsBinding) Encode() []byte {
	buf := make([]byte, 24)
	buf[0] = b.Type
	buf[1] = b.Noun
	buf[2] = b.Action
	buf[3] = b.Flags
	buf[4] = b.GPIO[0]
	buf[5] = b.GPIO[1]
	buf[6] = b.Event
	buf[7] = b.Target
	buf[8] = b.Index
	buf[9] = 0 // reserved
	binary.LittleEndian.PutUint16(buf[10:12], uint16(b.Value))
	binary.LittleEndian.PutUint16(buf[12:14], uint16(b.Step))
	binary.LittleEndian.PutUint16(buf[14:16], uint16(b.RangeMin))
	binary.LittleEndian.PutUint16(buf[16:18], uint16(b.RangeMax))
	// bytes 18..23 reserved, zero
	return buf
}

// DecodeCsBinding parses a 24-byte binding payload.
func DecodeCsBinding(raw []byte) (*CsBinding, error) {
	if len(raw) < 24 {
		return nil, fmt.Errorf("CS binding payload too short (got %d, need 24)", len(raw))
	}
	return &CsBinding{
		Type:     raw[0],
		Noun:     raw[1],
		Action:   raw[2],
		Flags:    raw[3],
		GPIO:     [2]uint8{raw[4], raw[5]},
		Event:    raw[6],
		Target:   raw[7],
		Index:    raw[8],
		Value:    int16(binary.LittleEndian.Uint16(raw[10:12])),
		Step:     int16(binary.LittleEndian.Uint16(raw[12:14])),
		RangeMin: int16(binary.LittleEndian.Uint16(raw[14:16])),
		RangeMax: int16(binary.LittleEndian.Uint16(raw[16:18])),
	}, nil
}

// IrCommand is one IR remote command; 16 bytes on the wire
// (REQ_SET/GET_CS_IR_CMD).  Protocol CsIRProtoNone marks the sub-slot empty.
type IrCommand struct {
	Noun     uint8 // CsNoun*
	Action   uint8 // CsAct* (button subset)
	Flags    uint8 // CsFlagWrap | CsFlagRepeat
	Target   uint8
	Index    uint8
	Protocol uint8 // CsIRProto*
	Value    int16
	Step     int16
	Code     uint32
}

// Encode serialises the command into the 16-byte wire layout.
func (c *IrCommand) Encode() []byte {
	buf := make([]byte, 16)
	buf[0] = c.Noun
	buf[1] = c.Action
	buf[2] = c.Flags
	buf[3] = c.Target
	buf[4] = c.Index
	buf[5] = c.Protocol
	binary.LittleEndian.PutUint16(buf[6:8], uint16(c.Value))
	binary.LittleEndian.PutUint16(buf[8:10], uint16(c.Step))
	binary.LittleEndian.PutUint32(buf[12:16], c.Code)
	return buf
}

// DecodeIrCommand parses a 16-byte IR command payload.
func DecodeIrCommand(raw []byte) (*IrCommand, error) {
	if len(raw) < 16 {
		return nil, fmt.Errorf("CS IR command payload too short (got %d, need 16)", len(raw))
	}
	return &IrCommand{
		Noun:     raw[0],
		Action:   raw[1],
		Flags:    raw[2],
		Target:   raw[3],
		Index:    raw[4],
		Protocol: raw[5],
		Value:    int16(binary.LittleEndian.Uint16(raw[6:8])),
		Step:     int16(binary.LittleEndian.Uint16(raw[8:10])),
		Code:     binary.LittleEndian.Uint32(raw[12:16]),
	}, nil
}

// CsStatusPacket is the 41-byte REQ_GET_CS_STATUS response (caps v6+).
// last_status/last_slot report the most recent deferred SET of any CS kind;
// IR command slots are reported as 0x80 | sub-slot.
type CsStatusPacket struct {
	LastStatus   uint8
	LastSlot     uint8
	MaxBindings  uint8
	Dirty        bool
	ActiveMask   uint16 // bit N = binding N live
	SlotStatus   [CsMaxBindings]uint8
	IRActiveMask uint16 // bit N = IR command N live (component up)
	IRLearnState uint8  // CsIrLearn*
	IRCmdStatus  [CsMaxIRCommands]uint8
}

// DecodeCsStatus parses a status payload.  Older firmware short-reads it
// (32 bytes for caps v3-v5, 22 pre-v3); fields beyond the returned bytes are
// left zero.
func DecodeCsStatus(raw []byte) (*CsStatusPacket, error) {
	pkt := &CsStatusPacket{}
	if len(raw) < 4 {
		return nil, fmt.Errorf("CS status payload too short (got %d, need >= 4)", len(raw))
	}
	pkt.LastStatus = raw[0]
	pkt.LastSlot = raw[1]
	pkt.MaxBindings = raw[2]
	pkt.Dirty = raw[3] != 0
	if len(raw) >= 6 {
		pkt.ActiveMask = binary.LittleEndian.Uint16(raw[4:6])
	}
	copy(pkt.SlotStatus[:], raw[6:])
	if len(raw) >= 6+CsMaxBindings+2 {
		ir := raw[6+CsMaxBindings : 6+CsMaxBindings+2]
		pkt.IRActiveMask = binary.LittleEndian.Uint16(ir)
	}
	if len(raw) >= 6+CsMaxBindings+2+1 {
		pkt.IRLearnState = raw[6+CsMaxBindings+2]
	}
	copy(pkt.IRCmdStatus[:], raw[6+CsMaxBindings+2+1:])
	return pkt, nil
}

// CsTypeDesc describes one component type in the caps type table.
type CsTypeDesc struct {
	Actions  uint16 // CsAct* bitmask this component can drive
	PinCount uint8  // GPIOs consumed (1 or 2)
	PinClass uint8  // 0 = any, 1 = ADC only
}

// CsCapsHeader is the 40-byte REQ_GET_CS_CAPS (wValue 0xFFFF) response:
// header + type table.
type CsCapsHeader struct {
	CapsVersion   uint8 // capability format version (7 on v1.1.5)
	MaxBindings   uint8
	TypeCount     uint8 // table follows, index = CsType*
	NounCount     uint8
	Types         [CsTypeCount]CsTypeDesc
	MaxIRCommands uint8
}

// DecodeCsCaps parses the 40-byte caps payload.
func DecodeCsCaps(raw []byte) (*CsCapsHeader, error) {
	if len(raw) < 37 {
		return nil, fmt.Errorf("CS caps payload too short (got %d, need >= 37)", len(raw))
	}
	h := &CsCapsHeader{
		CapsVersion:   raw[0],
		MaxBindings:   raw[1],
		TypeCount:     raw[2],
		NounCount:     raw[3],
		MaxIRCommands: raw[36],
	}
	for i := range CsTypeCount {
		off := 4 + i*4
		if off+4 > len(raw) {
			break
		}
		h.Types[i] = CsTypeDesc{
			Actions:  binary.LittleEndian.Uint16(raw[off : off+2]),
			PinCount: raw[off+2],
			PinClass: raw[off+3],
		}
	}
	return h, nil
}

// CsNounDesc is the 12-byte per-noun capability descriptor
// (REQ_GET_CS_CAPS with wValue = noun).
type CsNounDesc struct {
	Kind        uint8 // CsKind*
	EnumCount   uint8 // CsKindEnum only
	Actions     uint16
	MinQ        int16 // continuous range, unit-encoded
	MaxQ        int16
	Unit        uint8 // CsUnit*
	TargetKind  uint8 // CsTarget*
	TargetCount uint8
	DFlags      uint8
}

// DecodeCsNounDesc parses a 12-byte noun descriptor.
func DecodeCsNounDesc(raw []byte) (*CsNounDesc, error) {
	if len(raw) < 12 {
		return nil, fmt.Errorf("CS noun descriptor payload too short (got %d, need 12)", len(raw))
	}
	return &CsNounDesc{
		Kind:        raw[0],
		EnumCount:   raw[1],
		Actions:     binary.LittleEndian.Uint16(raw[2:4]),
		MinQ:        int16(binary.LittleEndian.Uint16(raw[4:6])),
		MaxQ:        int16(binary.LittleEndian.Uint16(raw[6:8])),
		Unit:        raw[8],
		TargetKind:  raw[9],
		TargetCount: raw[10],
		DFlags:      raw[11],
	}, nil
}

// CsIRLearnResult is the 8-byte REQ_CS_IR_LEARN (wValue 2) result.
type CsIRLearnResult struct {
	State    uint8 // CsIrLearn*
	Protocol uint8 // CsIRProto*
	Code     uint32
}

// DecodeCsIRLearnResult parses the 8-byte learn result.
func DecodeCsIRLearnResult(raw []byte) (*CsIRLearnResult, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("CS IR learn result too short (got %d, need 8)", len(raw))
	}
	return &CsIRLearnResult{
		State:    raw[0],
		Protocol: raw[1],
		Code:     binary.LittleEndian.Uint32(raw[4:8]),
	}, nil
}

// CsStatusName returns a human-readable name for a CS status code.
func CsStatusName(status uint8) string {
	switch status {
	case PinConfigSuccess:
		return "ok"
	case PinConfigInvalidPin:
		return "invalid pin"
	case PinConfigPinInUse:
		return "pin in use"
	case PinConfigInvalidOutput:
		return "invalid output"
	case PinConfigOutputActive:
		return "output active"
	case PinConfigInvalidParam:
		return "invalid parameter"
	case CsStatusInvalidSlot:
		return "invalid slot"
	case CsStatusInvalidType:
		return "invalid type"
	case CsStatusInvalidNoun:
		return "invalid noun"
	case CsStatusInvalidAction:
		return "invalid action"
	case CsStatusInvalidValue:
		return "invalid value"
	case CsStatusPinNotADC:
		return "pin not ADC"
	case CsStatusPending:
		return "pending"
	case CsStatusInvalidTarget:
		return "invalid target"
	case CsStatusInvalidEvent:
		return "invalid event"
	case CsStatusPWMConflict:
		return "PWM conflict"
	case CsStatusEventInUse:
		return "event in use"
	case CsStatusBusy:
		return "busy"
	case CsStatusFlashError:
		return "flash error"
	case CsStatusIRInUse:
		return "IR in use"
	case CsStatusNoIR:
		return "no IR"
	default:
		return fmt.Sprintf("unknown(0x%02X)", status)
	}
}

// CsTypeName returns the name of a CS component type.
func CsTypeName(t uint8) string {
	switch t {
	case CsTypeNone:
		return "none"
	case CsTypeButton:
		return "button"
	case CsTypeSwitch:
		return "switch"
	case CsTypePot:
		return "pot"
	case CsTypeEncoder:
		return "encoder"
	case CsTypeLED:
		return "led"
	case CsTypeLEDPWM:
		return "led-pwm"
	case CsTypeIR:
		return "ir"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// CsNounName returns the name of a CS noun.
func CsNounName(n uint8) string {
	names := [...]string{
		"user-volume", "master-volume", "user-mute", "loudness", "crossfeed",
		"leveller", "preset", "input-source", "clip", "eq-bypass", "lg-sync",
		"crossfeed-preset", "crossfeed-itd", "leveller-amount", "leveller-speed",
		"leveller-lookahead", "preamp", "output-gain", "output-mute", "output-enable",
		"filter-freq", "filter-gain", "filter-q", "filter-type", "filter-bypass",
		"siggen", "dac-mute-test", "clip-ch", "level", "spdif-lock",
		"sample-rate", "usb-streaming", "adat-active", "lg-present", "lg-muted",
		"upmix", "upmix-center-mode", "upmix-surround-mode", "upmix-strength",
		"upmix-width", "upmix-presence", "psybass", "psybass-cutoff",
		"psybass-harmonics", "psybass-drive", "psybass-character", "psybass-original",
		"output-delay", "preset-reload", "loudness-spl", "loudness-intensity",
	}
	if int(n) < len(names) {
		return names[n]
	}
	return fmt.Sprintf("unknown(%d)", n)
}

// ParseCsType resolves a CS component type by name or numeric id.
func ParseCsType(s string) (uint8, error) {
	names := map[string]uint8{
		"none": CsTypeNone, "button": CsTypeButton, "switch": CsTypeSwitch,
		"pot": CsTypePot, "encoder": CsTypeEncoder, "led": CsTypeLED,
		"led-pwm": CsTypeLEDPWM, "ir": CsTypeIR,
	}
	if t, ok := names[s]; ok {
		return t, nil
	}
	var id uint8
	if _, err := fmt.Sscanf(s, "%d", &id); err == nil && id < CsTypeCount {
		return id, nil
	}
	return 0, fmt.Errorf("unknown CS type %q", s)
}

// ParseCsNoun resolves a CS noun by name or numeric id.
func ParseCsNoun(s string) (uint8, error) {
	var id uint8
	if _, err := fmt.Sscanf(s, "%d", &id); err == nil && id < CsNounCount {
		return id, nil
	}
	for n := range uint8(CsNounCount) {
		if CsNounName(n) == s {
			return n, nil
		}
	}
	return 0, fmt.Errorf("unknown CS noun %q", s)
}

// ParseCsAction resolves a CS action by name or numeric id.
func ParseCsAction(s string) (uint8, error) {
	names := map[string]uint8{
		"adjust": CsActAdjust, "step": CsActStep, "inc": CsActInc,
		"dec": CsActDec, "toggle": CsActToggle, "set": CsActSet,
		"follow": CsActFollow, "trigger": CsActTrigger, "ind-equals": CsActIndEquals,
		"momentary": CsActMomentary, "ind-above": CsActIndAbove, "ind-level": CsActIndLevel,
	}
	if a, ok := names[s]; ok {
		return a, nil
	}
	var id uint8
	if _, err := fmt.Sscanf(s, "%d", &id); err == nil && id < CsActCount {
		return id, nil
	}
	return 0, fmt.Errorf("unknown CS action %q", s)
}

// SetCsBinding uploads a binding to a slot (REQ_SET_CS_BINDING).  The SET is
// an apply-live-only preview; the result lands in GetCsStatus.
func (d *Device) SetCsBinding(slot int, b *CsBinding) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	if slot < 0 || slot >= CsMaxBindings {
		return fmt.Errorf("slot %d out of range (0-%d)", slot, CsMaxBindings-1)
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCsBinding, uint16(slot), vendorInterface, b.Encode())
	if err != nil {
		return fmt.Errorf("REQ_SET_CS_BINDING: %w", err)
	}

	return nil
}

// GetCsBinding reads the live binding of a slot.
func (d *Device) GetCsBinding(slot int) (*CsBinding, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 24)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCsBinding, uint16(slot), vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_CS_BINDING: %w", err)
	}

	return DecodeCsBinding(buf[:n])
}

// GetCsCaps reads the capability header + type table (REQ_GET_CS_CAPS with
// wValue 0xFFFF).
func (d *Device) GetCsCaps() (*CsCapsHeader, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 40)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCsCaps, CsCapsAll, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_CS_CAPS: %w", err)
	}

	return DecodeCsCaps(buf[:n])
}

// GetCsNounDesc reads the capability descriptor of one noun.
func (d *Device) GetCsNounDesc(noun uint8) (*CsNounDesc, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 12)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCsCaps, uint16(noun), vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_CS_CAPS(noun): %w", err)
	}

	return DecodeCsNounDesc(buf[:n])
}

// GetCsStatus reads the 41-byte CS status packet (caps v6+).  Poll it after
// a deferred SET to learn the apply outcome.
func (d *Device) GetCsStatus() (*CsStatusPacket, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, CsStatusLen)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCsStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_CS_STATUS: %w", err)
	}

	return DecodeCsStatus(buf[:n])
}

// SetCsName sets the per-slot user label (REQ_SET_CS_NAME; wValue = slot,
// payload = 1-32 bytes; a single NUL byte clears it).  Apply-live-only
// preview; the result lands in GetCsStatus.
func (d *Device) SetCsName(slot int, name string) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	if slot < 0 || slot >= CsMaxBindings {
		return fmt.Errorf("slot %d out of range (0-%d)", slot, CsMaxBindings-1)
	}
	if len(name) >= CsNameLen {
		return fmt.Errorf("name too long (max %d bytes)", CsNameLen-1)
	}
	if name == "" {
		name = "\x00" // single NUL byte clears the name
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCsName, uint16(slot), vendorInterface, []byte(name))
	if err != nil {
		return fmt.Errorf("REQ_SET_CS_NAME: %w", err)
	}

	return nil
}

// GetCsName reads the live 32-byte name of a slot.
func (d *Device) GetCsName(slot int) (string, error) {
	if d.closed {
		return "", fmt.Errorf("device is closed")
	}

	buf := make([]byte, CsNameLen)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCsName, uint16(slot), vendorInterface, buf)
	if err != nil {
		return "", fmt.Errorf("REQ_GET_CS_NAME: %w", err)
	}

	end := bytes.IndexByte(buf[:n], 0)
	if end == -1 {
		end = n
	}
	return string(bytes.TrimSpace(buf[:end])), nil
}

// SetCsIrCommand uploads an IR remote command to a sub-slot (0..15;
// REQ_SET_CS_IR_CMD).  Apply-live-only preview; the result lands in
// GetCsStatus (reported as 0x80 | sub-slot).
func (d *Device) SetCsIrCommand(subslot int, c *IrCommand) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	if subslot < 0 || subslot >= CsMaxIRCommands {
		return fmt.Errorf("sub-slot %d out of range (0-%d)", subslot, CsMaxIRCommands-1)
	}

	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqSetCsIrCommand, uint16(subslot), vendorInterface, c.Encode())
	if err != nil {
		return fmt.Errorf("REQ_SET_CS_IR_CMD: %w", err)
	}

	return nil
}

// GetCsIrCommand reads the live IR command of a sub-slot.
func (d *Device) GetCsIrCommand(subslot int) (*IrCommand, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 16)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqGetCsIrCommand, uint16(subslot), vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_GET_CS_IR_CMD: %w", err)
	}

	return DecodeIrCommand(buf[:n])
}

// CsIrLearnArm arms the IR receiver to capture the next press.  Returns an
// error (with CsStatusNoIR) when no live IR component is bound.
func (d *Device) CsIrLearnArm() error {
	return d.csIrLearnControl(1)
}

// CsIrLearnCancel cancels an armed learn.
func (d *Device) CsIrLearnCancel() error {
	return d.csIrLearnControl(0)
}

func (d *Device) csIrLearnControl(action uint16) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqCsIrLearn, action, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_CS_IR_LEARN: %w", err)
	}

	return nil
}

// CsIrLearnRead reads the last learn result (REQ_CS_IR_LEARN wValue 2).
func (d *Device) CsIrLearnRead() (*CsIRLearnResult, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}

	buf := make([]byte, 8)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqCsIrLearn, 2, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_CS_IR_LEARN: %w", err)
	}

	return DecodeCsIRLearnResult(buf[:n])
}

// CsSave persists the whole live CS config (bindings + IR commands + slot
// names) to flash.  Deferred; the result lands in GetCsStatus.
func (d *Device) CsSave() error {
	return d.csSaveRevert(ReqCsSave, "REQ_CS_SAVE")
}

// CsRevert discards the live preview and re-applies the stored CS config.
// Deferred; the result lands in GetCsStatus.
func (d *Device) CsRevert() error {
	return d.csSaveRevert(ReqCsRevert, "REQ_CS_REVERT")
}

func (d *Device) csSaveRevert(bRequest uint8, name string) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}

	buf := make([]byte, 1)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, bRequest, 0, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}

	return nil
}
