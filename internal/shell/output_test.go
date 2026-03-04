package shell

import (
	"bytes"
	"testing"
)

func TestCdPrefix(t *testing.T) {
	if CdPrefix != "__WT_CD__:" {
		t.Errorf("CdPrefix = %q, want %q", CdPrefix, "__WT_CD__:")
	}
}

func TestPrintCdPath(t *testing.T) {
	var buf bytes.Buffer
	PrintCdPath(&buf, "/home/user/repo")

	got := buf.String()
	want := "__WT_CD__:/home/user/repo"
	if got != want {
		t.Errorf("PrintCdPath output = %q, want %q", got, want)
	}
}

func TestPrintDisplay(t *testing.T) {
	var buf bytes.Buffer
	PrintDisplay(&buf, "hello world")

	got := buf.String()
	if got != "hello world" {
		t.Errorf("PrintDisplay output = %q, want %q", got, "hello world")
	}
}
