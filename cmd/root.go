/*
Copyright © 2026 JACOB ARTHURS
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/platforms"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

var (
	flagBinary      string
	flagDist        string
	flagVersion     string
	flagPlatforms   []string
	flagNpmOrg      string
	flagPyPIPackage string
	flagTarget      string
)

var rootCmd = &cobra.Command{
	Use:   "shipbin",
	Short: "Publish Go binaries to npm and PyPI from GitHub Actions",
	Long: `shipbin publishes pre-built Go binaries to npm and PyPI.

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
	rootCmd.PersistentFlags().StringVar(&flagBinary, "binary", "", "binary name (defaults to Go module name)")
	rootCmd.PersistentFlags().StringVar(&flagDist, "dist", "./dist", "directory containing GoReleaser output")
	rootCmd.PersistentFlags().StringVar(&flagVersion, "version", "", "release version (defaults to current git tag)")
	rootCmd.PersistentFlags().StringSliceVar(&flagPlatforms, "platforms", nil, "comma-separated os/arch pairs (default: all 5 standard platforms)")
	rootCmd.PersistentFlags().StringVar(&flagNpmOrg, "npm-org", "", "npm org scope (e.g. 'myorg' produces @myorg/linux-x64)")
	rootCmd.PersistentFlags().StringVar(&flagPyPIPackage, "pypi-package", "", "PyPI package name")
	rootCmd.PersistentFlags().StringVar(&flagTarget, "target", "all", "publish target: all, npm, or pypi")
}

func buildConfig() (*config.Config, error) {
	binary := flagBinary
	if binary == "" {
		resolved, err := resolveModuleName()
		if err != nil {
			return nil, fmt.Errorf("--binary not provided and could not infer from go.mod: %w", err)
		}
		binary = resolved
	}

	version, err := config.ResolveVersion(flagVersion)
	if err != nil {
		return nil, err
	}

	target, err := config.ResolveTarget(flagTarget)
	if err != nil {
		return nil, err
	}

	mappings := platforms.All()
	if len(flagPlatforms) > 0 {
		mappings, err = platforms.Parse(flagPlatforms)
		if err != nil {
			return nil, err
		}
	}

	cfg := &config.Config{
		Binary:    binary,
		Dist:      flagDist,
		Version:   version,
		Platforms: mappings,
		Npm:       config.NpmConfig{Org: flagNpmOrg},
		PyPI:      config.PyPIConfig{Package: flagPyPIPackage},
		Target:    target,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func resolveModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("could not read go.mod: %w", err)
	}

	f, err := modfile.ParseLax("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("could not parse go.mod: %w", err)
	}

	return lastPathSegment(f.Module.Mod.Path), nil
}

func lastPathSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
