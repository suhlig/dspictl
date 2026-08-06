package dspi

import (
	"encoding/binary"
	"fmt"
	"math"
)

// UpmixConfigPacket is the 44-byte wire image of the stereo upmixer config,
// byte-identical to the firmware's UpmixConfigPacket / WireUpmixParams (V25+,
// presence byte V26+).  All floats are little-endian IEEE 754.
type UpmixConfigPacket struct {
	Enabled          bool
	CenterMode       uint8 // UpmixCenterMode*
	SurroundMode     uint8 // UpmixSurroundMode*
	PresenceQ1       int8  // centre presence bell, dB * 2 (V26+)
	StrengthPct      float32
	CenterWidthPct   float32
	CorrThresholdPct float32
	AttackMs         float32
	ReleaseMs        float32
	DetectorHpfHz    float32
	SurroundDelayMs  float32
	SurroundHpfHz    float32
	SurroundLpfHz    float32
	DecorrPct        float32
}

// Encode serialises the config into the 44-byte wire layout.
func (c *UpmixConfigPacket) Encode() []byte {
	buf := make([]byte, 44)
	if c.Enabled {
		buf[0] = 1
	}
	buf[1] = c.CenterMode
	buf[2] = c.SurroundMode
	buf[3] = uint8(c.PresenceQ1)
	putFloat32(buf[4:8], c.StrengthPct)
	putFloat32(buf[8:12], c.CenterWidthPct)
	putFloat32(buf[12:16], c.CorrThresholdPct)
	putFloat32(buf[16:20], c.AttackMs)
	putFloat32(buf[20:24], c.ReleaseMs)
	putFloat32(buf[24:28], c.DetectorHpfHz)
	putFloat32(buf[28:32], c.SurroundDelayMs)
	putFloat32(buf[32:36], c.SurroundHpfHz)
	putFloat32(buf[36:40], c.SurroundLpfHz)
	putFloat32(buf[40:44], c.DecorrPct)
	return buf
}

// DecodeUpmixConfig parses a 44-byte wire payload.
func DecodeUpmixConfig(raw []byte) (*UpmixConfigPacket, error) {
	if len(raw) < 44 {
		return nil, fmt.Errorf("upmix config payload too short (got %d, need 44)", len(raw))
	}
	return &UpmixConfigPacket{
		Enabled:          raw[0] != 0,
		CenterMode:       raw[1],
		SurroundMode:     raw[2],
		PresenceQ1:       int8(raw[3]),
		StrengthPct:      getFloat32(raw[4:8]),
		CenterWidthPct:   getFloat32(raw[8:12]),
		CorrThresholdPct: getFloat32(raw[12:16]),
		AttackMs:         getFloat32(raw[16:20]),
		ReleaseMs:        getFloat32(raw[20:24]),
		DetectorHpfHz:    getFloat32(raw[24:28]),
		SurroundDelayMs:  getFloat32(raw[28:32]),
		SurroundHpfHz:    getFloat32(raw[32:36]),
		SurroundLpfHz:    getFloat32(raw[36:40]),
		DecorrPct:        getFloat32(raw[40:44]),
	}, nil
}

// PresenceDB decodes the wire presence byte to dB (0.5 dB steps).
func (c *UpmixConfigPacket) PresenceDB() float32 {
	return 0.5 * float32(c.PresenceQ1)
}

// SetPresenceDB encodes a dB value into the wire presence byte, clamped to
// the firmware's [-12, +12] dB range.
func (c *UpmixConfigPacket) SetPresenceDB(db float32) {
	if db < -12 {
		db = -12
	}
	if db > 12 {
		db = 12
	}
	c.PresenceQ1 = int8(math.Round(float64(db) * 2))
}

