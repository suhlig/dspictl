package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/suhlig/dspi"
)

const scanInterval = 100

var (
	colBG      = lipgloss.Color("#1a1b26")
	colFG      = lipgloss.Color("#a9b1d6")
	colMuted   = lipgloss.Color("#565f89")
	colGreen   = lipgloss.Color("#9ece6a")
	colYellow  = lipgloss.Color("#e0af68")
	colRed     = lipgloss.Color("#f7768e")
	colBlue    = lipgloss.Color("#7aa2f7")
	colCyan    = lipgloss.Color("#73daca")
	colPurple  = lipgloss.Color("#bb9af7")
	colClipBg  = lipgloss.Color("#340000")
	colBarBg   = lipgloss.Color("#292e42")
	colBorder  = lipgloss.Color("#3b4261")
	colTitle   = lipgloss.Color("#c0caf5")
	colCPUOK   = lipgloss.Color("#9ece6a")
	colCPUWarn = lipgloss.Color("#e0af68")
	colCPUCrit = lipgloss.Color("#f7768e")
)

var (
	styleTitle = lipgloss.NewStyle().
			Foreground(colTitle).
			Bold(true).
			Align(lipgloss.Center)

	styleSubtitle = lipgloss.NewStyle().
			Foreground(colMuted).
			Align(lipgloss.Center)

	styleClipLabel = lipgloss.NewStyle().
			Foreground(colRed).
			Bold(true).
			PaddingLeft(2)

	styleClipChannel = lipgloss.NewStyle().
				Background(colClipBg).
				Foreground(colRed).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)

	styleFooter = lipgloss.NewStyle().
			Foreground(colMuted).
			Align(lipgloss.Center)

	styleTabActive = lipgloss.NewStyle().
			Foreground(colTitle).
			Bold(true).
			Padding(0, 2)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colMuted).
				Padding(0, 2)

	styleTabClip = lipgloss.NewStyle().
			Foreground(colRed).
			Bold(true).
			Padding(0, 1)
)

var channelColors = []lipgloss.Color{
	colBlue,
	colRed,
	colCyan,
	colGreen,
	colCyan,
	colGreen,
	colCyan,
	colGreen,
	colCyan,
	colGreen,
	colPurple,
}

type model struct {
	devices       []*dspi.Device
	snaps         []dspi.MeterSnapshot
	activeDevice  int
	ticksSinceScan int
	clippedCh     [][]int
	clipTimer     []int
	err           error
	width         int
	connected     bool
}

type tickMsg time.Time
type errMsg struct{ error }
type devicesMsg []*dspi.Device

func initialModel() model {
	return model{width: 80}
}

func tick() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		devs, err := dspi.OpenAll()

		if err != nil {
			return errMsg{fmt.Errorf("connect: %w", err)}
		}

		return devicesMsg(devs)
	}
}

