package ui

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/suhlig/dspi"
)

func RenderCPUSection(cpu0, cpu1 int, layout Layout) string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().
		Foreground(ColMuted).
		Bold(true).
		PaddingLeft(2).
		Render("Cores"))
	b.WriteString("\n")

	cpuLabelStyle := lipgloss.NewStyle().Width(layout.NameWidth).PaddingLeft(2)

	cpu0col := CPUColor(cpu0)
	cpu1col := CPUColor(cpu1)

	if layout.NameWidth < 14 {
		b.WriteString(cpuLabelStyle.Render("0") + DrawBar(float64(cpu0)/100.0, layout.BarWidth, cpu0col))
		b.WriteString("\n")
		b.WriteString(cpuLabelStyle.Render("1") + DrawBar(float64(cpu1)/100.0, layout.BarWidth, cpu1col))
	} else {
		b.WriteString(cpuLabelStyle.Render("Core 0") + DrawBar(float64(cpu0)/100.0, layout.BarWidth, cpu0col) + CPUSuffix(cpu0, layout.ShowDBFS))
		b.WriteString("\n")
		b.WriteString(cpuLabelStyle.Render("Core 1") + DrawBar(float64(cpu1)/100.0, layout.BarWidth, cpu1col) + CPUSuffix(cpu1, layout.ShowDBFS))
	}

	b.WriteString("\n\n")

	return b.String()
}

func RenderChannelGroup(g string, entries []GroupEntry, layout Layout, clippedCh []int, showHeaders bool, channelColors []color.Color) string {
	var b strings.Builder

	if showHeaders {
		b.WriteString(lipgloss.NewStyle().
			Foreground(ColMuted).
			Bold(true).
			PaddingLeft(2).
			Render(g))
		b.WriteString("\n")
	}

	for _, e := range entries {
		ch := e.Info
		idx := ch.Index
		peak := e.Peak
		col := channelColors[idx%len(channelColors)]

		isClipped := slices.Contains(clippedCh, idx)

		clipMark := ""
		if isClipped {
			clipMark = " " + StyleClipChannel.Render("CLIP")
		}

		chName := ch.Name
		if layout.NameWidth < 14 {
			chName = ShortenChannelName(chName)
		}

		nameStyle := lipgloss.NewStyle().
			Width(layout.NameWidth).
			Foreground(col).
			Bold(true).
			PaddingLeft(2)
		bar := DrawBar(peak.Linear(), layout.BarWidth, col)

		if layout.ShowDBFS {
			dbfsStr := peak.String()
			valStyle := lipgloss.NewStyle().
				Width(10).
				Align(lipgloss.Right).
				Foreground(ColMuted)
			fmt.Fprintf(&b, "%s%s %s%s",
				nameStyle.Render(chName),
				bar,
				valStyle.Render(dbfsStr),
				clipMark,
			)
		} else {
			fmt.Fprintf(&b, "%s%s%s",
				nameStyle.Render(chName),
				bar,
				clipMark,
			)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func RenderClipSection(clippedCh []int, channels []dspi.ChannelInfo) string {
	if len(clippedCh) == 0 {
		return ""
	}

	var clipNames []string

	for _, idx := range clippedCh {
		if idx < len(channels) {
			clipNames = append(clipNames, channels[idx].Name)
		}
	}

	var b strings.Builder
	b.WriteString(StyleClipLabel.Render("CLIP:"))
	b.WriteString(" ")

	for _, name := range clipNames {
		b.WriteString(StyleClipChannel.Render(name))
		b.WriteString(" ")
	}

	b.WriteString("\n\n")

	return b.String()
}
