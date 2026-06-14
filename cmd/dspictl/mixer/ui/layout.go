package ui

import "github.com/suhlig/dspi"

type GroupEntry struct {
	Info dspi.ChannelInfo
	Peak dspi.Level
}

func BuildGroups(channels []dspi.ChannelInfo, snap dspi.MeterSnapshot) (groups map[string][]GroupEntry, groupOrder []string, channelTotal int) {
	groupDisplay := map[string]string{
		"USB Input":     "Input",
		"S/PDIF Input":  "Input",
		"I2S Input":     "Input",
		"S/PDIF Output": "Output",
		"PDM Sub":       "Subwoofer",
	}
	groupOrder = []string{"Input", "Output", "Subwoofer"}
	groups = map[string][]GroupEntry{}

	for _, ch := range channels {
		if ch.Index >= snap.Channels {
			continue
		}

		displayName := groupDisplay[ch.Group]
		if displayName == "" {
			displayName = ch.Group
		}
		groups[displayName] = append(groups[displayName], GroupEntry{ch, snap.Peaks[ch.Index]})
		channelTotal++
	}

	return
}

type Layout struct {
	ShowCPUSection bool
	ShowHeaders    bool
	ShowSubtitle   bool
	ShowTopSep     bool
	ShowBottomSep  bool
	BarWidth       int
	NameWidth      int
	ShowDBFS       bool
}

func ComputeLayout(termWidth, termHeight int, numDevices int, hasClip bool, groups map[string][]GroupEntry, channelTotal int) Layout {
	bw, nameW, showDBFS := barWidth(termWidth)

	const cpuRows = 4 // heading + 2 cores + blank line

	fixedRows := 1 // title

	if numDevices > 1 {
		fixedRows++ // tabs
	}

	fixedRows += 3 // subtitle + top separator + bottom separator
	fixedRows += 1 // footer

	if hasClip {
		fixedRows += 2 // clip text + blank
	}

	fixedRowsNoSep := fixedRows - 3

	groupOrder := []string{"Input", "Output", "Subwoofer"}
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
	available := termHeight - 1

	totalRowsCPUVisible := fixedRows + cpuRows + chRowsFull
	totalRowsCPUHidden := fixedRows + chRowsFull
	totalRowsCompact := fixedRows + chRowsCompact
	totalRowsNoSubtitle := fixedRows - 1 + chRowsCompact
	totalRowsNoSep := fixedRowsNoSep + chRowsCompact

	l := Layout{BarWidth: bw, NameWidth: nameW, ShowDBFS: showDBFS}

	switch {
	case totalRowsCPUVisible <= available:
		l.ShowCPUSection = true
		l.ShowHeaders = true
		l.ShowSubtitle = true
		l.ShowTopSep = true
		l.ShowBottomSep = true
	case totalRowsCPUHidden <= available:
		l.ShowHeaders = true
		l.ShowSubtitle = true
		l.ShowTopSep = true
		l.ShowBottomSep = true
	case totalRowsCompact <= available:
		l.ShowSubtitle = true
		l.ShowTopSep = true
		l.ShowBottomSep = true
	case totalRowsNoSubtitle <= available:
		l.ShowTopSep = true
		l.ShowBottomSep = true
	case totalRowsNoSep <= available:
		l.ShowTopSep = false
		l.ShowBottomSep = false
	}

	return l
}

func barWidth(termWidth int) (bw int, nameW int, showDBFS bool) {
	const nameWidth = 14
	const dbfsValueWidth = 10
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

func AcceptWidth(w int) bool {
	return w >= 22
}
