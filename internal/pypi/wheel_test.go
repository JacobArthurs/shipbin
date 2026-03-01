package pypi

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jacobarthurs/shipbin/internal/config"
	"github.com/jacobarthurs/shipbin/internal/platforms"
)

func checkContains(t *testing.T, s string, substrs ...string) {
	t.Helper()
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			t.Errorf("missing expected substring %q in:\n%s", sub, s)
		}
	}
}

func TestToPyPIVersion(t *testing.T) {
	valid := []string{
		"1.0", "1.0.0", "1.2.3",
		"1.0.0a1", "1.0.0b2", "1.0.0rc3",
		"1.0.0.post1", "1.0.0.dev1",
		"2.0.0a1",
	}
	for _, v := range valid {
		t.Run("valid:"+v, func(t *testing.T) {
			got, err := toPyPIVersion(v)
			if err != nil {
				t.Errorf("toPyPIVersion(%q) unexpected error: %v", v, err)
			}
			if got != v {
				t.Errorf("toPyPIVersion(%q) = %q, want %q", v, got, v)
			}
		})
	}

	invalid := []string{
		"1.0.0-alpha",
		"1.0.0-beta.1",
		"v1.0.0",
		"abc",
		"",
	}
	for _, v := range invalid {
		t.Run("invalid:"+v, func(t *testing.T) {
			_, err := toPyPIVersion(v)
			if err == nil {
				t.Errorf("toPyPIVersion(%q) expected error, got nil", v)
			}
		})
	}
}

func TestReadReadme_ContentTypes(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		filename string
		wantType string
	}{
		{"README.md", "text/markdown"},
		{"README.markdown", "text/markdown"},
		{"README.rst", "text/x-rst"},
		{"README.txt", "text/plain"},
		{"README", "text/plain"},
		{"docs.MD", "text/markdown"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			path := filepath.Join(dir, tt.filename)
			if err := os.WriteFile(path, []byte("hello readme"), 0644); err != nil {
				t.Fatal(err)
			}
			content, contentType, err := readReadme(path)
			if err != nil {
				t.Fatalf("readReadme: %v", err)
			}
			if content != "hello readme" {
				t.Errorf("content = %q, want %q", content, "hello readme")
			}
			if contentType != tt.wantType {
				t.Errorf("contentType = %q, want %q", contentType, tt.wantType)
			}
		})
	}
}

func TestReadReadme_Empty(t *testing.T) {
	content, contentType, err := readReadme("")
	if err != nil {
		t.Fatalf("readReadme(\"\") unexpected error: %v", err)
	}
	if content != "" || contentType != "" {
		t.Errorf("expected empty strings, got content=%q type=%q", content, contentType)
	}
}

func TestReadReadme_MissingFile(t *testing.T) {
	_, _, err := readReadme("/does/not/exist/README.md")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestBuildMetadata_AllFields(t *testing.T) {
	meta := buildMetadata("mytool", "1.0.0", "A handy tool", "MIT", "# Readme content", "text/markdown")

	checkContains(t, meta,
		"Metadata-Version: 2.1",
		"Name: mytool",
		"Version: 1.0.0",
		"Summary: A handy tool",
		"License: MIT",
		"Requires-Python: >=3.7",
		"Description-Content-Type: text/markdown",
		"# Readme content",
	)
}

func TestBuildMetadata_MinimalFields(t *testing.T) {
	meta := buildMetadata("mylib", "2.0.0", "", "", "", "")

	checkContains(t, meta,
		"Metadata-Version: 2.1",
		"Name: mylib",
		"Version: 2.0.0",
		"Requires-Python: >=3.7",
	)

	if strings.Contains(meta, "Summary:") {
		t.Error("Summary field should be absent when empty")
	}
	if strings.Contains(meta, "License:") {
		t.Error("License field should be absent when empty")
	}
	if strings.Contains(meta, "Description-Content-Type:") {
		t.Error("Description-Content-Type should be absent when no readme")
	}
}

func TestBuildWheelMeta(t *testing.T) {
	tag := "manylinux_2_17_x86_64.manylinux2014_x86_64"
	meta := buildWheelMeta(tag)

	checkContains(t, meta,
		"Wheel-Version: 1.0",
		"Generator: shipbin",
		"Root-Is-Purelib: false",
		"Tag: py3-none-"+tag,
	)
}

func makeWheelArtifact(t *testing.T, dir, goos, goarch string) config.Artifact {
	t.Helper()
	m, err := platforms.Lookup(goos, goarch)
	if err != nil {
		t.Fatalf("platforms.Lookup: %v", err)
	}
	name := "binary-" + goos + "-" + goarch
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("fake binary data"), 0755); err != nil {
		t.Fatalf("failed to write test binary: %v", err)
	}
	return config.Artifact{
		Platform: platforms.Platform{GOOS: goos, GOARCH: goarch},
		Mapping:  m,
		Path:     path,
	}
}

