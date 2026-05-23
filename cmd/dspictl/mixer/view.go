package mixer

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/suhlig/dspi/cmd/dspictl/mixer/ui"
)

func (m model) View() tea.View {
	if m.err != nil {
		return ui.RenderError(m.err, m.width, m.height)
	}

	if !m.connected || m.dm.Len() == 0 {
		return ui.RenderConnecting(m.width, m.height)
	}

	dev := m.dm.Device(m.activeDevice)
	snap := m.dm.Snap(m.activeDevice)
	channels := m.dm.Channels(m.activeDevice)

	groups, groupOrder, channelTotal := ui.BuildGroups(channels, snap)
	clippedCh := m.dm.ClippedCh(m.activeDevice)
	layout := ui.ComputeLayout(m.width, m.height, m.dm.Len(), len(clippedCh) > 0, groups, channelTotal)

	var b strings.Builder

	b.WriteString(ui.StyleTitle.Width(m.width).Render("DSPi Mixer"))
	b.WriteString("\n")

	b.WriteString(ui.RenderTabs(m.dm.AllDevices(), m.dm.AllClippedCh(), m.activeDevice))

	if layout.ShowSubtitle {
		var devicePart string

		if m.dm.Len() > 1 {
			devicePart = fmt.Sprintf("Device %d/%d  |  ", m.activeDevice+1, m.dm.Len())
		}

		subStr := fmt.Sprintf("%s%s  |  Serial: %s",
			devicePart, dev.Platform(), dev.Serial())

		if !layout.ShowCPUSection {
			subStr += fmt.Sprintf("  |  CPU0: %d%%  CPU1: %d%%", snap.CPU0, snap.CPU1)
		}

		b.WriteString(ui.StyleSubtitle.Width(m.width).Render(subStr))
		b.WriteString("\n")
	}

	if layout.ShowTopSep {
		b.WriteString(lipgloss.NewStyle().
			Foreground(ui.ColBorder).
			Render(strings.Repeat("─", m.width)))
		b.WriteString("\n")
	}

	var left strings.Builder

	if layout.ShowCPUSection {
		left.WriteString(ui.RenderCPUSection(snap.CPU0, snap.CPU1, layout))
	}

	for i, g := range groupOrder {
		entries := groups[g]

		if len(entries) == 0 {
			continue
		}

		left.WriteString(ui.RenderChannelGroup(g, entries, layout, clippedCh, layout.ShowHeaders, ui.ChannelColors))

		if layout.ShowHeaders && i < len(groupOrder)-1 {
			left.WriteString("\n")
		}
	}

	leftStr := left.String()
	leftLines := strings.Count(leftStr, "\n")
	rightStr := ui.RenderMasterVolumeSlider(m.dm.MasterVolume(m.activeDevice), leftLines)

	maxLeftWidth := layout.NameWidth + layout.BarWidth + 5

	if layout.ShowDBFS {
		maxLeftWidth = layout.NameWidth + layout.BarWidth + 1 + 10 + 5
	}

	if layout.ShowCPUSection {
		cpuWidth := layout.NameWidth + layout.BarWidth

		if layout.ShowDBFS {
			cpuWidth += 5
		}

		if cpuWidth > maxLeftWidth {
			maxLeftWidth = cpuWidth
		}
	}

	contentWidth := m.width - 4
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

	b.WriteString(ui.RenderClipSection(clippedCh, channels))

	linesInB := strings.Count(b.String(), "\n")
	bottomLines := 1

	if layout.ShowBottomSep {
		bottomLines++
	}

	paddingNeeded := (m.height - 1) - linesInB - bottomLines

	for range paddingNeeded {
		b.WriteString("\n")
	}

	if layout.ShowBottomSep {
		b.WriteString(lipgloss.NewStyle().
			Foreground(ui.ColBorder).
			Render(strings.Repeat("─", m.width)))
		b.WriteString("\n")
	}

	var footer string

	if m.dm.Len() > 1 {
		footer = "q: quit  |  Tab: switch device  |  ↑↓: volume  |  m: mute  |  M: mute all  |  c: clear clips  |  r: rescan"
	} else {
		footer = "q: quit  |  ↑↓: volume  |  m: mute  |  c: clear clips  |  r: rescan"
	}

	b.WriteString(ui.StyleFooter.Width(m.width).Render(footer))

	v := tea.NewView(lipgloss.NewStyle().
		Background(ui.ColBG).
		PaddingTop(1).
		PaddingLeft(2).
		PaddingRight(2).
		Render(b.String()))
	v.AltScreen = true

	return v
}
