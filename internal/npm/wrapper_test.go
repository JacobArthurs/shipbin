package npm

import (
	"strings"
	"testing"
)

func TestWrapperScript_ReplacesPlaceholders(t *testing.T) {
	script := wrapperScript("mytool", "myorg")

	if strings.Contains(script, "__BIN_NAME__") {
		t.Error("wrapperScript did not replace __BIN_NAME__")
	}
	if strings.Contains(script, "__ORG_NAME__") {
		t.Error("wrapperScript did not replace __ORG_NAME__")
	}
}

func TestWrapperScript_ContainsPackageReferences(t *testing.T) {
	script := wrapperScript("cli", "acme")

	if !strings.Contains(script, `const BIN_NAME = "cli"`) {
		t.Error(`script does not contain: const BIN_NAME = "cli"`)
	}
	if !strings.Contains(script, `const ORG_NAME = "acme"`) {
		t.Error(`script does not contain: const ORG_NAME = "acme"`)
	}
}

func TestWrapperScript_ContainsRuntimeCalls(t *testing.T) {
	script := wrapperScript("mytool", "myorg")

	if !strings.Contains(script, "require.resolve") {
		t.Error("script does not contain require.resolve")
	}
	if !strings.Contains(script, "execFileSync") {
		t.Error("script does not contain execFileSync")
	}
}

func TestWrapperScript_DifferentInputs(t *testing.T) {
	s1 := wrapperScript("foo", "bar")
	s2 := wrapperScript("baz", "qux")

	if s1 == s2 {
		t.Error("scripts with different inputs should differ")
	}
	if !strings.Contains(s1, "foo") || strings.Contains(s1, "baz") {
		t.Error("s1 should contain 'foo' but not 'baz'")
	}
	if !strings.Contains(s2, "baz") || strings.Contains(s2, "foo") {
		t.Error("s2 should contain 'baz' but not 'foo'")
	}
}
