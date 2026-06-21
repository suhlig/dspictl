package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newManCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "man [output-dir]",
		Short:  "Generate man pages (internal use)",
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

	header := &doc.GenManHeader{
		Title:   "dspictl",
		Section: "1",
		Source:  "dspictl",
	}

	root := newRootCmd()
	return doc.GenManTree(root, header, dir)
}
