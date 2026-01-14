package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const ConfigFileName = ".wt.json"

// Config represents the .wt.json configuration file
type Config struct {
	Version  string          `json:"version"`
	Worktree WorktreeConfig  `json:"worktree"`
	Setup    SetupConfig     `json:"setup"`
}

// WorktreeConfig defines worktree creation settings
type WorktreeConfig struct {
	Basedir  string            `json:"basedir"`
	Naming   string            `json:"naming"`
	Sanitize map[string]string `json:"sanitize"`
}

// SetupConfig defines files to copy or link
type SetupConfig struct {
	Copy []string `json:"copy"`
	Link []string `json:"link"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Worktree: WorktreeConfig{
			Basedir:  "../",
			Naming:   "{repo}-{branch}",
			Sanitize: map[string]string{"/": "-", ":": "-"},
		},
		Setup: SetupConfig{
			Copy: []string{},
			Link: []string{},
		},
	}
}

// Load loads configuration from the given directory
// It searches upward to find .wt.json in a git repository
func Load(startDir string) (*Config, error) {
	configPath, err := FindConfigFile(startDir)
	if err != nil {
		// Return default config if not found
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// FindConfigFile searches for .wt.json starting from dir and going up
func FindConfigFile(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(absDir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Check if we've reached root
		parent := filepath.Dir(absDir)
		if parent == absDir {
			return "", os.ErrNotExist
		}
		absDir = parent
	}
}

// Save saves the configuration to the given path
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GenerateWorktreeName generates a worktree directory name from branch
func (c *Config) GenerateWorktreeName(repoName, branch string) string {
	name := c.Worktree.Naming
	name = strings.ReplaceAll(name, "{repo}", repoName)
	name = strings.ReplaceAll(name, "{branch}", branch)

	// Apply sanitization
	for old, new := range c.Worktree.Sanitize {
		name = strings.ReplaceAll(name, old, new)
	}

	return name
}

// GetWorktreeBasedir returns the absolute base directory for worktrees
func (c *Config) GetWorktreeBasedir(repoRoot string) (string, error) {
	if filepath.IsAbs(c.Worktree.Basedir) {
		return c.Worktree.Basedir, nil
	}
	return filepath.Abs(filepath.Join(repoRoot, c.Worktree.Basedir))
}
