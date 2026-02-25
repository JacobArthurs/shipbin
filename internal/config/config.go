package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/jacobarthurs/shipbin/internal/platforms"
)

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9][a-zA-Z0-9.-]*)?$`)

type Artifact struct {
	Platform platforms.Platform
	Mapping  platforms.Mapping
	Path     string
}

func ParseArtifacts(artifacts []string) ([]Artifact, error) {
	seen := make(map[platforms.Platform]string)
	var results []Artifact
	var errs []error

	for _, entry := range artifacts {
		platformStr, path, ok := strings.Cut(entry, ":")
		if !ok {
			errs = append(errs, fmt.Errorf("invalid --artifact %q: expected os/arch:path", entry))
			continue
		}

		goos, goarch, ok := strings.Cut(platformStr, "/")
		if !ok {
			errs = append(errs, fmt.Errorf("invalid --artifact %q: platform must be os/arch", entry))
			continue
		}

		m, err := platforms.Lookup(goos, goarch)
		if err != nil {
			errs = append(errs, fmt.Errorf("--artifact %q: %w", entry, err))
			continue
		}

		p := platforms.Platform{GOOS: goos, GOARCH: goarch}

		if prev, ok := seen[p]; ok {
			errs = append(errs, fmt.Errorf("duplicate --artifact for %s/%s: %q and %q", goos, goarch, prev, entry))
			continue
		}
		seen[p] = entry

		info, err := os.Stat(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("--artifact %q: %w", entry, err))
			continue
		}
		if info.IsDir() {
			errs = append(errs, fmt.Errorf("--artifact %q: path is a directory, not a file", entry))
			continue
		}
		if runtime.GOOS != "windows" && goos != "windows" && info.Mode()&0111 == 0 {
			errs = append(errs, fmt.Errorf("--artifact %q: file is not executable", entry))
			continue
		}

		results = append(results, Artifact{Platform: p, Mapping: m, Path: path})
	}

	return results, errors.Join(errs...)
}

func ResolveVersion(explicit string) (string, error) {
	if explicit != "" {
		return normalizeVersion(explicit)
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

	return normalizeVersion(version)
}

func normalizeVersion(v string) (string, error) {
	v = strings.TrimPrefix(v, "v")
	if !semverRe.MatchString(v) {
		return "", fmt.Errorf("invalid version %q: must be valid semver (e.g. 1.2.3, 1.2.3-beta.1)", v)
	}
	return v, nil
}