func TestBuildWheel_ZipStructure(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:    "mytool",
		Version: "1.0.0",
		Summary: "Test tool",
		License: "MIT",
	}

	wf, err := buildWheel(cfg, a)
	if err != nil {
		t.Fatalf("buildWheel: %v", err)
	}

	expectedFilename := "mytool-1.0.0-py3-none-manylinux_2_17_x86_64.manylinux2014_x86_64.whl"
	if wf.filename != expectedFilename {
		t.Errorf("filename = %q, want %q", wf.filename, expectedFilename)
	}
	if wf.version != "1.0.0" {
		t.Errorf("version = %q, want %q", wf.version, "1.0.0")
	}

	zr, err := zip.NewReader(bytes.NewReader(wf.data), int64(len(wf.data)))
	if err != nil {
		t.Fatalf("wheel is not a valid ZIP: %v", err)
	}

	entries := make(map[string]bool, len(zr.File))
	for _, f := range zr.File {
		entries[f.Name] = true
	}

	required := []string{
		"mytool/bin/mytool",
		"mytool/__init__.py",
		"mytool-1.0.0.dist-info/METADATA",
		"mytool-1.0.0.dist-info/WHEEL",
		"mytool-1.0.0.dist-info/entry_points.txt",
		"mytool-1.0.0.dist-info/RECORD",
	}
	for _, name := range required {
		if !entries[name] {
			t.Errorf("wheel missing expected entry: %s", name)
		}
	}
}

func TestBuildWheel_WindowsBinaryHasExeSuffix(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "windows", "amd64")

	cfg := &Config{Name: "mytool", Version: "1.0.0"}

	wf, err := buildWheel(cfg, a)
	if err != nil {
		t.Fatalf("buildWheel: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(wf.data), int64(len(wf.data)))
	if err != nil {
		t.Fatalf("not a valid ZIP: %v", err)
	}

	found := false
	for _, f := range zr.File {
		if f.Name == "mytool/bin/mytool.exe" {
			found = true
			break
		}
	}
	if !found {
		t.Error("windows wheel missing mytool/bin/mytool.exe")
	}
}

func TestBuildWheel_InvalidPEP440Version(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:    "mytool",
		Version: "1.0.0-beta",
	}

	_, err := buildWheel(cfg, a)
	if err == nil {
		t.Fatal("expected error for non-PEP 440 version, got nil")
	}
}

func TestBuildWheel_NameNormalization(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:    "my-tool",
		Version: "1.0.0",
	}

	wf, err := buildWheel(cfg, a)
	if err != nil {
		t.Fatalf("buildWheel: %v", err)
	}

	if strings.Contains(wf.filename, "my-tool") {
		t.Error("wheel filename should use underscores, not hyphens")
	}
	if !strings.Contains(wf.filename, "my_tool") {
		t.Errorf("wheel filename should contain normalized name, got %q", wf.filename)
	}
}

func TestBuildWheel_RecordContainsHashes(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "linux", "amd64")

	cfg := &Config{Name: "mytool", Version: "1.0.0"}

	wf, err := buildWheel(cfg, a)
	if err != nil {
		t.Fatalf("buildWheel: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(wf.data), int64(len(wf.data)))
	if err != nil {
		t.Fatalf("not a valid ZIP: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == "mytool-1.0.0.dist-info/RECORD" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open RECORD: %v", err)
			}
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(rc); err != nil {
				t.Fatalf("failed to read RECORD: %v", err)
			}
			_ = rc.Close()
			record := buf.String()
			if !strings.Contains(record, "sha256=") {
				t.Errorf("RECORD should contain sha256 hashes, got:\n%s", record)
			}
			return
		}
	}
	t.Error("RECORD file not found in wheel")
}

func TestBuildWheel_DotNameNormalization(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "linux", "amd64")

	cfg := &Config{
		Name:    "my.tool",
		Version: "1.0.0",
	}

	wf, err := buildWheel(cfg, a)
	if err != nil {
		t.Fatalf("buildWheel: %v", err)
	}

	if strings.Contains(wf.filename, "my.tool") {
		t.Error("wheel filename should not contain dots in the name segment")
	}
	if !strings.Contains(wf.filename, "my_tool") {
		t.Errorf("wheel filename should use underscores for dots, got %q", wf.filename)
	}

	zr, err := zip.NewReader(bytes.NewReader(wf.data), int64(len(wf.data)))
	if err != nil {
		t.Fatalf("not a valid ZIP: %v", err)
	}
	found := false
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "my_tool/") {
			found = true
			break
		}
	}
	if !found {
		t.Error("ZIP entries should use normalized name prefix my_tool/")
	}
}

func TestBuildWheel_ZipFileModes(t *testing.T) {
	dir := t.TempDir()
	a := makeWheelArtifact(t, dir, "linux", "amd64")

	cfg := &Config{Name: "mytool", Version: "1.0.0"}

	wf, err := buildWheel(cfg, a)
	if err != nil {
		t.Fatalf("buildWheel: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(wf.data), int64(len(wf.data)))
	if err != nil {
		t.Fatalf("not a valid ZIP: %v", err)
	}

	for _, f := range zr.File {
		perm := f.Mode().Perm()
		switch f.Name {
		case "mytool/bin/mytool":
			if perm != 0755 {
				t.Errorf("binary %s has mode %04o, want 0755", f.Name, perm)
			}
		default:
			if perm != 0644 {
				t.Errorf("file %s has mode %04o, want 0644", f.Name, perm)
			}
		}
	}
}
