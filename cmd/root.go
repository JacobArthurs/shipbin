/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "godist",
	Short: "Publish Go binaries to npm and PyPI from GitHub Actions",
	Long: `godist publishes pre-built Go binaries to npm and PyPI.

It reads binaries from your dist/ directory (GoReleaser output),
assembles platform-specific npm packages and PyPI wheels, and
publishes them in a single command.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
