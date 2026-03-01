package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func makeExe(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	mode := os.FileMode(0755)
	if runtime.GOOS == "windows" {
		mode = 0644
	}
	if err := os.WriteFile(path, []byte("fake binary"), mode); err != nil {
		t.Fatalf("failed to create test binary: %v", err)
	}
	return path
}

func TestParseArtifacts_Valid(t *testing.T) {
	dir := t.TempDir()
	p := makeExe(t, dir, "mytool")

	artifacts, err := ParseArtifacts([]string{"linux/amd64:" + p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(artifacts))
	}
	if artifacts[0].Platform.GOOS != "linux" || artifacts[0].Platform.GOARCH != "amd64" {
		t.Errorf("unexpected platform: %+v", artifacts[0].Platform)
	}
	if artifacts[0].Path != p {
		t.Errorf("path = %q, want %q", artifacts[0].Path, p)
	}
	if artifacts[0].Mapping.Npm.PackageSuffix != "linux-x64" {
		t.Errorf("npm PackageSuffix = %q, want %q", artifacts[0].Mapping.Npm.PackageSuffix, "linux-x64")
	}
}

func TestParseArtifacts_MultipleValid(t *testing.T) {
	dir := t.TempDir()
	p1 := makeExe(t, dir, "bin-linux")
	p2 := makeExe(t, dir, "bin-darwin")

	artifacts, err := ParseArtifacts([]string{
		"linux/amd64:" + p1,
		"darwin/arm64:" + p2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(artifacts))
	}
}

func TestParseArtifacts_MissingColon(t *testing.T) {
	_, err := ParseArtifacts([]string{"linux/amd64/path/to/bin"})
	if err == nil {
		t.Fatal("expected error for missing colon, got nil")
	}
}

func TestParseArtifacts_MissingSlash(t *testing.T) {
	_, err := ParseArtifacts([]string{"linuxamd64:/some/path"})
	if err == nil {
		t.Fatal("expected error for missing slash in platform, got nil")
	}
}

func TestParseArtifacts_UnsupportedPlatform(t *testing.T) {
	_, err := ParseArtifacts([]string{"freebsd/amd64:/some/path"})
	if err == nil {
		t.Fatal("expected error for unsupported platform, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported platform") {
		t.Errorf("error should mention unsupported platform, got: %v", err)
	}
}

func TestParseArtifacts_FileNotFound(t *testing.T) {
	_, err := ParseArtifacts([]string{"linux/amd64:/does/not/exist/binary"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseArtifacts_IsDirectory(t *testing.T) {
	dir := t.TempDir()
	_, err := ParseArtifacts([]string{"linux/amd64:" + dir})
	if err == nil {
		t.Fatal("expected error when path is a directory, got nil")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("error should mention directory, got: %v", err)
	}
}

func TestParseArtifacts_NotExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable bit check is not enforced on Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "notexe")
	if err := os.WriteFile(path, []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseArtifacts([]string{"linux/amd64:" + path})
	if err == nil {
		t.Fatal("expected error for non-executable file, got nil")
	}
	if !strings.Contains(err.Error(), "not executable") {
		t.Errorf("error should mention not executable, got: %v", err)
	}
}

func TestParseArtifacts_WindowsTargetSkipsExeCheck(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable bit check is not enforced on Windows hosts")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "tool.exe")
	if err := os.WriteFile(path, []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseArtifacts([]string{"windows/amd64:" + path})
	if err != nil {
		t.Fatalf("expected no error for windows target with no exe bit, got: %v", err)
	}
}

func TestParseArtifacts_DuplicatePlatform(t *testing.T) {
	dir := t.TempDir()
	p := makeExe(t, dir, "bin")
	_, err := ParseArtifacts([]string{
		"linux/amd64:" + p,
		"linux/amd64:" + p,
	})
	if err == nil {
		t.Fatal("expected error for duplicate platform, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention duplicate, got: %v", err)
	}
}

func TestParseArtifacts_CollectsMultipleErrors(t *testing.T) {
	_, err := ParseArtifacts([]string{
		"bad-entry-one",
		"bad-entry-two",
	})
	if err == nil {
		t.Fatal("expected errors, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bad-entry-one") {
		t.Errorf("error missing first entry, got: %s", msg)
	}
	if !strings.Contains(msg, "bad-entry-two") {
		t.Errorf("error missing second entry, got: %s", msg)
	}
}

func TestParseArtifacts_EmptyInput(t *testing.T) {
	artifacts, err := ParseArtifacts([]string{})
	if err != nil {
		t.Fatalf("unexpected error for empty input: %v", err)
	}
	if len(artifacts) != 0 {
		t.Errorf("expected 0 artifacts, got %d", len(artifacts))
	}
}

func TestResolveVersion_ExplicitValid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.2.3", "1.2.3"},
		{"v1.2.3", "1.2.3"},
		{"1.0.0-beta.1", "1.0.0-beta.1"},
		{"v2.0.0-rc.1", "2.0.0-rc.1"},
		{"0.0.1", "0.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ResolveVersion(tt.input)
			if err != nil {
				t.Fatalf("ResolveVersion(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ResolveVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveVersion_ExplicitInvalid(t *testing.T) {
	cases := []string{
		"not-a-version",
		"1.0",
		"1.0.0.0",
		"abc",
		"v1.0",
		"1.2.3.4",
	}
	for _, v := range cases {
		t.Run(v, func(t *testing.T) {
			_, err := ResolveVersion(v)
			if err == nil {
				t.Errorf("ResolveVersion(%q) expected error, got nil", v)
			}
		})
	}
}
