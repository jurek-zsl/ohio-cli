package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultApiUrl  = "https://api.ohiofiles.cloud"
	ConfigDirName  = "ohio"
	ConfigFileName = "config.json"
)

// Config represents the persisted CLI configuration.
type Config struct {
	ApiUrl      string `json:"apiUrl"`
	SessionKey  string `json:"sessionKey,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	ProfileSlug string `json:"profileSlug,omitempty"`
	DeviceId    string `json:"deviceId,omitempty"`
}

// GetConfigDir returns the directory path for the configuration file.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	primary := filepath.Join(home, ".config", "ohio")
	legacy := filepath.Join(home, ".config", "ohfs")
	if _, err := os.Stat(primary); err == nil {
		return primary, nil
	}
	if _, err := os.Stat(legacy); err == nil {
		return legacy, nil
	}
	return primary, nil
}

// GetConfigPath returns the absolute path to config.json.
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// Load loads the configuration from disk, creating defaults if not found.
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = cfg.Save()
			return cfg, nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), nil
	}

	if cfg.ApiUrl == "" {
		cfg.ApiUrl = DefaultApiUrl
	}
	if cfg.DeviceId == "" {
		cfg.EnsureDeviceId()
		_ = cfg.Save()
	}

	return cfg, nil
}

// DefaultConfig returns default configuration.
func DefaultConfig() *Config {
	cfg := &Config{
		ApiUrl: DefaultApiUrl,
	}
	cfg.EnsureDeviceId()
	return cfg
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ResolveApiUrl returns the resolved API URL, prioritizing environment variables OHIO_API_URL or OHFS_API_URL.
func (c *Config) ResolveApiUrl() string {
	if envUrl := strings.TrimSpace(os.Getenv("OHIO_API_URL")); envUrl != "" {
		return strings.TrimRight(envUrl, "/")
	}
	if envUrl := strings.TrimSpace(os.Getenv("OHFS_API_URL")); envUrl != "" {
		return strings.TrimRight(envUrl, "/")
	}
	if strings.TrimSpace(c.ApiUrl) != "" {
		return strings.TrimRight(strings.TrimSpace(c.ApiUrl), "/")
	}
	return DefaultApiUrl
}

// EnsureDeviceId ensures a persistent unique Device ID exists.
func (c *Config) EnsureDeviceId() string {
	if c.DeviceId != "" {
		return c.DeviceId
	}
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err == nil {
		c.DeviceId = "cli_" + hex.EncodeToString(bytes)
	} else {
		c.DeviceId = "cli_device_default"
	}
	return c.DeviceId
}

// SetActiveSession updates the active session and saves to disk.
func (c *Config) SetActiveSession(key, nickname, profileSlug string) error {
	c.SessionKey = strings.TrimSpace(key)
	c.Nickname = strings.TrimSpace(nickname)
	c.ProfileSlug = strings.TrimSpace(profileSlug)
	return c.Save()
}

// ClearActiveSession clears the current session.
func (c *Config) ClearActiveSession() error {
	c.SessionKey = ""
	c.Nickname = ""
	c.ProfileSlug = ""
	return c.Save()
}

// ValidateSession returns an error if no session key is configured.
func (c *Config) ValidateSession() error {
	if strings.TrimSpace(c.SessionKey) == "" {
		return errors.New("no active session configured; run 'ohfs session new' or 'ohfs session set <key>'")
	}
	return nil
}
