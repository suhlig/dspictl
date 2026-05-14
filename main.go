package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── colour palette ────────────────────────────────────────────────────

var (
	colBG       = lipgloss.Color("#1a1b26")
	colFG       = lipgloss.Color("#a9b1d6")
	colMuted    = lipgloss.Color("#565f89")
	colGreen    = lipgloss.Color("#9ece6a")
	colYellow   = lipgloss.Color("#e0af68")
	colRed      = lipgloss.Color("#f7768e")
	colBlue     = lipgloss.Color("#7aa2f7")
	colCyan     = lipgloss.Color("#73daca")
	colPurple   = lipgloss.Color("#bb9af7")
	colClipBg   = lipgloss.Color("#340000")
	colBarBg    = lipgloss.Color("#292e42")
	colBorder   = lipgloss.Color("#3b4261")
	colTitle    = lipgloss.Color("#c0caf5")
	colCPUOK    = lipgloss.Color("#9ece6a")
	colCPUWarn  = lipgloss.Color("#e0af68")
	colCPUCrit  = lipgloss.Color("#f7768e")
)

var (
	styleTitle = lipgloss.NewStyle().
			Foreground(colTitle).
			Bold(true).
			Align(lipgloss.Center).
			Width(80)

	styleSubtitle = lipgloss.NewStyle().
			Foreground(colMuted).
			Align(lipgloss.Center).
			Width(80)

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
			Align(lipgloss.Center).
			Width(80)
)

var channelColors = []lipgloss.Color{
	colBlue,   // USB L
	colRed,    // USB R
	colCyan,   // SPDIF 1 L
	colGreen,  // SPDIF 1 R
	colCyan,   // SPDIF 2 L
	colGreen,  // SPDIF 2 R
	colCyan,   // SPDIF 3 L
	colGreen,  // SPDIF 3 R
	colCyan,   // SPDIF 4 L
	colGreen,  // SPDIF 4 R
	colPurple, // PDM Sub
}

// ─── TUI model ─────────────────────────────────────────────────────────

type model struct {
	device      *DSPiDevice
	lastSnap    MeterSnapshot
	clippedCh   []int   // channels that have clipped (held for display)
	clipTimer   int     // counts down ticks before clearing clip display
	err         error
	width       int
	connected   bool
}

type tickMsg time.Time
type errMsg struct{ error }

func initialModel() model {
	return model{width: 80}
}

// ─── commands ──────────────────────────────────────────────────────────

func tick() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func connectCmd() tea.Msg {
	dev, err := OpenDSPi()
	if err != nil {
		return errMsg{fmt.Errorf("connect: %w", err)}
	}
	return dev
}

// ─── init / update / view ──────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(connectCmd, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			// Clear clip indicators
			m.clippedCh = nil
			m.clipTimer = 0
			if m.device != nil {
				m.device.ClearClips()
			}
			return m, nil
		case "r":
			// Reconnect
			if m.device != nil {
				m.device.Close()
			}
			m.connected = false
			m.err = nil
			return m, connectCmd
		}

	case errMsg:
		m.err = msg.error
		m.connected = false
		return m, nil

	case *DSPiDevice:
		m.device = msg
		m.connected = true
		m.err = nil
		return m, nil

	case tickMsg:
		if !m.connected || m.device == nil {
			return m, tick()
		}

		snap := m.device.ReadMeter()
		if snap.Err() != nil {
			m.err = snap.Err()
			m.connected = false
			return m, tick()
		}

		m.lastSnap = snap

		// Track newly clipped channels for display
		newClips := snap.ClipFlags
		var clipped []int
		for i := 0; i < snap.Channels; i++ {
			if newClips&(1<<i) != 0 {
				clipped = append(clipped, i)
			}
		}
		if len(clipped) > 0 {
			m.clippedCh = append(m.clippedCh, clipped...)
			m.clipTimer = 170 // ~10 seconds at 60ms tick
		}
		if m.clipTimer > 0 {
			m.clipTimer--
			if m.clipTimer == 0 {
				m.clippedCh = nil
				m.device.ClearClips()
			}
		}

		return m, tick()
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return renderError(m.err)
	}

	if !m.connected {
		return renderConnecting()
	}

	var b strings.Builder
	channels := channelTable(m.device.Platform())
	snap := m.lastSnap

	// Title
	b.WriteString(styleTitle.Render("DSPi Live Meter"))
	b.WriteString("\n")

	// Platform & CPU
	cpu0col := cpuColor(snap.CPU0)
	cpu1col := cpuColor(snap.CPU1)
	platStr := fmt.Sprintf("Platform: %s  |  CPU0: %d%%  CPU1: %d%%",
		m.device.Platform(), snap.CPU0, snap.CPU1)
	b.WriteString(styleSubtitle.Render(
		lipgloss.NewStyle().Foreground(colMuted).Render(platStr),
	))
	b.WriteString("\n\n")

	// Separator line
	b.WriteString(lipgloss.NewStyle().
		Foreground(colBorder).
		Render(strings.Repeat("─", min(m.width, 80))))
	b.WriteString("\n")

	// Draw CPU bars
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

	// Group channels by category
	type groupEntry struct {
		info ChannelInfo
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

		// Group header
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

			// Determine if this channel is clipped
			isClipped := false
			for _, c := range m.clippedCh {
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
			dbfsStr := DBFS(peak)
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

	// Clip summary line
	if len(m.clippedCh) > 0 {
		var clipNames []string
		for _, idx := range m.clippedCh {
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

	// Footer
	b.WriteString(lipgloss.NewStyle().
		Foreground(colBorder).
		Render(strings.Repeat("─", min(m.width, 80))))
	b.WriteString("\n")
	b.WriteString(styleFooter.Render(
		"q: quit  |  c: clear clips  |  r: reconnect",
	))

	return lipgloss.NewStyle().
		Background(colBG).
		Padding(1, 2).
		Render(b.String())
}

// ─── helpers ───────────────────────────────────────────────────────────

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

	fillChar := "█"
	emptyChar := "░"

	fillStr := strings.Repeat(fillChar, filled)
	emptyStr := strings.Repeat(emptyChar, empty)

	return lipgloss.NewStyle().
		Foreground(color).
		Background(colBarBg).
		Render(fillStr+emptyStr)
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

func renderConnecting() string {
	return lipgloss.NewStyle().
		Background(colBG).
		Padding(2, 2).
		Width(60).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styleTitle.Render("DSPi Live Meter"),
				"",
				lipgloss.NewStyle().Foreground(colYellow).Render("🔌 Connecting to DSPi..."),
				"",
				lipgloss.NewStyle().Foreground(colMuted).Render("Make sure the device is plugged in via USB"),
			),
		)
}

func renderError(err error) string {
	errStr := err.Error()
	return lipgloss.NewStyle().
		Background(colBG).
		Padding(2, 2).
		Width(60).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styleTitle.Render("DSPi Live Meter"),
				"",
				lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("⚠ Connection Error"),
				"",
				lipgloss.NewStyle().Foreground(colFG).Render(errStr),
				"",
				lipgloss.NewStyle().Foreground(colMuted).Render("Press 'r' to retry, 'q' to quit"),
			),
		)
}

// ─── entry point ───────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
