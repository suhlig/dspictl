package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cpuguy83/go-md2man/v2/md2man"
	"github.com/spf13/cobra"
	manpage "github.com/suhlig/dspi/man"
)

func newManCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "man [output-dir]",
		Short:  "Generate the man page",
		Hidden: true,
		Args:   cobra.MaximumNArgs(1),
		RunE:   runMan,
	}
}

func runMan(cmd *cobra.Command, args []string) error {
	dir := "man"

	if len(args) > 0 {
		dir = args[0]
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	rendered := md2man.Render([]byte(manpage.Content))

	f, err := os.Create(filepath.Join(dir, "dspictl.1"))
	if err != nil {
		return fmt.Errorf("creating man page: %w", err)
	}
	defer f.Close() //nolint:errcheck

	if _, err := f.Write(rendered); err != nil {
		return fmt.Errorf("writing man page: %w", err)
	}

	return nil
}
