package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("valid yaml", func(t *testing.T) {
		tmp := t.TempDir()
		yaml := `services:
  web:
    cmd: "npm start"
    dir: "frontend"
    color: "green"
  api:
    cmd: "go run ."
    dir: "backend"
    env:
      PORT: "8080"
`
		os.WriteFile(filepath.Join(tmp, FileName), []byte(yaml), 0644)

		cfg, err := Load(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Services) != 2 {
			t.Fatalf("got %d services, want 2", len(cfg.Services))
		}

		web := cfg.Services["web"]
		if web.Cmd != "npm start" {
			t.Errorf("web.Cmd = %q, want %q", web.Cmd, "npm start")
		}
		if web.Dir != "frontend" {
			t.Errorf("web.Dir = %q, want %q", web.Dir, "frontend")
		}
		if web.Color != "green" {
			t.Errorf("web.Color = %q, want %q", web.Color, "green")
		}

		api := cfg.Services["api"]
		if api.Env["PORT"] != "8080" {
			t.Errorf("api.Env[PORT] = %q, want %q", api.Env["PORT"], "8080")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		tmp := t.TempDir()
		_, err := Load(tmp)
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("malformed yaml", func(t *testing.T) {
		tmp := t.TempDir()
		os.WriteFile(filepath.Join(tmp, FileName), []byte("{{invalid"), 0644)

		_, err := Load(tmp)
		if err == nil {
			t.Fatal("expected error for malformed yaml")
		}
	})

	t.Run("no services", func(t *testing.T) {
		tmp := t.TempDir()
		os.WriteFile(filepath.Join(tmp, FileName), []byte("services:\n"), 0644)

		_, err := Load(tmp)
		if err == nil {
			t.Fatal("expected error for empty services")
		}
	})

	t.Run("service without cmd", func(t *testing.T) {
		tmp := t.TempDir()
		yaml := `services:
  web:
    dir: "frontend"
`
		os.WriteFile(filepath.Join(tmp, FileName), []byte(yaml), 0644)

		_, err := Load(tmp)
		if err == nil {
			t.Fatal("expected error for service without cmd")
		}
	})

	t.Run("default dir is dot", func(t *testing.T) {
		tmp := t.TempDir()
		yaml := `services:
  web:
    cmd: "npm start"
`
		os.WriteFile(filepath.Join(tmp, FileName), []byte(yaml), 0644)

		cfg, err := Load(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Services["web"].Dir != "." {
			t.Errorf("Dir = %q, want %q", cfg.Services["web"].Dir, ".")
		}
	})
}

func TestSave(t *testing.T) {
	tmp := t.TempDir()

	original := &Config{
		Services: map[string]Service{
			"web": {Cmd: "npm start", Dir: "frontend", Color: "cyan"},
			"api": {Cmd: "go run .", Dir: ".", Env: map[string]string{"PORT": "3000"}},
		},
	}

	if err := Save(tmp, original); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmp, FileName)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	// Round-trip: load it back
	loaded, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load after Save error: %v", err)
	}

	if loaded.Services["web"].Cmd != "npm start" {
		t.Errorf("round-trip web.Cmd = %q, want %q", loaded.Services["web"].Cmd, "npm start")
	}
	if loaded.Services["api"].Env["PORT"] != "3000" {
		t.Errorf("round-trip api.Env[PORT] = %q, want %q", loaded.Services["api"].Env["PORT"], "3000")
	}
}

func TestLoadOrEmpty(t *testing.T) {
	tmp := t.TempDir()

	cfg := LoadOrEmpty(tmp)
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Services == nil {
		t.Fatal("expected non-nil Services map")
	}
	if len(cfg.Services) != 0 {
		t.Errorf("got %d services, want 0", len(cfg.Services))
	}
}
