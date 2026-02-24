package platforms

import (
	"cmp"
	"fmt"
	"slices"
)

type Platform struct {
	GOOS   string
	GOARCH string
}

type NpmMapping struct {
	OS            string
	CPU           string
	PackageSuffix string
}

type PyPIMapping struct {
	WheelTag string
}

type Mapping struct {
	Npm  NpmMapping
	PyPI PyPIMapping
}

var table = map[Platform]Mapping{
	{GOOS: "linux", GOARCH: "amd64"}: {
		Npm:  NpmMapping{OS: "linux", CPU: "x64", PackageSuffix: "linux-x64"},
		PyPI: PyPIMapping{WheelTag: "manylinux_2_17_x86_64.manylinux2014_x86_64"},
	},
	{GOOS: "linux", GOARCH: "arm64"}: {
		Npm:  NpmMapping{OS: "linux", CPU: "arm64", PackageSuffix: "linux-arm64"},
		PyPI: PyPIMapping{WheelTag: "manylinux_2_17_aarch64.manylinux2014_aarch64"},
	},
	{GOOS: "darwin", GOARCH: "amd64"}: {
		Npm:  NpmMapping{OS: "darwin", CPU: "x64", PackageSuffix: "darwin-x64"},
		PyPI: PyPIMapping{WheelTag: "macosx_10_12_x86_64"},
	},
	{GOOS: "darwin", GOARCH: "arm64"}: {
		Npm:  NpmMapping{OS: "darwin", CPU: "arm64", PackageSuffix: "darwin-arm64"},
		PyPI: PyPIMapping{WheelTag: "macosx_11_0_arm64"},
	},
	{GOOS: "windows", GOARCH: "amd64"}: {
		Npm:  NpmMapping{OS: "win32", CPU: "x64", PackageSuffix: "win32-x64"},
		PyPI: PyPIMapping{WheelTag: "win_amd64"},
	},
	{GOOS: "windows", GOARCH: "arm64"}: {
		Npm:  NpmMapping{OS: "win32", CPU: "arm64", PackageSuffix: "win32-arm64"},
		PyPI: PyPIMapping{WheelTag: "win_arm64"},
	},
}

func Lookup(goos, goarch string) (Mapping, error) {
	m, ok := table[Platform{GOOS: goos, GOARCH: goarch}]
	if !ok {
		return Mapping{}, fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}
	return m, nil
}

func All() []Platform {
	result := make([]Platform, 0, len(table))
	for p := range table {
		result = append(result, p)
	}
	slices.SortFunc(result, func(a, b Platform) int {
		if a.GOOS != b.GOOS {
			return cmp.Compare(a.GOOS, b.GOOS)
		}
		return cmp.Compare(a.GOARCH, b.GOARCH)
	})
	return result
}
