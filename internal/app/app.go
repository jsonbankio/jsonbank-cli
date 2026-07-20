// Package app holds the CLI's runtime context: where it was invoked from,
// where it keeps its own files, and the config it remembers between runs.
package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	Host     string          `json:"host"`
	Keys     Keys            `json:"keys"`     // active account's keys (what the SDK uses)
	Accounts map[string]Keys `json:"accounts"` // all saved accounts, keyed by username
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
	// Keys often pick up stray whitespace when copy-pasted; trim before saving.
	a.Config.Keys = trimKeys(a.Config.Keys)
	for username, k := range a.Config.Accounts {
		a.Config.Accounts[username] = trimKeys(k)
	}

	data, err := json.MarshalIndent(a.Config, "", "  ")
	if err != nil {
		return err
	}
	// 0600: the file holds secret keys — keep it owner-only.
	return os.WriteFile(a.ConfigPath(), data, 0o600)
}

// ActiveKeys returns the keys to authenticate with. Environment variables
// override the saved active account, so CI can supply keys without a login step.
func (a *App) ActiveKeys() Keys {
	pub := strings.TrimSpace(os.Getenv("JSB_PUBLIC_KEY"))
	prv := strings.TrimSpace(os.Getenv("JSB_PRIVATE_KEY"))
	if pub != "" || prv != "" {
		return Keys{Public: pub, Private: prv}
	}
	return a.Config.Keys
}

// ActiveUsername returns the username of the account whose keys are currently
// active, or "" if none of the stored accounts matches.
func (c *Config) ActiveUsername() string {
	if c.Keys == (Keys{}) {
		return ""
	}
	for username, k := range c.Accounts {
		if k == c.Keys {
			return username
		}
	}
	return ""
}

// Activate makes the named account's keys the active ones.
func (c *Config) Activate(username string) {
	c.Keys = c.Accounts[username]
}

// RemoveAccount deletes an account. If it was the active one, the active keys
// are cleared too. Returns false if the account did not exist.
func (c *Config) RemoveAccount(username string) bool {
	k, ok := c.Accounts[username]
	if !ok {
		return false
	}
	delete(c.Accounts, username)
	if k == c.Keys {
		c.Keys = Keys{}
	}
	return true
}

// trimKeys returns a copy of k with surrounding whitespace removed from each key.
func trimKeys(k Keys) Keys {
	return Keys{
		Public:  strings.TrimSpace(k.Public),
		Private: strings.TrimSpace(k.Private),
	}
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
	cfg := &Config{Host: DefaultHost, Accounts: map[string]Keys{}}

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
	if cfg.Accounts == nil {
		cfg.Accounts = map[string]Keys{}
	}
	return cfg, nil
}
