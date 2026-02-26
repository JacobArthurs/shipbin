/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	flagName       string
	flagArtifacts  []string
	flagVersion    string
	flagSummary    string
	flagLicense    string
	flagDryRun     bool
	flagProvenance bool
	flagReadme     string
)

var rootCmd = &cobra.Command{
	Use:          "shipbin",
	SilenceUsage: true,
	Short:        "Ship binaries to npm and PyPI",
	Long: `Publishes pre-built binaries to npm and PyPI.

Assembles platform-specific packages from the provided artifacts,
then publishes them to the target registry.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagName, "name", "", "binary name")
	rootCmd.PersistentFlags().StringArrayVar(&flagArtifacts, "artifact", nil, "os/arch:path mapping (repeatable)")
	rootCmd.PersistentFlags().StringVar(&flagVersion, "version", "", "release version (defaults to current git tag)")
	rootCmd.PersistentFlags().StringVar(&flagSummary, "summary", "", "short description of the package (optional)")
	rootCmd.PersistentFlags().StringVar(&flagLicense, "license", "", "license identifier (e.g. MIT, Apache-2.0)")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print what would be published without publishing")
	rootCmd.PersistentFlags().BoolVar(&flagProvenance, "provenance", true, "publish with provenance attestation (requires CI environment)")
	rootCmd.PersistentFlags().StringVar(&flagReadme, "readme", "", "path to README to include in the published package (optional)")

	rootCmd.MarkPersistentFlagRequired("name")
	rootCmd.MarkPersistentFlagRequired("artifact")

	rootCmd.AddCommand(npmCmd)
	rootCmd.AddCommand(pypiCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