func rescanCmd() tea.Cmd {
	return func() tea.Msg {
		devs, err := dspi.OpenAll()

		if err != nil {
			return nil
		}

		return devicesMsg(devs)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(connectCmd(), tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			for _, dev := range m.devices {
				dev.Close()
			}

			return m, tea.Quit

		case "tab", "right":
			if len(m.devices) > 1 {
				m.activeDevice = (m.activeDevice + 1) % len(m.devices)
			}

			return m, nil

		case "shift+tab", "left":
			if len(m.devices) > 1 {
				m.activeDevice = (m.activeDevice - 1 + len(m.devices)) % len(m.devices)
			}

			return m, nil

		case "c":
			for _, dev := range m.devices {
				_ = dev.ClearClips()
			}

			m.clippedCh = make([][]int, len(m.devices))
			m.clipTimer = make([]int, len(m.devices))

			return m, nil

		case "r":
			for _, dev := range m.devices {
				dev.Close()
			}

			m.devices = nil
			m.snaps = nil
			m.clippedCh = nil
			m.clipTimer = nil
			m.connected = false
			m.err = nil

			return m, connectCmd()
		}

	case errMsg:
		m.err = msg.error
		m.connected = false

		return m, nil

	case devicesMsg:
		for _, dev := range m.devices {
			dev.Close()
		}

		m.devices = msg
		m.snaps = make([]dspi.MeterSnapshot, len(msg))
		m.clippedCh = make([][]int, len(msg))
		m.clipTimer = make([]int, len(msg))

		if m.activeDevice >= len(m.devices) && len(m.devices) > 0 {
			m.activeDevice = len(m.devices) - 1
		}

		m.connected = true
		m.err = nil

		return m, nil

	case tickMsg:
		if !m.connected || len(m.devices) == 0 {
			return m, tick()
		}

		for i, dev := range m.devices {
			snap := dev.ReadMeter()

			m.snaps[i] = snap

			if snap.Err() != nil {
				continue
			}

			newClips := snap.ClipFlags
			var clipped []int

			for j := 0; j < snap.Channels; j++ {
				if newClips&(1<<j) != 0 {
					clipped = append(clipped, j)
				}
			}

			if len(clipped) > 0 {
				m.clippedCh[i] = append(m.clippedCh[i], clipped...)
				m.clipTimer[i] = 170
			}

			if m.clipTimer[i] > 0 {
				m.clipTimer[i]--

				if m.clipTimer[i] == 0 {
					m.clippedCh[i] = nil
					_ = dev.ClearClips()
				}
			}
		}

		var cmds []tea.Cmd

		m.ticksSinceScan++

		if m.ticksSinceScan >= scanInterval {
			m.ticksSinceScan = 0
			cmds = append(cmds, rescanCmd())
		}

		cmds = append(cmds, tick())

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return renderError(m.err, m.width)
	}

	if !m.connected || len(m.devices) == 0 {
		return renderConnecting(m.width)
	}

	dev := m.devices[m.activeDevice]
	snap := m.snaps[m.activeDevice]
	channels := dspi.ChannelTable(dev.Platform())

	var b strings.Builder

	b.WriteString(styleTitle.Width(m.width).Render("DSPi Live Meter"))
	b.WriteString("\n")

	b.WriteString(m.renderTabs())

	cpu0col := cpuColor(snap.CPU0)
	cpu1col := cpuColor(snap.CPU1)
	subStr := fmt.Sprintf("Device %d/%d  |  %s  |  Serial: %s  |  CPU0: %d%%  CPU1: %d%%",
		m.activeDevice+1, len(m.devices), dev.Platform(), dev.Serial(), snap.CPU0, snap.CPU1)
	b.WriteString(styleSubtitle.Width(m.width).Render(subStr))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(colBorder).
		Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	cpuBar := drawBar(float64(snap.CPU0)/100.0, 30, cpu0col)
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		fmt.Sprintf("Core 0 %s %3d%%", cpuBar, snap.CPU0),
	))
	b.WriteString("\n")
	cpuBar = drawBar(float64(snap.CPU1)/100.0, 30, cpu1col)
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		fmt.Sprintf("Core 1 %s %3d%%", cpuBar, snap.CPU1),
	))
	b.WriteString("\n\n")

	type groupEntry struct {
		info dspi.ChannelInfo
		peak float64
	}

	groups := map[string][]groupEntry{}

	for _, ch := range channels {
		if ch.Index >= snap.Channels {
			continue
		}
		groups[ch.Group] = append(groups[ch.Group], groupEntry{ch, snap.Peaks[ch.Index]})
	}

	groupOrder := []string{"USB Input", "S/PDIF Output", "PDM Sub"}
	for _, g := range groupOrder {
		entries, ok := groups[g]

		if !ok || len(entries) == 0 {
			continue
		}

		b.WriteString(lipgloss.NewStyle().
			Foreground(colMuted).
			Bold(true).
			PaddingLeft(2).
			Render(g))
		b.WriteString("\n")

		for _, e := range entries {
			ch := e.info
			idx := ch.Index
			peak := e.peak
			col := channelColors[idx%len(channelColors)]

			isClipped := false
			for _, c := range m.clippedCh[m.activeDevice] {
				if c == idx {
					isClipped = true

					break
				}
			}

			clipMark := ""
			if isClipped {
				clipMark = styleClipChannel.Render("CLIP")
			}

			nameStyle := lipgloss.NewStyle().
				Width(14).
				Foreground(col).
				Bold(true).
				PaddingLeft(2)
			dbfsStr := dspi.DBFS(peak)
			valStyle := lipgloss.NewStyle().
				Width(8).
				Align(lipgloss.Right).
				Foreground(colMuted)

			bar := drawBar(peak, 30, col)
			line := fmt.Sprintf("%s%s %s %s",
				nameStyle.Render(ch.Name),
				bar,
				valStyle.Render(dbfsStr+" dB"),
				clipMark,
			)
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(m.clippedCh[m.activeDevice]) > 0 {
		var clipNames []string

		for _, idx := range m.clippedCh[m.activeDevice] {
			if idx < len(channels) {
				clipNames = append(clipNames, channels[idx].Name)
			}
		}
		b.WriteString(styleClipLabel.Render("CLIP:"))
		b.WriteString(" ")

		for _, name := range clipNames {
			b.WriteString(styleClipChannel.Render(name))
			b.WriteString(" ")
		}
		b.WriteString("\n\n")
	}

	b.WriteString(lipgloss.NewStyle().
		Foreground(colBorder).
		Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	b.WriteString(styleFooter.Width(m.width).Render(
		"q: quit  |  Tab/← →: switch device  |  c: clear clips  |  r: rescan",
	))

	return lipgloss.NewStyle().
		Background(colBG).
		Padding(1, 2).
		Render(b.String())
}

func (m model) renderTabs() string {
	if len(m.devices) <= 1 {
		return ""
	}

	var tabs []string

	for i, dev := range m.devices {
		serial := dev.Serial()
		label := fmt.Sprintf("%d: %s", i+1, serial)

		hasClip := len(m.clippedCh[i]) > 0

		if i == m.activeDevice && hasClip {
			tabs = append(tabs, styleTabClip.Render(label))
		} else if i == m.activeDevice {
			tabs = append(tabs, styleTabActive.Render(label))
		} else if hasClip {
			tabs = append(tabs, styleTabClip.Render(label))
		} else {
			tabs = append(tabs, styleTabInactive.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}

func drawBar(fraction float64, width int, color lipgloss.Color) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	filled := int(fraction * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	fillStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", empty)

	return lipgloss.NewStyle().
		Foreground(color).
		Background(colBarBg).
		Render(fillStr + emptyStr)
}

func cpuColor(load int) lipgloss.Color {
	switch {
	case load >= 90:
		return colCPUCrit
	case load >= 60:
		return colCPUWarn
	default:
		return colCPUOK
	}
}

func renderConnecting(width int) string {
	return lipgloss.NewStyle().
		Background(colBG).
		Padding(2, 2).
		Width(width).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styleTitle.Width(width).Render("DSPi Live Meter"),
				"",
				lipgloss.NewStyle().Foreground(colYellow).Render("Connecting to DSPi..."),
				"",
				lipgloss.NewStyle().Foreground(colMuted).Render("Make sure the device(s) are plugged in via USB"),
			),
		)
}

func renderError(err error, width int) string {
	errStr := err.Error()

	return lipgloss.NewStyle().
		Background(colBG).
		Padding(2, 2).
		Width(width).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styleTitle.Width(width).Render("DSPi Live Meter"),
				"",
				lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("Connection Error"),
				"",
				lipgloss.NewStyle().Foreground(colFG).Render(errStr),
				"",
				lipgloss.NewStyle().Foreground(colMuted).Render("Press 'r' to retry, 'q' to quit"),
			),
		)
}

func main() {
	err := mainE()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

func mainE() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	_, err := p.Run()

	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	return nil
}
