package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/suhlig/dspi"
)

func DrawBar(fraction float64, width int, color color.Color) string {
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
		bld.WriteString([]string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"}[rem-1]) //#nosec G602 -- rem is 1..7 (rem = filledSubCells % 8)
		full++
	}
	if empty := width - full; empty > 0 {
		bld.WriteString(strings.Repeat("░", empty))
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Background(ColBarBg).
		Render(bld.String())
}

func CPUColor(load int) color.Color {
	switch {
	case load >= 90:
		return ColCPUCrit
	case load >= 60:
		return ColCPUWarn
	default:
		return ColCPUOK
	}
}

func CPUSuffix(load int, showDBFS bool) string {
	if !showDBFS {
		return ""
	}
	return fmt.Sprintf(" %3d%%", load)
}

func ShortenChannelName(name string) string {
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

func RenderMasterVolumeSlider(mv dspi.Gain, height int) string {
	x := (mv.DB() + 128.0) / 128.0
	x = max(0, min(1, x))
	fraction := math.Pow(x, 3.19)

	barHeight := max(height-2, 1)

	barStyle := lipgloss.NewStyle().Foreground(ColVolume).Background(ColBarBg)

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
