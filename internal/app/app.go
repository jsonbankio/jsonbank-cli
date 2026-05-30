// Package app holds the CLI's runtime context: where it was invoked from,
// where it keeps its own files, and the config it remembers between runs.
package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// DefaultHost is the JSONBank API the CLI talks to unless overridden.
const DefaultHost = "https://api.jsonbank.io"

const configFile = "config.json"

// Keys are the JSONBank API keys.
type Keys struct {
	Public  string `json:"pub"`
	Private string `json:"prv"`
}

// Config is what the CLI persists in its memory directory.
type Config struct {
	Host string `json:"host"`
	Keys Keys   `json:"keys"`
}

// App is the initialized CLI context.
type App struct {
	Cwd    string  // directory the command was run from
	Dir    string  // the CLI's own memory directory
	Config *Config // config loaded from Dir/config.json
}

// Init sets up the CLI: resolves the working directory, ensures the memory
// directory exists, and loads (or starts) the config.
func Init() (*App, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dir, err := memoryDir()
	if err != nil {
		return nil, err
	}

	cfg, err := loadConfig(filepath.Join(dir, configFile))
	if err != nil {
		return nil, err
	}

	return &App{Cwd: cwd, Dir: dir, Config: cfg}, nil
}

// ConfigPath is the full path to the config file.
func (a *App) ConfigPath() string {
	return filepath.Join(a.Dir, configFile)
}

// Save writes the current config back to the memory directory.
func (a *App) Save() error {
	data, err := json.MarshalIndent(a.Config, "", "  ")
	if err != nil {
		return err
	}
	// 0600 — the file can hold a private key.
	return os.WriteFile(a.ConfigPath(), data, 0o600)
}

// memoryDir returns the CLI's config directory, creating it if needed.
func memoryDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "jsb")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// loadConfig reads config.json, returning a fresh default if none exists.
func loadConfig(path string) (*Config, error) {
	cfg := &Config{Host: DefaultHost}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Host == "" {
		cfg.Host = DefaultHost
	}
	return cfg, nil
}
