/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/pypi"
	"github.com/spf13/cobra"
)

var (
	flagPyPIPackage string
)

var pypiCmd = &cobra.Command{
	Use:   "pypi",
	Short: "Publish binaries to PyPI",
	Long: `Publishes pre-built binaries to PyPI.

Builds a platform-specific wheel for each artifact containing the binary and a
Python shim that locates and executes it. Users install the package with pip and
the correct wheel is resolved automatically based on their platform.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := buildPypiConfig()
		if err != nil {
			return err
		}
		return pypi.Publish(cfg)
	},
}

func buildPypiConfig() (*pypi.Config, error) {
	version, err := config.ResolveVersion(flagVersion)
	if err != nil {
		return nil, err
	}

	artifacts, err := config.ParseArtifacts(flagArtifacts)
	if err != nil {
		return nil, err
	}

	cfg := &pypi.Config{
		Name:        flagName,
		Version:     version,
		PyPIPackage: flagPyPIPackage,
		Artifacts:   artifacts,
		Summary:     flagSummary,
		License:     flagLicense,
		Readme:      flagReadme,
		DryRun:      flagDryRun,
	}

	return cfg, nil
}

func init() {
	pypiCmd.Flags().StringVar(&flagPyPIPackage, "package", "", "PyPI package name (e.g. myapp)")

	pypiCmd.MarkFlagRequired("package")
}
