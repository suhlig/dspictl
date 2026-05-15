package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func RenderFrame(width int, height int, body string) tea.View {
	v := tea.NewView(lipgloss.NewStyle().
		Background(ColBG).
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				StyleTitle.Width(width).Render("DSPi Live Meter"),
				"",
				body,
			),
		))
	v.AltScreen = true

	return v
}

func RenderConnecting(width int, height int) tea.View {
	return RenderFrame(width, height,
		lipgloss.JoinVertical(lipgloss.Center,
			"",
			lipgloss.NewStyle().Foreground(ColYellow).Render("Connecting to DSPi..."),
			"",
			lipgloss.NewStyle().Foreground(ColMuted).Render("Make sure the device(s) are plugged in via USB"),
		),
	)
}

func RenderError(err error, width int, height int) tea.View {
	return RenderFrame(width, height,
		lipgloss.JoinVertical(lipgloss.Center,
			"",
			lipgloss.NewStyle().Foreground(ColRed).Bold(true).Render("Connection Error"),
			"",
			lipgloss.NewStyle().Foreground(ColFG).Render(err.Error()),
			"",
			lipgloss.NewStyle().Foreground(ColMuted).Render("Press 'r' to retry, 'q' to quit"),
		),
	)
}
