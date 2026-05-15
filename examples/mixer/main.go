package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/suhlig/dspi"
)

const (
	tickInterval      = 60 * time.Millisecond
	scanInterval      = 100 // in ticks (~6s at 60ms/tick)
	clipTimerDuration = 170 // in ticks (~10s at 60ms/tick)
	nameWidth         = 14
	dbfsValueWidth    = 10
)

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
	colVolume  = lipgloss.Color("#7aa2f7")
	colCPUOK   = lipgloss.Color("#9ece6a")
	colCPUWarn = lipgloss.Color("#e0af68")
	colCPUCrit = lipgloss.Color("#f7768e")
)

var (
	styleTitle = lipgloss.NewStyle().
			Foreground(colTitle).
			Bold(true).
			Background(colBG).
			Align(lipgloss.Center)

	styleSubtitle = lipgloss.NewStyle().
			Foreground(colMuted).
			Background(colBG).
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
			Background(colBG).
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

var channelColors = []color.Color{
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
	devices        []*dspi.Device
	snaps          []dspi.MeterSnapshot
	activeDevice   int
	ticksSinceScan int
	clippedCh      [][]int
	clipTimer      []int
	masterVolume   []dspi.Gain
	err            error
	width          int
	height         int
	connected      bool
}

type tickMsg time.Time
type errMsg struct{ error }
type devicesMsg []*dspi.Device

func initialModel() model {
	return model{width: 80, height: 24}
}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
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
			return errMsg{fmt.Errorf("rescan: %w", err)}
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
		if !acceptWidth(msg.Width) {
			return m, nil
		}

		m.width = msg.Width
		m.height = msg.Height

		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			for _, dev := range m.devices {
				dev.Close()
			}

			return m, tea.Quit

		case "tab":
			if len(m.devices) > 1 {
				m.activeDevice = (m.activeDevice + 1) % len(m.devices)
			}

			return m, nil

		case "shift+tab":
			if len(m.devices) > 1 {
				m.activeDevice = (m.activeDevice - 1 + len(m.devices)) % len(m.devices)
			}

			return m, nil

		case "up":
			if len(m.devices) > 0 && m.activeDevice < len(m.masterVolume) {
				db := m.masterVolume[m.activeDevice].DB()
				m.masterVolume[m.activeDevice] = dspi.NewGain(min(db+1, 0))
				_ = m.devices[m.activeDevice].SetMasterVolume(m.masterVolume[m.activeDevice])
			}

			return m, nil

		case "down":
			if len(m.devices) > 0 && m.activeDevice < len(m.masterVolume) {
				db := m.masterVolume[m.activeDevice].DB()
				m.masterVolume[m.activeDevice] = dspi.NewGain(max(db-1, -128))
				_ = m.devices[m.activeDevice].SetMasterVolume(m.masterVolume[m.activeDevice])
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
			m.masterVolume = nil
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
		if len(m.devices) > 0 {
			newSerials := make(map[string]struct{}, len(msg))
			for _, dev := range msg {
				newSerials[dev.Serial()] = struct{}{}
			}

			for i := len(m.devices) - 1; i >= 0; i-- {
				if _, exists := newSerials[m.devices[i].Serial()]; !exists {
					m.devices[i].Close()
					m.devices = append(m.devices[:i], m.devices[i+1:]...)
					m.snaps = append(m.snaps[:i], m.snaps[i+1:]...)
					m.clippedCh = append(m.clippedCh[:i], m.clippedCh[i+1:]...)
					m.clipTimer = append(m.clipTimer[:i], m.clipTimer[i+1:]...)
					m.masterVolume = append(m.masterVolume[:i], m.masterVolume[i+1:]...)
				}
			}

			remainingSerials := make(map[string]struct{}, len(m.devices))
			for _, dev := range m.devices {
				remainingSerials[dev.Serial()] = struct{}{}
			}

			for _, dev := range msg {
				if _, exists := remainingSerials[dev.Serial()]; !exists {
					m.devices = append(m.devices, dev)
					m.snaps = append(m.snaps, dspi.MeterSnapshot{})
					m.masterVolume = append(m.masterVolume, dspi.NewGain(-20))
					m.clippedCh = append(m.clippedCh, nil)
					m.clipTimer = append(m.clipTimer, 0)
				} else {
					dev.Close()
				}
			}

			if m.activeDevice >= len(m.devices) && len(m.devices) > 0 {
				m.activeDevice = len(m.devices) - 1
			}

			m.connected = true
			m.err = nil

			return m, nil
		}

		m.devices = msg
		m.snaps = make([]dspi.MeterSnapshot, len(msg))
		m.masterVolume = make([]dspi.Gain, len(msg))

		for i := range m.masterVolume {
			m.masterVolume[i] = dspi.NewGain(-20)
		}

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
				m.clipTimer[i] = clipTimerDuration
			}

			if m.clipTimer[i] > 0 {
				m.clipTimer[i]--

				if m.clipTimer[i] == 0 {
					m.clippedCh[i] = nil
					_ = dev.ClearClips()
				}
			}

			if i == m.activeDevice && i < len(m.masterVolume) {
				if mv, err := dev.GetMasterVolume(); err == nil {
					m.masterVolume[i] = mv
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

func acceptWidth(w int) bool {
	return w >= 22
}

// barWidth computes the bar graph width based on the available terminal width.
// It also returns the effective name width and whether dBFS/CPU% should be shown.
func barWidth(termWidth int) (bw int, nameW int, showDBFS bool) {
	const outerPad = 4
	const namePad = 2
	const shortNameW = 3
	const sepAfterBar = 1
	const sepBeforeDbfs = 1
	const sepBeforeClip = 1
	const leftRightSep = 2
	const sliderW = 7

	wideOverhead := outerPad + namePad + nameWidth + sepBeforeDbfs + sepAfterBar + dbfsValueWidth + sepBeforeClip + leftRightSep + sliderW

	bw = termWidth - wideOverhead

	if bw >= 10 {
		return bw, nameWidth, true
	}

	mediumOverhead := outerPad + namePad + nameWidth + sepAfterBar + sepBeforeClip + leftRightSep + sliderW
	bw = termWidth - mediumOverhead

	if bw >= 4 {
		return bw, nameWidth, false
	}

	narrowOverhead := outerPad + namePad + shortNameW + sepAfterBar + sepBeforeClip + leftRightSep + sliderW
	bw = termWidth - narrowOverhead

	return bw, shortNameW, false
}

func shortenChannelName(name string) string {
	switch {
	case strings.HasPrefix(name, "USB "):
		return name[4:]
	case strings.HasPrefix(name, "SPDIF "):
		parts := strings.Split(name, " ")
		if len(parts) == 3 {
			return parts[1] + parts[2]
		}
		return name
	case name == "PDM Sub":
		return "Sub"
	default:
		return name
	}
}

func cpuSuffix(load int, showDBFS bool) string {
	if !showDBFS {
		return ""
	}
	return fmt.Sprintf(" %3d%%", load)
}

func (m model) View() tea.View {
	if m.err != nil {
		return renderError(m.err, m.width, m.height)
	}

	if !m.connected || len(m.devices) == 0 {
		return renderConnecting(m.width, m.height)
	}

	dev := m.devices[m.activeDevice]
	snap := m.snaps[m.activeDevice]
	channels := dspi.ChannelTable(dev.Platform())

	bw, nameW, showDBFS := barWidth(m.width)

	cpu0col := cpuColor(snap.CPU0)
	cpu1col := cpuColor(snap.CPU1)

	type groupEntry struct {
		info dspi.ChannelInfo
		peak dspi.Level
	}

	groupDisplay := map[string]string{
		"USB Input":     "Input",
		"S/PDIF Output": "Output",
		"PDM Sub":       "Subwoofer",
	}
	groupOrder := []string{"Input", "Output", "Subwoofer"}
	groups := map[string][]groupEntry{}
	channelTotal := 0

	for _, ch := range channels {
		if ch.Index >= snap.Channels {
			continue
		}
		displayName := groupDisplay[ch.Group]
		if displayName == "" {
			displayName = ch.Group
		}
		groups[displayName] = append(groups[displayName], groupEntry{ch, snap.Peaks[ch.Index]})
		channelTotal++
	}

	// Count rows needed at each compression level
	// Fixed: title + tabs + subtitle + top separator + bottom separator + footer
	fixedRows := 1 // title
	if len(m.devices) > 1 {
		fixedRows++ // tabs
	}
	fixedRows += 1 // subtitle (no blank line after, per user request)
	fixedRows += 1 // top separator
	fixedRows += 1 // bottom separator
	fixedRows += 1 // footer
	if len(m.clippedCh[m.activeDevice]) > 0 {
		fixedRows += 2 // clip text + blank
	}

	fixedRowsNoSep := fixedRows - 3 // no subtitle, no top sep, no bottom sep

	cpuRows := 4 // heading + 2 cores + blank line

	// Channel rows with full headers and blanks between sections
	chRowsFull := 0
	for i, g := range groupOrder {
		entries := groups[g]

		if len(entries) == 0 {
			continue
		}

		chRowsFull += 1 // heading
		chRowsFull += len(entries)

		if i < len(groupOrder)-1 {
			chRowsFull += 1 // blank after section
		}
	}

	chRowsCompact := channelTotal

	totalRowsCPUVisible := fixedRows + cpuRows + chRowsFull
	totalRowsCPUHidden := fixedRows + chRowsFull
	totalRowsCompact := fixedRows + chRowsCompact
	totalRowsNoSubtitle := fixedRows - 1 + chRowsCompact
	totalRowsNoSep := fixedRowsNoSep + chRowsCompact

	available := m.height - 1 // outer PaddingTop(1)

	var showCPUSection bool
	var showHeaders bool
	var showSubtitle bool
	var showTopSep bool
	var showBottomSep bool

	switch {
	case totalRowsCPUVisible <= available:
		showCPUSection = true
		showHeaders = true
		showSubtitle = true
		showTopSep = true
		showBottomSep = true
	case totalRowsCPUHidden <= available:
		showCPUSection = false
		showHeaders = true
		showSubtitle = true
		showTopSep = true
		showBottomSep = true
	case totalRowsCompact <= available:
		showCPUSection = false
		showHeaders = false
		showSubtitle = true
		showTopSep = true
		showBottomSep = true
	case totalRowsNoSubtitle <= available:
		showCPUSection = false
		showHeaders = false
		showSubtitle = false
		showTopSep = true
		showBottomSep = true
	case totalRowsNoSep <= available:
		showCPUSection = false
		showHeaders = false
		showSubtitle = false
		showTopSep = false
		showBottomSep = false
	default:
		showCPUSection = false
		showHeaders = false
		showSubtitle = false
		showTopSep = false
		showBottomSep = false
	}

	var b strings.Builder

	b.WriteString(styleTitle.Width(m.width).Render("DSPi Live Meter"))
	b.WriteString("\n")

	b.WriteString(m.renderTabs())

	if showSubtitle {
		var devicePart string
		if len(m.devices) > 1 {
			devicePart = fmt.Sprintf("Device %d/%d  |  ", m.activeDevice+1, len(m.devices))
		}
		subStr := fmt.Sprintf("%s%s  |  Serial: %s",
			devicePart, dev.Platform(), dev.Serial())
		if !showCPUSection {
			subStr += fmt.Sprintf("  |  CPU0: %d%%  CPU1: %d%%", snap.CPU0, snap.CPU1)
		}
		b.WriteString(styleSubtitle.Width(m.width).Render(subStr))
		b.WriteString("\n")
	}

	if showTopSep {
		b.WriteString(lipgloss.NewStyle().
			Foreground(colBorder).
			Render(strings.Repeat("─", m.width)))
		b.WriteString("\n")
	}

	// Build left content (CPU + channels)
	var left strings.Builder

	if showCPUSection {
		left.WriteString(lipgloss.NewStyle().
			Foreground(colMuted).
			Bold(true).
			PaddingLeft(2).
			Render("Cores"))
		left.WriteString("\n")

		cpuLabelStyle := lipgloss.NewStyle().Width(nameW).PaddingLeft(2)

		if nameW < nameWidth {
			left.WriteString(cpuLabelStyle.Render("0") + drawBar(float64(snap.CPU0)/100.0, bw, cpu0col))
			left.WriteString("\n")
			left.WriteString(cpuLabelStyle.Render("1") + drawBar(float64(snap.CPU1)/100.0, bw, cpu1col))
		} else {
			left.WriteString(cpuLabelStyle.Render("Core 0") + drawBar(float64(snap.CPU0)/100.0, bw, cpu0col) + cpuSuffix(snap.CPU0, showDBFS))
			left.WriteString("\n")
			left.WriteString(cpuLabelStyle.Render("Core 1") + drawBar(float64(snap.CPU1)/100.0, bw, cpu1col) + cpuSuffix(snap.CPU1, showDBFS))
		}

		left.WriteString("\n\n")
	}

	for i, g := range groupOrder {
		entries := groups[g]

		if len(entries) == 0 {
			continue
		}

		if showHeaders {
			left.WriteString(lipgloss.NewStyle().
				Foreground(colMuted).
				Bold(true).
				PaddingLeft(2).
				Render(g))
			left.WriteString("\n")
		}

		for _, e := range entries {
			ch := e.info
			idx := ch.Index
			peak := e.peak
			col := channelColors[idx%len(channelColors)]

			isClipped := slices.Contains(m.clippedCh[m.activeDevice], idx)

			clipMark := ""
			if isClipped {
				clipMark = " " + styleClipChannel.Render("CLIP")
			}

			chName := ch.Name
			if nameW < nameWidth {
				chName = shortenChannelName(chName)
			}

			nameStyle := lipgloss.NewStyle().
				Width(nameW).
				Foreground(col).
				Bold(true).
				PaddingLeft(2)
			bar := drawBar(peak.Linear(), bw, col)

			if showDBFS {
				dbfsStr := peak.String()
				valStyle := lipgloss.NewStyle().
					Width(dbfsValueWidth).
					Align(lipgloss.Right).
					Foreground(colMuted)
				fmt.Fprintf(&left, "%s%s %s%s",
					nameStyle.Render(chName),
					bar,
					valStyle.Render(dbfsStr),
					clipMark,
				)
			} else {
				fmt.Fprintf(&left, "%s%s%s",
					nameStyle.Render(chName),
					bar,
					clipMark,
				)
			}
			left.WriteString("\n")
		}

		if showHeaders && i < len(groupOrder)-1 {
			left.WriteString("\n")
		}
	}

	// Join left content with vertical slider
	leftStr := left.String()
	leftLines := strings.Count(leftStr, "\n")
	rightStr := m.renderMasterVolumeSlider(leftLines)

	// Compute stable max left width (theoretical, accounting for clip marks and dBFS)
	maxLeftWidth := nameW + bw + 5 // channel line without dBFS

	if showDBFS {
		maxLeftWidth = nameW + bw + 1 + dbfsValueWidth + 5 // channel with dBFS + clip
	}

	if showCPUSection {
		cpuWidth := nameW + bw

		if showDBFS {
			cpuWidth += 5 // " 100%" suffix
		}

		if cpuWidth > maxLeftWidth {
			maxLeftWidth = cpuWidth
		}
	}

	// Pad left content to push slider to the rightmost position
	contentWidth := m.width - 4 // outer PaddingLeft(2) + PaddingRight(2)
	separatorWidth := 2
	sliderWidth := 6
	targetLeftWidth := contentWidth - separatorWidth - sliderWidth

	if targetLeftWidth > maxLeftWidth {
		padWidth := targetLeftWidth - maxLeftWidth
		leftLines := strings.Split(strings.TrimSuffix(leftStr, "\n"), "\n")

		for i, line := range leftLines {
			lineWidth := lipgloss.Width(line)
			if lineWidth < maxLeftWidth+padWidth {
				leftLines[i] = line + strings.Repeat(" ", maxLeftWidth+padWidth-lineWidth)
			}
		}

		leftStr = strings.Join(leftLines, "\n") + "\n"
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftStr, "  ", rightStr))
	b.WriteString("\n")

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

	// Pad to push the bottom separator and footer to the last rows
	linesInB := strings.Count(b.String(), "\n")
	bottomLines := 1 // footer always shown
	if showBottomSep {
		bottomLines++ // bottom separator
	}
	paddingNeeded := (m.height - 1) - linesInB - bottomLines // -1 for outer top padding

	for range paddingNeeded {
		b.WriteString("\n")
	}

	if showBottomSep {
		b.WriteString(lipgloss.NewStyle().
			Foreground(colBorder).
			Render(strings.Repeat("─", m.width)))
		b.WriteString("\n")
	}
	var footer string
	if len(m.devices) > 1 {
		footer = "q: quit  |  Tab: switch device  |  ↑↓: master volume  |  c: clear clips  |  r: rescan"
	} else {
		footer = "q: quit  |  ↑↓: master volume  |  c: clear clips  |  r: rescan"
	}
	b.WriteString(styleFooter.Width(m.width).Render(footer))

	v := tea.NewView(lipgloss.NewStyle().
		Background(colBG).
		PaddingTop(1).
		PaddingLeft(2).
		PaddingRight(2).
		Render(b.String()))
	v.AltScreen = true

	return v
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

func drawBar(fraction float64, width int, color color.Color) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}

	totalSubCells := width * 8
	filledSubCells := max(0, min(totalSubCells, int(math.Round(fraction*float64(totalSubCells)))))

	full := filledSubCells / 8
	rem := filledSubCells % 8

	var bld strings.Builder
	bld.WriteString(strings.Repeat("█", full))
	if rem > 0 {
		bld.WriteString([]string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"}[rem-1])
		full++
	}
	if empty := width - full; empty > 0 {
		bld.WriteString(strings.Repeat("░", empty))
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Background(colBarBg).
		Render(bld.String())
}

func (m model) renderMasterVolumeSlider(height int) string {
	mv := m.masterVolume[m.activeDevice]

	x := (mv.DB() + 128.0) / 128.0
	x = max(0, min(1, x))
	fraction := math.Pow(x, 3.19)

	barHeight := max(height-2, 1)

	barStyle := lipgloss.NewStyle().Foreground(colVolume).Background(colBarBg)

	partials := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	effectiveFilled := fraction * float64(barHeight)
	fullFilled := int(effectiveFilled)
	remainder := effectiveFilled - float64(fullFilled)

	var bld strings.Builder

	bld.WriteString(barStyle.Render("  VOL "))
	bld.WriteString("\n")

	for i := range barHeight {
		distFromBottom := barHeight - 1 - i

		var ch string
		switch {
		case distFromBottom < fullFilled:
			ch = "█"
		case distFromBottom == fullFilled && remainder > 0:
			idx := int(math.Round(remainder * 8))

			if idx == 0 {
				ch = "░"
			} else {
				ch = partials[idx-1]
			}
		default:
			ch = "░"
		}

		bld.WriteString(barStyle.Render(fmt.Sprintf("  %s  ", ch)))
		bld.WriteString("\n")
	}

	bld.WriteString(barStyle.Render(fmt.Sprintf("%6s", mv.String())))

	return bld.String()
}

func cpuColor(load int) color.Color {
	switch {
	case load >= 90:
		return colCPUCrit
	case load >= 60:
		return colCPUWarn
	default:
		return colCPUOK
	}
}

func renderFrame(width int, height int, body string) tea.View {
	v := tea.NewView(lipgloss.NewStyle().
		Background(colBG).
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styleTitle.Width(width).Render("DSPi Live Meter"),
				"",
				body,
			),
		))
	v.AltScreen = true

	return v
}

func renderConnecting(width int, height int) tea.View {
	return renderFrame(width, height,
		lipgloss.JoinVertical(lipgloss.Center,
			"",
			lipgloss.NewStyle().Foreground(colYellow).Render("Connecting to DSPi..."),
			"",
			lipgloss.NewStyle().Foreground(colMuted).Render("Make sure the device(s) are plugged in via USB"),
		),
	)
}

func renderError(err error, width int, height int) tea.View {
	return renderFrame(width, height,
		lipgloss.JoinVertical(lipgloss.Center,
			"",
			lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("Connection Error"),
			"",
			lipgloss.NewStyle().Foreground(colFG).Render(err.Error()),
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
	p := tea.NewProgram(initialModel())

	_, err := p.Run()

	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	return nil
}
