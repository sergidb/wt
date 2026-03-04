package runner

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/sergidb/wt/internal/colors"
	"github.com/sergidb/wt/internal/config"
)

func TestFormatPrefix(t *testing.T) {
	tests := []struct {
		name   string
		maxLen int
		color  string
		want   string // substring that must appear
	}{
		{
			name:   "web",
			maxLen: 5,
			color:  "green",
			want:   "[web  ]",
		},
		{
			name:   "api-server",
			maxLen: 10,
			color:  "cyan",
			want:   "[api-server]",
		},
		{
			name:   "svc",
			maxLen: 3,
			color:  "invalid",
			want:   "[svc]", // falls back to white
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrefix(tt.name, tt.maxLen, tt.color)
			if !strings.Contains(got, tt.want) {
				t.Errorf("formatPrefix(%q, %d, %q) = %q, want to contain %q", tt.name, tt.maxLen, tt.color, got, tt.want)
			}

			// Should contain ANSI codes
			if !strings.Contains(got, "\033[") {
				t.Error("expected ANSI color codes in output")
			}

			// Should end with reset
			if !strings.HasSuffix(got, colors.ANSIReset) {
				t.Error("expected color reset at end")
			}
		})
	}
}

func TestFormatPrefixAlignment(t *testing.T) {
	p1 := formatPrefix("a", 5, "green")
	p2 := formatPrefix("abcde", 5, "green")

	// Both should have the same length (same ANSI overhead + same maxLen)
	if len(p1) != len(p2) {
		t.Errorf("prefixes should have same length: len(%q)=%d, len(%q)=%d", p1, len(p1), p2, len(p2))
	}
}

func TestFormatPrefixInvalidColor(t *testing.T) {
	got := formatPrefix("svc", 3, "nonexistent")
	// Should fall back to white
	if !strings.Contains(got, colors.ANSICodes["white"]) {
		t.Error("expected white fallback for invalid color")
	}
}

func TestRunNoServices(t *testing.T) {
	var buf bytes.Buffer
	err := Run(map[string]config.Service{}, "/tmp", &buf)
	if err == nil {
		t.Fatal("expected error for empty services")
	}
	if !strings.Contains(err.Error(), "no services") {
		t.Errorf("error = %q, want to contain 'no services'", err.Error())
	}
}

func TestRunSimpleServices(t *testing.T) {
	var buf bytes.Buffer
	services := map[string]config.Service{
		"echo1": {Cmd: "echo hello", Dir: "."},
		"echo2": {Cmd: "echo world", Dir: ".", Color: "cyan"},
	}

	err := Run(services, "/tmp", &buf)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "hello") {
		t.Error("expected 'hello' in output")
	}
	if !strings.Contains(output, "world") {
		t.Error("expected 'world' in output")
	}
}

func TestRunServiceWithEnv(t *testing.T) {
	var buf bytes.Buffer
	services := map[string]config.Service{
		"env": {
			Cmd:   "echo $MY_TEST_VAR",
			Dir:   ".",
			Color: "green",
			Env:   map[string]string{"MY_TEST_VAR": "it_works"},
		},
	}

	err := Run(services, "/tmp", &buf)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(buf.String(), "it_works") {
		t.Errorf("expected env var in output, got: %s", buf.String())
	}
}

func TestRunServiceFailure(t *testing.T) {
	var buf bytes.Buffer
	services := map[string]config.Service{
		"fail": {Cmd: "exit 1", Dir: "."},
	}

	err := Run(services, "/tmp", &buf)
	if err == nil {
		t.Fatal("expected error for failing service")
	}
}

func TestStreamLines(t *testing.T) {
	var buf bytes.Buffer
	r := strings.NewReader("line1\nline2\nline3\n")

	var wg sync.WaitGroup
	wg.Add(1)
	go streamLines(&wg, r, "[test]", &buf)
	wg.Wait()

	output := buf.String()
	for _, want := range []string{"[test] line1", "[test] line2", "[test] line3"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got: %s", want, output)
		}
	}
}
