package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Notifications struct {
	Macos    bool `yaml:"macos"`
	TmuxBell bool `yaml:"tmux_bell"`
}

type Queue struct {
	MaxMessageLength int `yaml:"max_message_length"`
}

type Config struct {
	Notifications Notifications `yaml:"notifications"`
	Queue         Queue         `yaml:"queue"`
}

func DefaultConfig() Config {
	return Config{
		Notifications: Notifications{Macos: true, TmuxBell: true},
		Queue:         Queue{MaxMessageLength: 100},
	}
}

func ConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".hive", "config", "config.yaml")
}

func Load() (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	return cfg, yaml.Unmarshal(data, &cfg)
}

func Save(cfg Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
