package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

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
