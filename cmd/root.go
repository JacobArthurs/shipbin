/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"os"

	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/publisher"
	"github.com/spf13/cobra"
)

var (
	flagName          string
	flagArtifacts     []string
	flagVersion       string
	flagLicense       string
	flagNpmOrg        string
	flagNpmTag        string
	flagNpmProvenance bool
	flagNpmReadme     string
	flagPyPIPackage   string
	flagTarget        string
	flagDryRun        bool
)

var rootCmd = &cobra.Command{
	Use:   "shipbin",
	Short: "Ship binaries to npm and PyPI",
	Long: `Publishes pre-built binaries to npm and PyPI.

Assembles platform-specific npm packages and PyPI wheels from the
provided artifacts, then publishes them concurrently.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := buildConfig()
		if err != nil {
			return err
		}

		return publisher.Publish(cfg)
	},
}

func buildConfig() (*config.Config, error) {
	version, err := config.ResolveVersion(flagVersion)
	if err != nil {
		return nil, err
	}

	targets, err := config.ResolveTargets(flagTarget)
	if err != nil {
		return nil, err
	}

	artifacts, err := config.ParseArtifacts(flagArtifacts)
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{
		Name:          flagName,
		Version:       version,
		License:       flagLicense,
		Artifacts:     artifacts,
		NpmOrg:        flagNpmOrg,
		NpmTag:        flagNpmTag,
		NpmProvenance: flagNpmProvenance,
		NpmReadme:     flagNpmReadme,
		PyPIPackage:   flagPyPIPackage,
		Targets:       targets,
		DryRun:        flagDryRun,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func init() {
	rootCmd.Flags().StringVar(&flagName, "name", "", "binary name")
	rootCmd.Flags().StringArrayVar(&flagArtifacts, "artifact", nil, "os/arch:path mapping (repeatable)")
	rootCmd.Flags().StringVar(&flagVersion, "version", "", "release version (defaults to current git tag)")
	rootCmd.Flags().StringVar(&flagLicense, "license", "", "license identifier (e.g. MIT, Apache-2.0)")
	rootCmd.Flags().StringVar(&flagNpmOrg, "npm-org", "", "npm org scope (e.g. 'myorg' produces @myorg/linux-x64)")
	rootCmd.Flags().StringVar(&flagNpmTag, "npm-tag", "latest", "npm dist-tag to publish under (e.g. latest, next, beta)")
	rootCmd.Flags().BoolVar(&flagNpmProvenance, "npm-provenance", true, "publish with npm provenance attestation (requires CI environment)")
	rootCmd.Flags().StringVar(&flagNpmReadme, "npm-readme", "", "path to README to include in the root npm package (optional)")
	rootCmd.Flags().StringVar(&flagPyPIPackage, "pypi-package", "", "PyPI package name")
	rootCmd.Flags().StringVar(&flagTarget, "target", "all", "publish target: all, npm, or pypi")
	rootCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "print what would be published without publishing")

	rootCmd.MarkFlagRequired("name")
	rootCmd.MarkFlagRequired("artifact")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
