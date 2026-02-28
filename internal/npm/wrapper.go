package npm

import (
	_ "embed"
	"strings"
)

//go:embed wrapper.js
var wrapperJS string

func wrapperScript(name, org string) string {
	s := strings.ReplaceAll(wrapperJS, "__BIN_NAME__", name)
	return strings.ReplaceAll(s, "__ORG_NAME__", org)
}
