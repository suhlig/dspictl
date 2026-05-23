package mixer

import (
	"time"

	"github.com/suhlig/dspi"
)

type tickMsg time.Time

type errMsg struct{ error }

type devicesMsg []*dspi.Device

type model struct {
	dm             *deviceManager
	activeDevice   int
	ticksSinceScan int
	err            error
	width          int
	height         int
	connected      bool
	targetSerial   string
}

func newModel(targetSerial string) model {
	return model{
		dm:           newDeviceManager(),
		width:        80,
		height:       24,
		targetSerial: targetSerial,
	}
}

type deviceManager struct {
	devices       []*dspi.Device
	snaps         []dspi.MeterSnapshot
	clippedCh     [][]int
	clipTimer     []int
	masterVolume  []dspi.Gain
	preMuteVolume []dspi.Gain
	channels      [][]dspi.ChannelInfo
}

func newDeviceManager() *deviceManager {
	return &deviceManager{}
}

func (dm *deviceManager) Initialize(devices []*dspi.Device) {
	dm.devices = devices
	n := len(devices)
	dm.snaps = make([]dspi.MeterSnapshot, n)
	dm.masterVolume = make([]dspi.Gain, n)
	dm.preMuteVolume = make([]dspi.Gain, n)
	dm.clippedCh = make([][]int, n)
	dm.clipTimer = make([]int, n)
	dm.channels = make([][]dspi.ChannelInfo, n)

	for i := range dm.masterVolume {
		dm.masterVolume[i] = dspi.NewGain(-20)
		dm.preMuteVolume[i] = dspi.NewGain(-20)
	}

	for i, dev := range devices {
		if chs, err := dev.Channels(); err == nil {
			dm.channels[i] = chs
		}
	}
}

func (dm *deviceManager) Resync(devices []*dspi.Device) {
	newSerials := make(map[string]struct{}, len(devices))

	for _, dev := range devices {
		newSerials[dev.Serial()] = struct{}{}
	}

	for i := len(dm.devices) - 1; i >= 0; i-- {
		if _, exists := newSerials[dm.devices[i].Serial()]; !exists {
			dm.devices[i].Close()
			dm.devices = append(dm.devices[:i], dm.devices[i+1:]...)
			dm.snaps = append(dm.snaps[:i], dm.snaps[i+1:]...)
			dm.clippedCh = append(dm.clippedCh[:i], dm.clippedCh[i+1:]...)
			dm.clipTimer = append(dm.clipTimer[:i], dm.clipTimer[i+1:]...)
			dm.masterVolume = append(dm.masterVolume[:i], dm.masterVolume[i+1:]...)
			dm.preMuteVolume = append(dm.preMuteVolume[:i], dm.preMuteVolume[i+1:]...)
			dm.channels = append(dm.channels[:i], dm.channels[i+1:]...)
		}
	}

	remainingSerials := make(map[string]struct{}, len(dm.devices))

	for _, dev := range dm.devices {
		remainingSerials[dev.Serial()] = struct{}{}
	}

	for _, dev := range devices {
		if _, exists := remainingSerials[dev.Serial()]; !exists {
			dm.devices = append(dm.devices, dev)
			dm.snaps = append(dm.snaps, dspi.MeterSnapshot{})
			dm.masterVolume = append(dm.masterVolume, dspi.NewGain(-20))
			dm.preMuteVolume = append(dm.preMuteVolume, dspi.NewGain(-20))
			dm.clippedCh = append(dm.clippedCh, nil)
			dm.clipTimer = append(dm.clipTimer, 0)

			if chs, err := dev.Channels(); err == nil {
				dm.channels = append(dm.channels, chs)
			} else {
				dm.channels = append(dm.channels, nil)
			}
		} else {
			dev.Close()
		}
	}
}

func (dm *deviceManager) Len() int {
	return len(dm.devices)
}

