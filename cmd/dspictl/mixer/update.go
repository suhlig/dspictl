package mixer

import (
	tea "charm.land/bubbletea/v2"
	"github.com/suhlig/dspi"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(connectCmd(m.targetSerial), tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case errMsg:
		m.err = msg.error
		m.connected = false

		return m, nil

	case devicesMsg:
		return m.handleDevices(msg)

	case tickMsg:
		return m.handleTick()
	}

	return m, nil
}

func (m model) handleWindowSize(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.dm.CloseAll()

		return m, tea.Quit

	case "tab":
		if m.dm.Len() > 1 {
			m.activeDevice = (m.activeDevice + 1) % m.dm.Len()
		}

		return m, nil

	case "shift+tab":
		if m.dm.Len() > 1 {
			m.activeDevice = (m.activeDevice - 1 + m.dm.Len()) % m.dm.Len()
		}

		return m, nil

	case "up":
		if m.dm.Len() > 0 && m.activeDevice < m.dm.Len() {
			db := m.dm.MasterVolume(m.activeDevice).DB()
			m.dm.SetMasterVolume(m.activeDevice, dspi.NewGain(min(db+1, 0)))
			_ = m.dm.Device(m.activeDevice).SetMasterVolume(m.dm.MasterVolume(m.activeDevice))
		}

		return m, nil

	case "down":
		if m.dm.Len() > 0 && m.activeDevice < m.dm.Len() {
			db := m.dm.MasterVolume(m.activeDevice).DB()
			m.dm.SetMasterVolume(m.activeDevice, dspi.NewGain(max(db-1, -128)))
			_ = m.dm.Device(m.activeDevice).SetMasterVolume(m.dm.MasterVolume(m.activeDevice))
		}

		return m, nil

	case "c":
		m.dm.ClearAllClips()

		return m, nil

	case "r":
		m.dm.CloseAll()
		m.dm = newDeviceManager()
		m.activeDevice = 0
		m.connected = false
		m.err = nil

		return m, connectCmd(m.targetSerial)

	case "m":
		if m.dm.Len() > 0 && m.activeDevice < m.dm.Len() {
			m.dm.ToggleMute(m.activeDevice)
		}

		return m, nil

	case "M":
		m.dm.ToggleMuteAll()

		return m, nil
	}

	return m, nil
}

func (m model) handleDevices(msg devicesMsg) (model, tea.Cmd) {
	if m.dm.Len() > 0 {
		m.dm.Resync(msg)
	} else {
		m.dm.Initialize(msg)
	}

	if m.activeDevice >= m.dm.Len() && m.dm.Len() > 0 {
		m.activeDevice = m.dm.Len() - 1
	}

	m.connected = true
	m.err = nil

	return m, nil
}

func (m model) handleTick() (model, tea.Cmd) {
	if !m.connected || m.dm.Len() == 0 {
		return m, tick()
	}

	for i := range m.dm.Len() {
		m.dm.ReadMeter(i)
		m.dm.ProcessClips(i)
		m.dm.TickClipTimer(i)

		if i == m.activeDevice {
			m.dm.RefreshMasterVolume(i)
		}
	}

	var cmds []tea.Cmd

	m.ticksSinceScan++

	if m.ticksSinceScan >= scanInterval {
		m.ticksSinceScan = 0
		cmds = append(cmds, rescanCmd(m.targetSerial))
	}

	cmds = append(cmds, tick())

	return m, tea.Batch(cmds...)
}
