package config

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jacobarthurs/shipbin/internal/platforms"
)

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
