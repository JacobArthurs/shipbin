package npm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/platforms"
)

func makeTestBinary(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	mode := os.FileMode(0755)
	if runtime.GOOS == "windows" {
		mode = 0644
	}
	if err := os.WriteFile(path, []byte("fake binary content"), mode); err != nil {
		t.Fatalf("failed to create test binary: %v", err)
	}
	return path
}

func makeArtifact(t *testing.T, dir, goos, goarch string) config.Artifact {
	t.Helper()
	m, err := platforms.Lookup(goos, goarch)
	if err != nil {
		t.Fatalf("platforms.Lookup(%q, %q): %v", goos, goarch, err)
	}
	binName := "binary-" + goos + "-" + goarch
	if goos == "windows" {
		binName += ".exe"
	}
	return config.Artifact{
		Platform: platforms.Platform{GOOS: goos, GOARCH: goarch},
		Mapping:  m,
		Path:     makeTestBinary(t, dir, binName),
	}
}

func TestBuildPlatformPackages_SingleArtifact(t *testing.T) {
	dir := t.TempDir()
	a := makeArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:      "mytool",
		Version:   "1.0.0",
		License:   "MIT",
		Org:       "myorg",
		Artifacts: []config.Artifact{a},
	}

	pkgs, cleanup, err := buildPlatformPackages(cfg)
	if err != nil {
		t.Fatalf("buildPlatformPackages: %v", err)
	}
	defer cleanup()

	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	pkg := pkgs[0]
	if pkg.name != "@myorg/mytool-linux-x64" {
		t.Errorf("name = %q, want %q", pkg.name, "@myorg/mytool-linux-x64")
	}

	binPath := filepath.Join(pkg.dir, "bin", "mytool")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("binary not found at %s: %v", binPath, err)
	}

	data, err := os.ReadFile(filepath.Join(pkg.dir, "package.json"))
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	var pj packageJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		t.Fatalf("package.json is not valid JSON: %v", err)
	}
	if pj.Name != "@myorg/mytool-linux-x64" {
		t.Errorf("package.json name = %q", pj.Name)
	}
	if pj.Version != "1.0.0" {
		t.Errorf("package.json version = %q", pj.Version)
	}
	if len(pj.OS) != 1 || pj.OS[0] != "linux" {
		t.Errorf("package.json os = %v, want [linux]", pj.OS)
	}
	if len(pj.CPU) != 1 || pj.CPU[0] != "x64" {
		t.Errorf("package.json cpu = %v, want [x64]", pj.CPU)
	}
	if pj.License != "MIT" {
		t.Errorf("package.json license = %q, want MIT", pj.License)
	}
}

func TestBuildPlatformPackages_MultipleArtifacts(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Name:    "mytool",
		Version: "2.0.0",
		Org:     "myorg",
		Artifacts: []config.Artifact{
			makeArtifact(t, dir, "linux", "amd64"),
			makeArtifact(t, dir, "darwin", "arm64"),
			makeArtifact(t, dir, "windows", "amd64"),
		},
	}

	pkgs, cleanup, err := buildPlatformPackages(cfg)
	if err != nil {
		t.Fatalf("buildPlatformPackages: %v", err)
	}
	defer cleanup()

	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
}

func TestBuildPlatformPackages_WindowsBinaryHasExeSuffix(t *testing.T) {
	dir := t.TempDir()
	a := makeArtifact(t, dir, "windows", "amd64")

	cfg := &Config{
		Name:      "mytool",
		Version:   "1.0.0",
		Org:       "myorg",
		Artifacts: []config.Artifact{a},
	}

	pkgs, cleanup, err := buildPlatformPackages(cfg)
	if err != nil {
		t.Fatalf("buildPlatformPackages: %v", err)
	}
	defer cleanup()

	binPath := filepath.Join(pkgs[0].dir, "bin", "mytool.exe")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("windows binary not found at %s: %v", binPath, err)
	}
}

func TestBuildPlatformPackages_CleanupRemovesDirs(t *testing.T) {
	dir := t.TempDir()
	a := makeArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:      "mytool",
		Version:   "1.0.0",
		Org:       "myorg",
		Artifacts: []config.Artifact{a},
	}

	pkgs, cleanup, err := buildPlatformPackages(cfg)
	if err != nil {
		t.Fatalf("buildPlatformPackages: %v", err)
	}

	pkgDir := pkgs[0].dir
	if _, err := os.Stat(pkgDir); err != nil {
		t.Fatalf("package dir should exist before cleanup: %v", err)
	}

	cleanup()

	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Error("package dir should be removed after cleanup")
	}
}

