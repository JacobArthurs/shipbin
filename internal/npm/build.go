package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type packageJSON struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	License      string            `json:"license"`
	OS           []string          `json:"os,omitempty"`
	CPU          []string          `json:"cpu,omitempty"`
	Files        []string          `json:"files"`
	Bin          map[string]string `json:"bin,omitempty"`
	OptionalDeps map[string]string `json:"optionalDependencies,omitempty"`
}

type builtPackage struct {
	dir  string
	name string
}

func buildPlatformPackages(cfg *Config) ([]builtPackage, func(), error) {
	var packages []builtPackage
	var dirs []string

	cleanup := func() {
		for _, d := range dirs {
			os.RemoveAll(d)
		}
	}

	for _, a := range cfg.Artifacts {
		pkgName := fmt.Sprintf("@%s/%s-%s", cfg.Org, cfg.Name, a.Mapping.Npm.PackageSuffix)

		dir, err := os.MkdirTemp("", "shipbin-npm-*")
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to create temp dir for %s: %w", pkgName, err)
		}
		dirs = append(dirs, dir)

		binDir := filepath.Join(dir, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to create bin dir for %s: %w", pkgName, err)
		}

		binaryName := cfg.Name
		if a.Platform.GOOS == "windows" {
			binaryName += ".exe"
		}
		destBinary := filepath.Join(binDir, binaryName)
		if err := copyFile(a.Path, destBinary, 0755); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to copy binary for %s: %w", pkgName, err)
		}

		pkg := packageJSON{
			Name:        pkgName,
			Version:     cfg.Version,
			Description: fmt.Sprintf("%s binary for %s", cfg.Name, a.Mapping.Npm.PackageSuffix),
			License:     cfg.License,
			OS:          []string{a.Mapping.Npm.OS},
			CPU:         []string{a.Mapping.Npm.CPU},
			Files:       []string{"bin"},
		}
		if err := writeJSON(filepath.Join(dir, "package.json"), pkg); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to write package.json for %s: %w", pkgName, err)
		}

		packages = append(packages, builtPackage{dir: dir, name: pkgName})
	}

	return packages, cleanup, nil
}

func buildRootPackage(cfg *Config) (builtPackage, func(), error) {
	rootName := cfg.Name

	dir, err := os.MkdirTemp("", "shipbin-npm-root-*")
	if err != nil {
		return builtPackage{}, nil, fmt.Errorf("failed to create temp dir for root package: %w", err)
	}
	cleanup := func() { os.RemoveAll(dir) }

	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		cleanup()
		return builtPackage{}, nil, fmt.Errorf("failed to create bin dir for root package: %w", err)
	}
	wrapperPath := filepath.Join(binDir, cfg.Name)
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript(cfg.Name)), 0755); err != nil {
		cleanup()
		return builtPackage{}, nil, fmt.Errorf("failed to write wrapper script: %w", err)
	}

	optDeps := make(map[string]string, len(cfg.Artifacts))
	for _, a := range cfg.Artifacts {
		pkgName := fmt.Sprintf("@%s/%s-%s", cfg.Org, cfg.Name, a.Mapping.Npm.PackageSuffix)
		optDeps[pkgName] = cfg.Version
	}

	pkg := packageJSON{
		Name:         rootName,
		Version:      cfg.Version,
		Description:  fmt.Sprintf("Install %s — native binary distributed via npm", cfg.Name),
		License:      cfg.License,
		Files:        []string{"bin"},
		Bin:          map[string]string{cfg.Name: fmt.Sprintf("bin/%s", cfg.Name)},
		OptionalDeps: optDeps,
	}
	if err := writeJSON(filepath.Join(dir, "package.json"), pkg); err != nil {
		cleanup()
		return builtPackage{}, nil, fmt.Errorf("failed to write root package.json: %w", err)
	}

	if cfg.Readme != "" {
		if err := copyFile(cfg.Readme, filepath.Join(dir, "README.md"), 0644); err != nil {
			cleanup()
			return builtPackage{}, nil, fmt.Errorf("failed to copy README: %w", err)
		}
	}

	return builtPackage{dir: dir, name: rootName}, cleanup, nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