// UpmixStatus is the 16-byte live telemetry from REQ_UPMIX_GET_STATUS.
type UpmixStatus struct {
	Active        bool   // 1 = processing this packet stream
	ParkedReason  uint8  // UpmixParked*
	CorrQ14       int16  // smoothed L/R correlation, Q14
	BalanceQ14    int16  // smoothed |L/R| balance, Q14
	CenterGainQ15 uint16 // smoothed centre extraction gain, Q15
	LsGainQ15     uint16 // surround steering gains, Q15
	RsGainQ15     uint16
}

// DecodeUpmixStatus parses a 16-byte status payload.
func DecodeUpmixStatus(raw []byte) (*UpmixStatus, error) {
	if len(raw) < 16 {
		return nil, fmt.Errorf("upmix status payload too short (got %d, need 16)", len(raw))
	}
	return &UpmixStatus{
		Active:        raw[0] != 0,
		ParkedReason:  raw[1],
		CorrQ14:       int16(binary.LittleEndian.Uint16(raw[2:4])),
		BalanceQ14:    int16(binary.LittleEndian.Uint16(raw[4:6])),
		CenterGainQ15: binary.LittleEndian.Uint16(raw[6:8]),
		LsGainQ15:     binary.LittleEndian.Uint16(raw[8:10]),
		RsGainQ15:     binary.LittleEndian.Uint16(raw[10:12]),
	}, nil
}

// ParkedReasonName returns a human-readable name for an upmix parked reason.
func ParkedReasonName(reason uint8) string {
	switch reason {
	case UpmixParkedActive:
		return "active"
	case UpmixParkedDisabled:
		return "disabled"
	case UpmixParkedNotStereo:
		return "input not stereo"
	case UpmixParkedRateTooHigh:
		return "sample rate above 48 kHz"
	default:
		return fmt.Sprintf("unknown(%d)", reason)
	}
}

// UpmixCenterModeName returns a human-readable name for a centre mode.
func UpmixCenterModeName(mode uint8) string {
	switch mode {
	case UpmixCenterModePassive:
		return "passive"
	case UpmixCenterModeAdaptive:
		return "adaptive"
	case UpmixCenterModeOff:
		return "off"
	default:
		return fmt.Sprintf("unknown(%d)", mode)
	}
}

// UpmixSurroundModeName returns a human-readable name for a surround mode.
func UpmixSurroundModeName(mode uint8) string {
	switch mode {
	case UpmixSurroundModeOff:
		return "off"
	case UpmixSurroundModePassive:
		return "passive"
	case UpmixSurroundModeAdaptive:
		return "adaptive"
	default:
		return fmt.Sprintf("unknown(%d)", mode)
	}
}

// UpmixParamName returns the canonical name of an upmixer parameter id.
func UpmixParamName(param uint16) string {
	switch param {
	case UpmixParamEnabled:
		return "enabled"
	case UpmixParamCenterMode:
		return "center-mode"
	case UpmixParamSurroundMode:
		return "surround-mode"
	case UpmixParamStrength:
		return "strength"
	case UpmixParamCenterWidth:
		return "center-width"
	case UpmixParamThreshold:
		return "threshold"
	case UpmixParamAttack:
		return "attack"
	case UpmixParamRelease:
		return "release"
	case UpmixParamDetHpf:
		return "det-hpf"
	case UpmixParamSurDelay:
		return "surround-delay"
	case UpmixParamSurHpf:
		return "surround-hpf"
	case UpmixParamSurLpf:
		return "surround-lpf"
	case UpmixParamDecorr:
		return "decorr"
	case UpmixParamPresence:
		return "presence"
	default:
		return fmt.Sprintf("unknown(%d)", param)
	}
}

