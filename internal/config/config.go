package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

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

// Status struct controls how `hive status` formats its output for
// human and tmux consumers. JSON output is fixed-schema and not templated.
type Status struct {
	HumanFormat string `yaml:"human_format"`
	TmuxFormat  string `yaml:"tmux_format"`
}

type Config struct {
	Notifications Notifications `yaml:"notifications"`
	Queue         Queue         `yaml:"queue"`
	History       History       `yaml:"history"`
	Status        Status        `yaml:"status"`
}

const defaultHumanFormat = `{{- if eq .Count 0 -}}
🐝 No agents waiting
{{- else -}}
🐝 {{ bold .Count }} agent{{ if gt .Count 1 }}s{{ end }} waiting

  ▸ {{ bold .Next.Session }}{{ if .Next.Message }} — {{ .Next.Message }}{{ end }} {{ dim (printf "(%s)" .Next.Age) }}
{{- range slice .Queue 1 }}
    {{ .Session }}{{ if .Message }} — {{ .Message }}{{ end }} {{ dim (printf "(%s)" .Age) }}
{{- end }}
{{- end -}}`

const defaultTmuxFormat = `{{- if .Next -}}#[fg=colour220,bold]🐝 {{ .Next.Session }}{{ if .Next.Message }}: {{ .Next.Message }}{{ end }} ({{ .Next.Age }}){{ if gt .Count 1 }} | +{{ len (slice .Queue 1) }}{{ end }}#[fg=default,nobold] {{ end -}}`

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
		Status: Status{
			HumanFormat: defaultHumanFormat,
			TmuxFormat:  defaultTmuxFormat,
		},
	}
}

// ConfigPath returns the path to the config file. Exported so the cmd
// package can reference it (e.g., in `hive config edit`).
func ConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".hive", "config.yaml")
}

// Load reads ~/.hive/config.yaml and returns a Config populated by overlaying
// the file's contents on top of defaults. On first run, when the file does not
// exist, Load creates it with a full snapshot of the defaults so the user
// always has a discoverable, editable file containing every key with its value.
//
// Trade-off: a written-out default is pinned, so a default changed in a future
// hive version won't reach a file that already exists. A schema_version +
// migration mechanism is planned to address this (see TODO.txt).
func Load() (*Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			if err := createDefaultFile(); err != nil {
				return nil, err
			}
			data, err = os.ReadFile(ConfigPath())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	// Unmarshal into the defaults-populated struct so keys absent from the
	// YAML keep their default values (partial config files are valid).
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", ConfigPath(), err)
	}
	return cfg, nil
}

// createDefaultFile writes a full snapshot of the defaults to ConfigPath using
// O_CREATE|O_EXCL, so a concurrent invocation racing to create the same file
// cannot clobber one another process just wrote. An "already exists" error is
// treated as success: the caller proceeds to read whatever is on disk.
func createDefaultFile() error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(defaultConfig())
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func Save(cfg *Config) error {
	if err := validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
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
//
// For string-typed fields, a value starting with "@" is interpreted as a
// preset name (e.g., `@compact`). The preset library is consulted via
// resolvePreset; unknown preset names return an error listing what's
// available for the target key.
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
			if field.Kind() == reflect.String && strings.HasPrefix(value, "@") {
				resolved, err := resolvePreset(key, value[1:])
				if err != nil {
					return err
				}
				value = resolved
			}
			return assignField(field, value)
		}
		v = field
	}
	return nil
}

// Reset restores the field at the dotted yaml-tag path key to its default
// value, leaving the rest of cfg untouched. Returns an error on unknown
// key or mid-path leaf.
func Reset(cfg *Config, key string) error {
	defaults := defaultConfig()
	parts := strings.Split(key, ".")
	target := reflect.ValueOf(cfg).Elem()
	source := reflect.ValueOf(defaults).Elem()
	for i, part := range parts {
		if target.Kind() != reflect.Struct {
			return fmt.Errorf("config: %q is not a leaf key", strings.Join(parts[:i], "."))
		}
		tField, ok := fieldByYAMLTag(target, part)
		if !ok {
			return fmt.Errorf("config: unknown key %q", key)
		}
		sField, _ := fieldByYAMLTag(source, part)
		if i == len(parts)-1 {
			tField.Set(sField)
			return nil
		}
		target = tField
		source = sField
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
	if _, err := template.New("status.human_format").Funcs(TemplateFuncs()).Parse(cfg.Status.HumanFormat); err != nil {
		return fmt.Errorf("status.human_format: %w", err)
	}
	if _, err := template.New("status.tmux_format").Funcs(TemplateFuncs()).Parse(cfg.Status.TmuxFormat); err != nil {
		return fmt.Errorf("status.tmux_format: %w", err)
	}
	return nil
}

// TemplateFuncs returns the helper functions available inside
// status.human_format and status.tmux_format templates, on top of the
// text/template built-ins. Both the validator (this package) and the
// executor (cmd package) register these names so any template that parses
// cleanly at startup also runs cleanly at status time.
//
// `bold` and `dim` are no-ops here (return the value as plain text). The
// cmd-side execution path overrides them with ANSI-emitting versions when
// the destination writer is a terminal — see cmd/status.go.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add":  func(a, b int) int { return a + b },
		"bold": func(v any) string { return fmt.Sprintf("%v", v) },
		"dim":  func(v any) string { return fmt.Sprintf("%v", v) },
	}
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
