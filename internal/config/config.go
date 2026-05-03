package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Notifications struct {
	Macos    bool `yaml:"macos"`
	TmuxBell bool `yaml:"tmux_bell"`
}

type Queue struct {
	MaxMessageLength int `yaml:"max_message_length"`
}

type History struct {
	// RetentionDays controls how long resolved sessions stay in the history
	// table. The purge runs on every db.Open and deletes rows whose
	// resolved_at is older than (now - RetentionDays days). Setting
	// RetentionDays to 0 wipes the table on every invocation, effectively
	// disabling history. Must be >= 0; negative values are rejected at
	// validation time.
	RetentionDays int `yaml:"retention_days"`
}

type Config struct {
	Notifications Notifications `yaml:"notifications"`
	Queue         Queue         `yaml:"queue"`
	History       History       `yaml:"history"`
}

func defaultConfig() *Config {
	return &Config{
		Notifications: Notifications{
			Macos:    true,
			TmuxBell: true,
		},
		Queue: Queue{
			MaxMessageLength: 100,
		},
		History: History{
			RetentionDays: 7,
		},
	}
}

func configPath() string {
	return filepath.Join(os.Getenv("HOME"), ".hive", "config.yaml")
}

// Load reads ~/.hive/config.yaml and returns a Config populated by overlaying
// the file's contents on top of DefaultConfig. If the file does not exist,
// Load writes the defaults to disk before returning, so subsequent invocations
// (and the user) can find and edit the file.
func Load() (*Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			// If no configuration file exists, we write the defaults to disk so
			// the user can find and edit it.
			if saveErr := Save(cfg); saveErr != nil {
				return nil, fmt.Errorf("write default config: %w", saveErr)
			}
			return cfg, nil
		}
		return nil, err
	}
	// Unmarshal into the defaults-populated struct so keys absent from the
	// YAML keep their default values (partial config files are valid).
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", configPath(), err)
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	if err := validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Set assigns value to the field at the dotted yaml-tag path key.
// Supports bool, int, and string field types. Returns an error on unknown
// key, mid-path leaf, or value that doesn't parse as the field's type.
func Set(cfg *Config, key, value string) error {
	parts := strings.Split(key, ".")
	v := reflect.ValueOf(cfg).Elem()

	for i, part := range parts {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("config: %q is not a leaf key", strings.Join(parts[:i], "."))
		}
		field, ok := fieldByYAMLTag(v, part)
		if !ok {
			return fmt.Errorf("config: unknown key %q", key)
		}
		if i == len(parts)-1 {
			return assignField(field, value)
		}
		v = field
	}
	return nil
}

// validate returns an error if any field of cfg holds a value that hive cannot handle.
func validate(cfg *Config) error {
	if cfg.Queue.MaxMessageLength <= 0 {
		return fmt.Errorf("queue.max_message_length must be > 0, got %d", cfg.Queue.MaxMessageLength)
	}
	if cfg.History.RetentionDays < 0 {
		return fmt.Errorf("history.retention_days must be >= 0, got %d", cfg.History.RetentionDays)
	}
	return nil
}

func fieldByYAMLTag(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		tag := strings.Split(t.Field(i).Tag.Get("yaml"), ",")[0]
		if tag == name {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func assignField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("config: field is not settable")
	}
	switch field.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("config: expected bool, got %q", value)
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("config: expected int, got %q", value)
		}
		field.SetInt(n)
	case reflect.String:
		field.SetString(value)
	default:
		return fmt.Errorf("config: unsupported field type %s", field.Kind())
	}
	return nil
}