// ParseUpmixParam resolves an upmixer parameter id by name or numeric id.
func ParseUpmixParam(s string) (uint16, error) {
	switch s {
	case "enabled":
		return UpmixParamEnabled, nil
	case "center-mode":
		return UpmixParamCenterMode, nil
	case "surround-mode":
		return UpmixParamSurroundMode, nil
	case "strength":
		return UpmixParamStrength, nil
	case "center-width":
		return UpmixParamCenterWidth, nil
	case "threshold":
		return UpmixParamThreshold, nil
	case "attack":
		return UpmixParamAttack, nil
	case "release":
		return UpmixParamRelease, nil
	case "det-hpf":
		return UpmixParamDetHpf, nil
	case "surround-delay":
		return UpmixParamSurDelay, nil
	case "surround-hpf":
		return UpmixParamSurHpf, nil
	case "surround-lpf":
		return UpmixParamSurLpf, nil
	case "decorr":
		return UpmixParamDecorr, nil
	case "presence":
		return UpmixParamPresence, nil
	}
	var id uint16
	if _, err := fmt.Sscanf(s, "%d", &id); err == nil && id < UpmixParamCount {
		return id, nil
	}
	return 0, fmt.Errorf("unknown upmix parameter %q", s)
}

// SetUpmixConfig applies a whole upmixer config atomically (REQ_UPMIX_SET_CONFIG).
// RP2350 only; the firmware STALLs on RP2040.
func (d *Device) SetUpmixConfig(cfg *UpmixConfigPacket) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqUpmixSetConfig, 0, vendorInterface, cfg.Encode())
	if err != nil {
		return fmt.Errorf("REQ_UPMIX_SET_CONFIG: %w", err)
	}
	return nil
}

// GetUpmixConfig reads the live upmixer config (REQ_UPMIX_GET_CONFIG).
func (d *Device) GetUpmixConfig() (*UpmixConfigPacket, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}
	buf := make([]byte, 44)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqUpmixGetConfig, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_UPMIX_GET_CONFIG: %w", err)
	}
	return DecodeUpmixConfig(buf[:n])
}

// SetUpmixParam sets a single upmixer parameter (REQ_UPMIX_SET_PARAM).
// Mode/enable params are rounded by the firmware; the rest are clamped
// inside upmix_compute_coefficients.
func (d *Device) SetUpmixParam(param uint16, value float32) error {
	if d.closed {
		return fmt.Errorf("device is closed")
	}
	if param >= UpmixParamCount {
		return fmt.Errorf("upmix parameter %d out of range (0-%d)", param, UpmixParamCount-1)
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(value))
	_, err := d.usb.ControlTransfer(vendorInterfaceOutRequest, ReqUpmixSetParam, param, vendorInterface, buf)
	if err != nil {
		return fmt.Errorf("REQ_UPMIX_SET_PARAM(%s): %w", UpmixParamName(param), err)
	}
	return nil
}

// GetUpmixParam reads a single upmixer parameter (REQ_UPMIX_GET_PARAM).
func (d *Device) GetUpmixParam(param uint16) (float32, error) {
	if d.closed {
		return 0, fmt.Errorf("device is closed")
	}
	if param >= UpmixParamCount {
		return 0, fmt.Errorf("upmix parameter %d out of range (0-%d)", param, UpmixParamCount-1)
	}
	buf := make([]byte, 4)
	_, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqUpmixGetParam, param, vendorInterface, buf)
	if err != nil {
		return 0, fmt.Errorf("REQ_UPMIX_GET_PARAM(%s): %w", UpmixParamName(param), err)
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf)), nil
}

// GetUpmixStatus reads the upmixer live telemetry (REQ_UPMIX_GET_STATUS).
func (d *Device) GetUpmixStatus() (*UpmixStatus, error) {
	if d.closed {
		return nil, fmt.Errorf("device is closed")
	}
	buf := make([]byte, 16)
	n, err := d.usb.ControlTransfer(vendorInterfaceInRequest, ReqUpmixGetStatus, 0, vendorInterface, buf)
	if err != nil {
		return nil, fmt.Errorf("REQ_UPMIX_GET_STATUS: %w", err)
	}
	return DecodeUpmixStatus(buf[:n])
}

func putFloat32(b []byte, v float32) {
	binary.LittleEndian.PutUint32(b, math.Float32bits(v))
}

func getFloat32(b []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}
