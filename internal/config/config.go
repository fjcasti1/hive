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
	// RetentionDays purges history entries older than this on `db.Open`.
	// 0 or negative disables purging.
	RetentionDays int `yaml:"retention_days"`
}

type Config struct {
	Notifications Notifications `yaml:"notifications"`
	Queue         Queue         `yaml:"queue"`
	History       History       `yaml:"history"`
}

func DefaultConfig() Config {
	return Config{
		Notifications: Notifications{Macos: true, TmuxBell: true},
		Queue:         Queue{MaxMessageLength: 100},
		History:       History{RetentionDays: 7},
	}
}

func ConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".hive", "config.yaml")
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
