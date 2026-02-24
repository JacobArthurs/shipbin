/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish Go binaries to npm and PyPI",
	Long: `Publishes pre-built Go binaries from dist/ to npm and PyPI.

Assembles platform-specific npm packages and PyPI wheels from
GoReleaser output, then publishes them concurrently. npm and
PyPI errors are both surfaced if either fails.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("publish called")
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
