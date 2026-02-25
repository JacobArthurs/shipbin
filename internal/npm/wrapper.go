package npm

import (
	_ "embed"
	"strings"
)

//go:embed wrapper.js
var wrapperJS string

func wrapperScript(name string) string {
	return strings.ReplaceAll(wrapperJS, "__BIN_NAME__", name)
}
