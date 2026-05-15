package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/suhlig/dspi"
)

func RenderTabs(devices []*dspi.Device, clippedCh [][]int, activeDevice int) string {
	if len(devices) <= 1 {
		return ""
	}

	var tabs []string

	for i, dev := range devices {
		serial := dev.Serial()
		label := fmt.Sprintf("%d: %s", i+1, serial)

		hasClip := len(clippedCh[i]) > 0

		if i == activeDevice && hasClip {
			tabs = append(tabs, StyleTabClip.Render(label))
		} else if i == activeDevice {
			tabs = append(tabs, StyleTabActive.Render(label))
		} else if hasClip {
			tabs = append(tabs, StyleTabClip.Render(label))
		} else {
			tabs = append(tabs, StyleTabInactive.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}
