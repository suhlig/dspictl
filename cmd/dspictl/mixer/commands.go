package mixer

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/suhlig/dspi"
)

const (
	tickInterval      = 60 * time.Millisecond
	scanInterval      = 100
	clipTimerDuration = 170
)

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func connectCmd(targetSerial string) tea.Cmd {
	return func() tea.Msg {
		var devs []*dspi.Device
		var err error

		if targetSerial != "" {
			dev, err := dspi.Open(dspi.DeviceInfo{Serial: targetSerial})
			if err != nil {
				return errMsg{fmt.Errorf("connect: %w", err)}
			}
			devs = []*dspi.Device{dev}
		} else {
			devs, err = dspi.OpenAll()
			if err != nil {
				return errMsg{fmt.Errorf("connect: %w", err)}
			}
		}

		return devicesMsg(devs)
	}
}

func rescanCmd(targetSerial string) tea.Cmd {
	return func() tea.Msg {
		var devs []*dspi.Device
		var err error

		if targetSerial != "" {
			dev, err := dspi.Open(dspi.DeviceInfo{Serial: targetSerial})
			if err != nil {
				return errMsg{fmt.Errorf("rescan: %w", err)}
			}
			devs = []*dspi.Device{dev}
		} else {
			devs, err = dspi.OpenAll()
			if err != nil {
				return errMsg{fmt.Errorf("rescan: %w", err)}
			}
		}

		return devicesMsg(devs)
	}
}
