package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	DefaultAgent = "claude"
)

type Project struct {
	Path  string `json:"path"`
	Agent string `json:"agent"`
}

type Config struct {
	DefaultProjectsDir string             `json:"defaultProjectsDir"`
	DefaultAgent       string             `json:"defaultAgent"`
	Projects           map[string]Project `json:"projects"`
}

func Default() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find home directory: %w", err)
	}
	return &Config{
		DefaultProjectsDir: filepath.Join(home, "Projects", "personal"),
		DefaultAgent:       DefaultAgent,
		Projects:           map[string]Project{},
	}, nil
}

func Path(override string) (string, error) {
	if override != "" {
		return expandPath(override)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user config directory: %w", err)
	}
	return filepath.Join(dir, "devx", "config.json"), nil
}

func Load(pathOverride string) (*Config, string, error) {
	path, err := Path(pathOverride)
	if err != nil {
		return nil, "", err
	}

	cfg, err := Default()
	if err != nil {
		return nil, "", err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := Save(path, cfg); err != nil {
			return nil, "", err
		}
		return cfg, path, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, "", fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Projects == nil {
		cfg.Projects = map[string]Project{}
	}
	if cfg.DefaultAgent == "" {
		cfg.DefaultAgent = DefaultAgent
	}
	if cfg.DefaultProjectsDir == "" {
		defaultCfg, _ := Default()
		cfg.DefaultProjectsDir = defaultCfg.DefaultProjectsDir
	}
	return cfg, path, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write temporary config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func Names(cfg *Config) []string {
	names := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ExpandPath(path string) (string, error) {
	return expandPath(path)
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path == "~" || len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}
