package pypi

import (
	_ "embed"
	"strings"
)

//go:embed shim.py
var shimTemplate string

func renderShim(name string) ([]byte, error) {
	return []byte(strings.ReplaceAll(shimTemplate, "__BIN_NAME__", name)), nil
}
