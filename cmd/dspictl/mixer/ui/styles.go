package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	ColBG      = lipgloss.Color("#1a1b26")
	ColFG      = lipgloss.Color("#a9b1d6")
	ColMuted   = lipgloss.Color("#565f89")
	ColGreen   = lipgloss.Color("#9ece6a")
	ColYellow  = lipgloss.Color("#e0af68")
	ColRed     = lipgloss.Color("#f7768e")
	ColBlue    = lipgloss.Color("#7aa2f7")
	ColCyan    = lipgloss.Color("#73daca")
	ColPurple  = lipgloss.Color("#bb9af7")
	ColClipBg  = lipgloss.Color("#340000")
	ColBarBg   = lipgloss.Color("#292e42")
	ColBorder  = lipgloss.Color("#3b4261")
	ColTitle   = lipgloss.Color("#c0caf5")
	ColVolume  = lipgloss.Color("#7aa2f7")
	ColCPUOK   = lipgloss.Color("#9ece6a")
	ColCPUWarn = lipgloss.Color("#e0af68")
	ColCPUCrit = lipgloss.Color("#f7768e")
)

var (
	StyleTitle = lipgloss.NewStyle().
			Foreground(ColTitle).
			Bold(true).
			Background(ColBG).
			Align(lipgloss.Center)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColMuted).
			Background(ColBG).
			Align(lipgloss.Center)

	StyleClipLabel = lipgloss.NewStyle().
			Foreground(ColRed).
			Bold(true).
			PaddingLeft(2)

	StyleClipChannel = lipgloss.NewStyle().
				Background(ColClipBg).
				Foreground(ColRed).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)

	StyleFooter = lipgloss.NewStyle().
			Foreground(ColMuted).
			Background(ColBG).
			Align(lipgloss.Center)

	StyleTabActive = lipgloss.NewStyle().
			Foreground(ColTitle).
			Bold(true).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(ColMuted).
				Padding(0, 2)

	StyleTabClip = lipgloss.NewStyle().
			Foreground(ColRed).
			Bold(true).
			Padding(0, 1)
)

var ChannelColors = []color.Color{
	ColBlue,
	ColRed,
	ColCyan,
	ColGreen,
	ColCyan,
	ColGreen,
	ColCyan,
	ColGreen,
	ColCyan,
	ColGreen,
	ColPurple,
}
