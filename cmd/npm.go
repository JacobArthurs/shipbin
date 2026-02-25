/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/npm"
	"github.com/spf13/cobra"
)

var (
	flagOrg string
	flagTag string
)

var npmCmd = &cobra.Command{
	Use:   "npm",
	Short: "Publish binaries to npm",
	Long: `Publishes pre-built binaries to npm.

Builds a platform-specific package for each artifact (e.g. @org/name-linux-x64)
containing the binary, then publishes a root package (e.g. name) that declares
them as optional dependencies. Users install the root package and npm resolves
the correct platform package automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := buildNpmConfig()
		if err != nil {
			return err
		}
		return npm.Publish(cfg)
	},
}

func buildNpmConfig() (*npm.Config, error) {
	version, err := config.ResolveVersion(flagVersion)
	if err != nil {
		return nil, err
	}

	artifacts, err := config.ParseArtifacts(flagArtifacts)
	if err != nil {
		return nil, err
	}

	cfg := &npm.Config{
		Name:       flagName,
		Version:    version,
		License:    flagLicense,
		Artifacts:  artifacts,
		DryRun:     flagDryRun,
		Org:        flagOrg,
		Tag:        flagTag,
		Provenance: flagProvenance,
		Readme:     flagReadme,
	}

	return cfg, nil
}

func init() {
	npmCmd.Flags().StringVar(&flagOrg, "org", "", "npm org scope (e.g. 'myorg' produces @myorg/name-linux-x64)")
	npmCmd.Flags().StringVar(&flagTag, "tag", "latest", "dist-tag to publish under (e.g. latest, next, beta)")

	npmCmd.MarkFlagRequired("org")
}
