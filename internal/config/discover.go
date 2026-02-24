package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jacobarthurs/shipbin/internal/platforms"
)

type BinaryPath struct {
	Mapping platforms.Mapping
	Path    string
}

func goReleaserPath(dist, binary string, m platforms.Mapping) string {
	dir := filepath.Join(dist, fmt.Sprintf("%s_%s_%s", binary, m.Platform.GOOS, m.GoReleaserDirSuffix))

	binaryName := binary
	if m.Platform.GOOS == "windows" {
		binaryName += ".exe"
	}

	return filepath.Join(dir, binaryName)
}

func DiscoverBinaries(cfg *Config) ([]BinaryPath, error) {
	var missing []string
	var results []BinaryPath

	for _, m := range cfg.Platforms {
		path := goReleaserPath(cfg.Dist, cfg.Binary, m)

		info, err := os.Stat(path)
		if err != nil {
			missing = append(missing, fmt.Sprintf("  %s/%s → %s", m.Platform.GOOS, m.Platform.GOARCH, path))
			continue
		}
		if info.IsDir() {
			return nil, fmt.Errorf("expected a file but found a directory: %s", path)
		}

		results = append(results, BinaryPath{
			Mapping: m,
			Path:    path,
		})
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"missing binaries for %d platform(s):\n%s\n\nensure GoReleaser has run and --dist points to the correct directory",
			len(missing),
			strings.Join(missing, "\n"),
		)
	}

	return results, nil
}