func (dm *deviceManager) Device(i int) *dspi.Device {
	return dm.devices[i]
}

func (dm *deviceManager) Snap(i int) dspi.MeterSnapshot {
	return dm.snaps[i]
}

func (dm *deviceManager) MasterVolume(i int) dspi.Gain {
	return dm.masterVolume[i]
}

func (dm *deviceManager) SetMasterVolume(i int, g dspi.Gain) {
	dm.masterVolume[i] = g
}

func (dm *deviceManager) HasClip(i int) bool {
	return len(dm.clippedCh[i]) > 0
}

func (dm *deviceManager) ClippedCh(i int) []int {
	return dm.clippedCh[i]
}

func (dm *deviceManager) AllDevices() []*dspi.Device {
	return dm.devices
}

func (dm *deviceManager) AllClippedCh() [][]int {
	return dm.clippedCh
}

func (dm *deviceManager) Channels(i int) []dspi.ChannelInfo {
	return dm.channels[i]
}

func (dm *deviceManager) ClearAllClips() {
	for _, dev := range dm.devices {
		_ = dev.ClearClips()
	}

	dm.clippedCh = make([][]int, len(dm.devices))
	dm.clipTimer = make([]int, len(dm.devices))
}

func (dm *deviceManager) CloseAll() {
	for _, dev := range dm.devices {
		dev.Close()
	}
}

func (dm *deviceManager) ReadMeter(i int) {
	snap := dm.devices[i].ReadMeter()
	dm.snaps[i] = snap
}

func (dm *deviceManager) ProcessClips(i int) {
	snap := dm.snaps[i]

	if snap.Err() != nil {
		return
	}

	var clipped []int
	newClips := snap.ClipFlags

	for j := 0; j < snap.Channels; j++ {
		if newClips&(1<<j) != 0 {
			clipped = append(clipped, j)
		}
	}

	if len(clipped) > 0 {
		dm.clippedCh[i] = append(dm.clippedCh[i], clipped...)
		dm.clipTimer[i] = clipTimerDuration
	}
}

func (dm *deviceManager) TickClipTimer(i int) {
	if dm.clipTimer[i] <= 0 {
		return
	}

	dm.clipTimer[i]--

	if dm.clipTimer[i] == 0 {
		dm.clippedCh[i] = nil
		_ = dm.devices[i].ClearClips()
	}
}

func (dm *deviceManager) RefreshMasterVolume(i int) {
	if mv, err := dm.devices[i].GetMasterVolume(); err == nil {
		dm.masterVolume[i] = mv
	}
}

const muteThresholdDB = -120

func (dm *deviceManager) IsMuted(i int) bool {
	return dm.masterVolume[i].DB() <= muteThresholdDB
}

func (dm *deviceManager) ToggleMute(i int) {
	if dm.IsMuted(i) {
		dm.masterVolume[i] = dm.preMuteVolume[i]
	} else {
		dm.preMuteVolume[i] = dm.masterVolume[i]
		dm.masterVolume[i] = dspi.NewGain(-128)
	}
	_ = dm.devices[i].SetMasterVolume(dm.masterVolume[i])
}

func (dm *deviceManager) ToggleMuteAll() {
	if dm.Len() == 0 {
		return
	}
	allMuted := true
	for i := range dm.Len() {
		if !dm.IsMuted(i) {
			allMuted = false
			break
		}
	}
	if allMuted {
		for i := range dm.Len() {
			dm.masterVolume[i] = dm.preMuteVolume[i]
			_ = dm.devices[i].SetMasterVolume(dm.masterVolume[i])
		}
	} else {
		for i := range dm.Len() {
			if !dm.IsMuted(i) {
				dm.preMuteVolume[i] = dm.masterVolume[i]
			}
			dm.masterVolume[i] = dspi.NewGain(-128)
			_ = dm.devices[i].SetMasterVolume(dm.masterVolume[i])
		}
	}
}
