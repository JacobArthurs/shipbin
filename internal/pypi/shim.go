package pypi

import (
	_ "embed"
	"strings"
)

//go:embed shim.py
var shimTemplate string

func renderShim(binName, pkgName string) ([]byte, error) {
	s := strings.ReplaceAll(shimTemplate, "__BIN_NAME__", binName)
	s = strings.ReplaceAll(s, "__PKG_NAME__", pkgName)
	return []byte(s), nil
}