func TestBuildRootPackage(t *testing.T) {
	dir := t.TempDir()
	a := makeArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:      "mytool",
		Version:   "1.2.3",
		Summary:   "A great tool",
		License:   "MIT",
		Org:       "myorg",
		Artifacts: []config.Artifact{a},
	}

	pkg, cleanup, err := buildRootPackage(cfg)
	if err != nil {
		t.Fatalf("buildRootPackage: %v", err)
	}
	defer cleanup()

	if pkg.name != "mytool" {
		t.Errorf("root package name = %q, want %q", pkg.name, "mytool")
	}

	wrapperPath := filepath.Join(pkg.dir, "bin", "mytool")
	if _, err := os.Stat(wrapperPath); err != nil {
		t.Errorf("wrapper script not found at %s: %v", wrapperPath, err)
	}

	data, err := os.ReadFile(filepath.Join(pkg.dir, "package.json"))
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	var pj packageJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		t.Fatalf("package.json invalid JSON: %v", err)
	}
	if pj.Name != "mytool" {
		t.Errorf("name = %q, want %q", pj.Name, "mytool")
	}
	if pj.Version != "1.2.3" {
		t.Errorf("version = %q, want %q", pj.Version, "1.2.3")
	}
	if pj.Description != "A great tool" {
		t.Errorf("description = %q", pj.Description)
	}
	if _, ok := pj.OptionalDeps["@myorg/mytool-linux-x64"]; !ok {
		t.Errorf("optionalDependencies missing @myorg/mytool-linux-x64, got: %v", pj.OptionalDeps)
	}
	if pj.OptionalDeps["@myorg/mytool-linux-x64"] != "1.2.3" {
		t.Errorf("optionalDependency version = %q, want %q",
			pj.OptionalDeps["@myorg/mytool-linux-x64"], "1.2.3")
	}
	if binScript, ok := pj.Bin["mytool"]; !ok || binScript != "bin/mytool" {
		t.Errorf("bin[mytool] = %q, want %q", binScript, "bin/mytool")
	}
}

func TestBuildRootPackage_WithReadme(t *testing.T) {
	dir := t.TempDir()
	a := makeArtifact(t, dir, "linux", "amd64")
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# My Tool"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Name:      "mytool",
		Version:   "1.0.0",
		Org:       "myorg",
		Readme:    readmePath,
		Artifacts: []config.Artifact{a},
	}

	pkg, cleanup, err := buildRootPackage(cfg)
	if err != nil {
		t.Fatalf("buildRootPackage: %v", err)
	}
	defer cleanup()

	if _, err := os.Stat(filepath.Join(pkg.dir, "README.md")); err != nil {
		t.Errorf("README.md was not copied: %v", err)
	}
}

func TestBuildRootPackage_OptionalDepsForAllPlatforms(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Name:    "mytool",
		Version: "1.0.0",
		Org:     "acme",
		Artifacts: []config.Artifact{
			makeArtifact(t, dir, "linux", "amd64"),
			makeArtifact(t, dir, "darwin", "arm64"),
			makeArtifact(t, dir, "windows", "amd64"),
		},
	}

	pkg, cleanup, err := buildRootPackage(cfg)
	if err != nil {
		t.Fatalf("buildRootPackage: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(pkg.dir, "package.json"))
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	var pj packageJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(pj.OptionalDeps) != 3 {
		t.Errorf("expected 3 optionalDependencies, got %d: %v", len(pj.OptionalDeps), pj.OptionalDeps)
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	content := []byte("hello world")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, dst, 0755); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(dst)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&0755 != 0755 {
			t.Errorf("mode = %o, want 0755", info.Mode())
		}
	}
}

func TestCopyFile_MissingSrc(t *testing.T) {
	dir := t.TempDir()
	err := copyFile(filepath.Join(dir, "nonexistent"), filepath.Join(dir, "dst"), 0644)
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestWriteJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	v := map[string]string{"key": "value"}
	if err := writeJSON(path, v); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("got[key] = %q, want %q", got["key"], "value")
	}
	if data[len(data)-1] != '\n' {
		t.Error("writeJSON output does not end with a newline")
	}
}
