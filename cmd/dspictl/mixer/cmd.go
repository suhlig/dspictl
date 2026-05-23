package mixer

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mixer",
		Short: "Interactive mixer TUI",
		RunE:  runMixer,
	}
}

func runMixer(cmd *cobra.Command, args []string) error {
	targetSerial, _ := cmd.Flags().GetString("target")

	p := tea.NewProgram(newModel(targetSerial))

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("running mixer: %w", err)
	}

	return nil
}
