package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = ".wt.yaml"

// Service defines a single service to run.
type Service struct {
	Cmd   string            `yaml:"cmd"`
	Dir   string            `yaml:"dir"`
	Env   map[string]string `yaml:"env"`
	Color string            `yaml:"color"`
}

// Config is the root configuration from .wt.yaml.
type Config struct {
	WorktreesDir string             `yaml:"worktrees_dir,omitempty"`
	Services     map[string]Service `yaml:"services"`
}

// GetWorktreesDir returns the resolved worktrees directory path.
// If WorktreesDir is empty, it checks for a legacy .claude/worktrees/ dir
// and falls back to .worktrees/.
func (c *Config) GetWorktreesDir(repoRoot string) string {
	if c.WorktreesDir != "" {
		if filepath.IsAbs(c.WorktreesDir) {
			return c.WorktreesDir
		}
		return filepath.Join(repoRoot, c.WorktreesDir)
	}

	// Legacy fallback: if .claude/worktrees/ exists, use it
	legacyDir := filepath.Join(repoRoot, ".claude", "worktrees")
	if info, err := os.Stat(legacyDir); err == nil && info.IsDir() {
		return legacyDir
	}

	return filepath.Join(repoRoot, ".worktrees")
}

// Load reads and parses .wt.yaml from the given repo root.
func Load(repoRoot string) (*Config, error) {
	path := filepath.Join(repoRoot, FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s found in %s — create one to define services", FileName, repoRoot)
		}
		return nil, fmt.Errorf("reading %s: %w", FileName, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", FileName, err)
	}

	if len(cfg.Services) == 0 {
		return nil, fmt.Errorf("%s has no services defined", FileName)
	}

	// Default dir to "."
	for name, svc := range cfg.Services {
		if svc.Dir == "" {
			svc.Dir = "."
		}
		if svc.Cmd == "" {
			return nil, fmt.Errorf("service '%s' has no cmd defined", name)
		}
		cfg.Services[name] = svc
	}

	return &cfg, nil
}

// LoadOrEmpty reads .wt.yaml if it exists, or returns an empty config.
func LoadOrEmpty(repoRoot string) *Config {
	cfg, err := Load(repoRoot)
	if err != nil {
		return &Config{Services: make(map[string]Service)}
	}
	return cfg
}

// Save writes the config to .wt.yaml in the given repo root.
func Save(repoRoot string, cfg *Config) error {
	path := filepath.Join(repoRoot, FileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
