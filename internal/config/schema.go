package config

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jacobarthurs/shipbin/internal/platforms"
)

type Target string

const (
	TargetAll  Target = "all"
	TargetNpm  Target = "npm"
	TargetPyPI Target = "pypi"
)

type Config struct {
	Binary    string
	Dist      string
	Version   string
	Platforms []platforms.Mapping
	Npm       NpmConfig
	PyPI      PyPIConfig
	Target    Target
}

type NpmConfig struct {
	Org string
}

type PyPIConfig struct {
	Package string
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

func ResolveTarget(raw string) (Target, error) {
	switch Target(strings.ToLower(raw)) {
	case TargetAll:
		return TargetAll, nil
	case TargetNpm:
		return TargetNpm, nil
	case TargetPyPI:
		return TargetPyPI, nil
	default:
		return "", fmt.Errorf("invalid --target %q: must be one of: all, npm, pypi", raw)
	}
}

func (c *Config) Validate() error {
	var errs []error

	if c.Binary == "" {
		errs = append(errs, fmt.Errorf("--binary is required"))
	}
	if c.Version == "" {
		errs = append(errs, fmt.Errorf("version could not be resolved: provide --version or run from a tagged commit"))
	}
	if c.Dist == "" {
		errs = append(errs, fmt.Errorf("--dist is required"))
	}
	if len(c.Platforms) == 0 {
		errs = append(errs, fmt.Errorf("no platforms configured"))
	}

	if c.Target == TargetAll || c.Target == TargetNpm {
		if c.Npm.Org == "" {
			errs = append(errs, fmt.Errorf("--npm-org is required when target is npm or all"))
		}
	}

	if c.Target == TargetAll || c.Target == TargetPyPI {
		if c.PyPI.Package == "" {
			errs = append(errs, fmt.Errorf("--pypi-package is required when target is pypi or all"))
		}
	}

	return errors.Join(errs...)
}
