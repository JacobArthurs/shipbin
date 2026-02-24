package config

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

const (
	TargetAll  string = "all"
	TargetNpm  string = "npm"
	TargetPyPI string = "pypi"
)

type Config struct {
	Name        string
	Version     string
	Artifacts   []Artifact
	NpmOrg      string
	PyPIPackage string
	Targets     []string
}

func ResolveVersion(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("--version not provided and no exact git tag found: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		return "", fmt.Errorf("--version not provided and git describe returned empty output")
	}

	return version, nil
}

func ResolveTargets(raw string) ([]string, error) {
	switch strings.ToLower(raw) {
	case TargetAll:
		return []string{TargetNpm, TargetPyPI}, nil
	case TargetNpm:
		return []string{TargetNpm}, nil
	case TargetPyPI:
		return []string{TargetPyPI}, nil
	default:
		return nil, fmt.Errorf("invalid --target %q: must be one of: all, npm, pypi", raw)
	}
}

func (c *Config) Validate() error {
	var errs []error

	if slices.Contains(c.Targets, TargetNpm) {
		if c.NpmOrg == "" {
			errs = append(errs, fmt.Errorf("--npm-org is required when target is npm or all"))
		}
	}

	if slices.Contains(c.Targets, TargetPyPI) {
		if c.PyPIPackage == "" {
			errs = append(errs, fmt.Errorf("--pypi-package is required when target is pypi or all"))
		}
	}

	return errors.Join(errs...)
}
