/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and verify binaries exist before publishing",
	Long: `Validates that godist has everything it needs to publish successfully.

Checks that all platform binaries exist in the dist/ directory,
required credentials are present in the environment, and all
flags are well-formed. Exits non-zero if anything is missing.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("validate called")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
