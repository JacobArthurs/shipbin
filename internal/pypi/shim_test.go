package pypi

import (
	"strings"
	"testing"
)

func TestRenderShim_ReplacesPlaceholder(t *testing.T) {
	shim, err := renderShim("mytool")
	if err != nil {
		t.Fatalf("renderShim: %v", err)
	}
	s := string(shim)

	if strings.Contains(s, "__BIN_NAME__") {
		t.Error("renderShim did not replace __BIN_NAME__")
	}
	if !strings.Contains(s, "mytool") {
		t.Error("renderShim does not contain the binary name")
	}
}

func TestRenderShim_ContainsRuntimeCalls(t *testing.T) {
	shim, err := renderShim("mytool")
	if err != nil {
		t.Fatalf("renderShim: %v", err)
	}
	s := string(shim)

	if !strings.Contains(s, "os.execv") {
		t.Error("shim missing os.execv (Unix exec path)")
	}
	if !strings.Contains(s, "subprocess.run") {
		t.Error("shim missing subprocess.run (Windows exec path)")
	}
}

func TestRenderShim_DifferentNames(t *testing.T) {
	s1, err := renderShim("foo")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := renderShim("bar")
	if err != nil {
		t.Fatal(err)
	}

	if string(s1) == string(s2) {
		t.Error("shims with different names should differ")
	}
	if !strings.Contains(string(s1), "foo") || strings.Contains(string(s1), "bar") {
		t.Error("s1 should reference 'foo' not 'bar'")
	}
	if !strings.Contains(string(s2), "bar") || strings.Contains(string(s2), "foo") {
		t.Error("s2 should reference 'bar' not 'foo'")
	}
}
