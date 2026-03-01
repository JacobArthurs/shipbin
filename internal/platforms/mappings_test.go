package platforms

import (
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		goos          string
		goarch        string
		wantNpmSuffix string
		wantNpmOS     string
		wantNpmCPU    string
		wantWheelTag  string
	}{
		{"linux", "amd64", "linux-x64", "linux", "x64", "manylinux_2_17_x86_64.manylinux2014_x86_64"},
		{"linux", "arm64", "linux-arm64", "linux", "arm64", "manylinux_2_17_aarch64.manylinux2014_aarch64"},
		{"darwin", "amd64", "darwin-x64", "darwin", "x64", "macosx_10_12_x86_64"},
		{"darwin", "arm64", "darwin-arm64", "darwin", "arm64", "macosx_11_0_arm64"},
		{"windows", "amd64", "win32-x64", "win32", "x64", "win_amd64"},
		{"windows", "arm64", "win32-arm64", "win32", "arm64", "win_arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			m, err := Lookup(tt.goos, tt.goarch)
			if err != nil {
				t.Fatalf("Lookup(%q, %q) unexpected error: %v", tt.goos, tt.goarch, err)
			}
			if m.Npm.PackageSuffix != tt.wantNpmSuffix {
				t.Errorf("Npm.PackageSuffix = %q, want %q", m.Npm.PackageSuffix, tt.wantNpmSuffix)
			}
			if m.Npm.OS != tt.wantNpmOS {
				t.Errorf("Npm.OS = %q, want %q", m.Npm.OS, tt.wantNpmOS)
			}
			if m.Npm.CPU != tt.wantNpmCPU {
				t.Errorf("Npm.CPU = %q, want %q", m.Npm.CPU, tt.wantNpmCPU)
			}
			if m.PyPI.WheelTag != tt.wantWheelTag {
				t.Errorf("PyPI.WheelTag = %q, want %q", m.PyPI.WheelTag, tt.wantWheelTag)
			}
		})
	}
}

func TestLookup_Unsupported(t *testing.T) {
	cases := [][2]string{
		{"freebsd", "amd64"},
		{"linux", "386"},
		{"", ""},
		{"windows", "x86"},
		{"darwin", "386"},
	}
	for _, c := range cases {
		_, err := Lookup(c[0], c[1])
		if err == nil {
			t.Errorf("Lookup(%q, %q) expected error, got nil", c[0], c[1])
		}
	}
}

func TestAll(t *testing.T) {
	all := All()

	if len(all) != 6 {
		t.Fatalf("All() returned %d platforms, want 6", len(all))
	}

	for i := 1; i < len(all); i++ {
		a, b := all[i-1], all[i]
		if a.GOOS > b.GOOS || (a.GOOS == b.GOOS && a.GOARCH > b.GOARCH) {
			t.Errorf("All() not sorted at index %d: %s/%s comes before %s/%s",
				i, b.GOOS, b.GOARCH, a.GOOS, a.GOARCH)
		}
	}

	for _, p := range all {
		if _, err := Lookup(p.GOOS, p.GOARCH); err != nil {
			t.Errorf("All() returned unresolvable platform %s/%s", p.GOOS, p.GOARCH)
		}
	}
}
