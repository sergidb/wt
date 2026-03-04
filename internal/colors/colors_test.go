package colors

import (
	"strings"
	"testing"
)

func TestAllNamesHaveANSICodes(t *testing.T) {
	for _, name := range Names {
		code, ok := ANSICodes[name]
		if !ok {
			t.Errorf("color %q in Names but not in ANSICodes", name)
		}
		if !strings.HasPrefix(code, "\033[") {
			t.Errorf("ANSICodes[%q] = %q, expected ANSI escape prefix", name, code)
		}
	}
}

func TestANSIReset(t *testing.T) {
	if ANSIReset != "\033[0m" {
		t.Errorf("ANSIReset = %q, want %q", ANSIReset, "\033[0m")
	}
}

func TestNamesNotEmpty(t *testing.T) {
	if len(Names) == 0 {
		t.Error("Names should not be empty")
	}
}

func TestANSICodesNotEmpty(t *testing.T) {
	if len(ANSICodes) == 0 {
		t.Error("ANSICodes should not be empty")
	}
}
