package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all queryli configuration.
type Config struct {
	ActiveProfile string             `yaml:"active_profile" json:"active_profile"`
	Profiles      map[string]Profile `yaml:"profiles" json:"profiles"`
}

// Profile holds connection details for a database.
type Profile struct {
	Type     string `yaml:"type" json:"type"`         // postgres | mysql | sqlite | oracle
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"dbname" json:"dbname"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`   // postgres
	Path     string `yaml:"path" json:"path"`         // sqlite
	Service  string `yaml:"service" json:"service"`   // oracle
}

// Load reads the config from a YAML file. Returns an empty config if file doesn't exist.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return cfg, nil
}

// Save writes the config to a YAML file.
func Save(cfg *Config, path string) error {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		path = filepath.Join(home, ".queryli", "config.yaml")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// ResolveProfile returns the active profile name: arg > active_profile in config.
func ResolveProfile(cfg *Config, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	if cfg.ActiveProfile != "" {
		return cfg.ActiveProfile, nil
	}
	return "", fmt.Errorf("no profile specified and no active_profile set; use --profile flag or 'queryli profile use <name>'")
}
