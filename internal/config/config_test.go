package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if !cfg.Notifications.Macos {
		t.Error("expected Notifications.Macos=true by default")
	}
	if !cfg.Notifications.TmuxBell {
		t.Error("expected Notifications.TmuxBell=true by default")
	}
	if got, want := cfg.Queue.MaxMessageLength, 100; got != want {
		t.Errorf("Queue.MaxMessageLength = %d, want %d", got, want)
	}
}

func TestLoadMissingFileReturnsDefaultsWithoutWriting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Queue.MaxMessageLength != 100 {
		t.Errorf("expected defaults when config file is absent, got %+v", cfg)
	}
	// Load must NOT create the file. Auto-creating it leads to staleness when
	// new fields are added in future versions — old files would shadow new
	// defaults. The file appears only when the user explicitly customizes.
	if _, err := os.Stat(ConfigPath()); !os.IsNotExist(err) {
		t.Errorf("expected Load to leave %s missing, got stat err: %v", ConfigPath(), err)
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := defaultConfig()
	cfg.Notifications.Macos = false
	cfg.Queue.MaxMessageLength = 50

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Notifications.Macos {
		t.Error("expected Macos=false after round-trip")
	}
	if loaded.Notifications.TmuxBell != true {
		t.Error("expected TmuxBell=true to round-trip from default")
	}
	if got, want := loaded.Queue.MaxMessageLength, 50; got != want {
		t.Errorf("MaxMessageLength = %d, want %d", got, want)
	}
}

// TestLoadPartialFileKeepsDefaults guards against regressing the
// defaults-overlay behavior. A YAML file that only sets some keys must not
// zero out the others.
func TestLoadPartialFileKeepsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := os.MkdirAll(filepath.Dir(ConfigPath()), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	partial := []byte("queue:\n    max_message_length: 50\n")
	if err := os.WriteFile(ConfigPath(), partial, 0o644); err != nil {
		t.Fatalf("write partial config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, want := cfg.Queue.MaxMessageLength, 50; got != want {
		t.Errorf("MaxMessageLength = %d, want %d (override)", got, want)
	}
	if !cfg.Notifications.Macos {
		t.Error("Notifications.Macos = false; want true (default preserved)")
	}
	if !cfg.Notifications.TmuxBell {
		t.Error("Notifications.TmuxBell = false; want true (default preserved)")
	}
	if got, want := cfg.History.RetentionDays, 7; got != want {
		t.Errorf("History.RetentionDays = %d, want %d (default preserved)", got, want)
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("HOME", "/test/home")
	got := ConfigPath()
	want := filepath.Join("/test/home", ".hive", "config.yaml")
	if got != want {
		t.Errorf("ConfigPath = %q, want %q", got, want)
	}
}

func TestSetBool(t *testing.T) {
	cfg := defaultConfig()
	if err := Set(cfg, "notifications.macos", "false"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if cfg.Notifications.Macos {
		t.Error("expected Macos=false after Set")
	}
}

func TestSetInt(t *testing.T) {
	cfg := defaultConfig()
	if err := Set(cfg, "queue.max_message_length", "250"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, want := cfg.Queue.MaxMessageLength, 250; got != want {
		t.Errorf("MaxMessageLength = %d, want %d", got, want)
	}
}

func TestSetUnknownKey(t *testing.T) {
	cfg := defaultConfig()
	err := Set(cfg, "foo.bar", "baz")
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
}

func TestSetUnknownNestedKey(t *testing.T) {
	cfg := defaultConfig()
	err := Set(cfg, "notifications.unknown_field", "true")
	if err == nil {
		t.Fatal("expected error for unknown nested key, got nil")
	}
}

func TestSetWrongType(t *testing.T) {
	cfg := defaultConfig()
	err := Set(cfg, "notifications.macos", "maybe")
	if err == nil {
		t.Fatal("expected error for non-bool value, got nil")
	}
}

func TestValidate_Defaults(t *testing.T) {
	if err := validate(defaultConfig()); err != nil {
		t.Errorf("DefaultConfig should be valid, got: %v", err)
	}
}

func TestValidate_NegativeRetention(t *testing.T) {
	cfg := defaultConfig()
	cfg.History.RetentionDays = -1
	if err := validate(cfg); err == nil {
		t.Error("expected error for negative retention_days, got nil")
	}
}

func TestValidate_ZeroRetentionAllowed(t *testing.T) {
	// 0 is the documented "no history" semantics — must pass validation.
	cfg := defaultConfig()
	cfg.History.RetentionDays = 0
	if err := validate(cfg); err != nil {
		t.Errorf("retention_days=0 should be valid, got: %v", err)
	}
}

func TestValidate_NonPositiveMessageLength(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		cfg := defaultConfig()
		cfg.Queue.MaxMessageLength = n
		if err := validate(cfg); err == nil {
			t.Errorf("expected error for max_message_length=%d, got nil", n)
		}
	}
}

func TestSaveRejectsInvalidConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := defaultConfig()
	cfg.History.RetentionDays = -5
	if err := Save(cfg); err == nil {
		t.Error("expected Save to reject invalid config, got nil")
	}
}

func TestLoadRejectsInvalidFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(ConfigPath()), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	bad := []byte("history:\n    retention_days: -3\n")
	if err := os.WriteFile(ConfigPath(), bad, 0o644); err != nil {
		t.Fatalf("write bad config: %v", err)
	}
	if _, err := Load(); err == nil {
		t.Error("expected Load to reject invalid file, got nil")
	}
}

func TestValidate_BadHumanTemplate(t *testing.T) {
	cfg := defaultConfig()
	cfg.Status.HumanFormat = `{{ .BadTemplate`
	if err := validate(cfg); err == nil {
		t.Error("expected validate to reject malformed human_format template, got nil")
	}
}

func TestValidate_BadTmuxTemplate(t *testing.T) {
	cfg := defaultConfig()
	cfg.Status.TmuxFormat = `{{ if .Next`
	if err := validate(cfg); err == nil {
		t.Error("expected validate to reject malformed tmux_format template, got nil")
	}
}

func TestValidate_DefaultTemplatesPass(t *testing.T) {
	// Defensive: the defaults shipped in defaultConfig must parse cleanly.
	if err := validate(defaultConfig()); err != nil {
		t.Errorf("default templates should validate, got: %v", err)
	}
}

func TestSetNonLeafKey(t *testing.T) {
	cfg := defaultConfig()
	err := Set(cfg, "notifications", "true")
	if err == nil {
		t.Fatal("expected error when assigning to a non-leaf key, got nil")
	}
}

func TestReset_BoolField(t *testing.T) {
	cfg := defaultConfig()
	cfg.Notifications.Macos = false
	if err := Reset(cfg, "notifications.macos"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if !cfg.Notifications.Macos {
		t.Error("expected Macos=true (default) after Reset")
	}
}

func TestReset_IntField(t *testing.T) {
	cfg := defaultConfig()
	cfg.Queue.MaxMessageLength = 50
	if err := Reset(cfg, "queue.max_message_length"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if got, want := cfg.Queue.MaxMessageLength, 100; got != want {
		t.Errorf("MaxMessageLength = %d, want %d (default)", got, want)
	}
}

func TestReset_StringField(t *testing.T) {
	cfg := defaultConfig()
	original := cfg.Status.HumanFormat
	cfg.Status.HumanFormat = "custom"
	if err := Reset(cfg, "status.human_format"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if cfg.Status.HumanFormat != original {
		t.Errorf("HumanFormat = %q, want %q (default)", cfg.Status.HumanFormat, original)
	}
}

func TestReset_UnknownKey(t *testing.T) {
	cfg := defaultConfig()
	if err := Reset(cfg, "foo.bar"); err == nil {
		t.Error("expected Reset to reject unknown key, got nil")
	}
}

func TestSet_StoresPresetReferenceAsIs(t *testing.T) {
	cfg := defaultConfig()
	if err := Set(cfg, "status.human_format", "compact"); err != nil {
		t.Fatalf("Set with preset name: %v", err)
	}
	if cfg.Status.HumanFormat != "compact" {
		t.Errorf("HumanFormat = %q, want %q", cfg.Status.HumanFormat, "compact")
	}
}

func TestSet_UnknownNameCaughtByValidate(t *testing.T) {
	cfg := defaultConfig()
	// Set itself doesn't resolve template names — it just stores the value.
	if err := Set(cfg, "status.human_format", "nonexistent"); err != nil {
		t.Errorf("Set should accept unknown template name (validate catches it later): %v", err)
	}
	// validate runs ResolveTemplate which reports the error.
	if err := validate(cfg); err == nil {
		t.Error("expected validate to reject unknown template name, got nil")
	}
}

func TestResolveTemplate_PresetName(t *testing.T) {
	got, err := ResolveTemplate("status.human_format", "compact")
	if err != nil {
		t.Fatalf("ResolveTemplate: %v", err)
	}
	if got != statusHumanPresets["compact"] {
		t.Error("ResolveTemplate did not return the compact preset content")
	}
}

func TestResolveTemplate_KeyScopedPresets(t *testing.T) {
	// "minimal" is a tmux preset only — must NOT resolve under human_format.
	if _, err := ResolveTemplate("status.human_format", "minimal"); err == nil {
		t.Error("expected ResolveTemplate to reject 'minimal' for human_format (it's only a tmux preset)")
	}
	// And it MUST resolve under tmux_format.
	if _, err := ResolveTemplate("status.tmux_format", "minimal"); err != nil {
		t.Errorf("ResolveTemplate failed for tmux_format/minimal: %v", err)
	}
}

func TestResolveTemplate_CustomTemplateFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := filepath.Join(os.Getenv("HOME"), ".hive", "templates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	contents := "hello {{ .Next.Session }}"
	if err := os.WriteFile(filepath.Join(dir, "mine.tmpl"), []byte(contents), 0o644); err != nil {
		t.Fatalf("write tmpl: %v", err)
	}
	got, err := ResolveTemplate("status.human_format", "mine")
	if err != nil {
		t.Fatalf("ResolveTemplate: %v", err)
	}
	if got != contents {
		t.Errorf("ResolveTemplate returned %q, want %q", got, contents)
	}
}

func TestResolveTemplate_MissingName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if _, err := ResolveTemplate("status.human_format", "nonexistent"); err == nil {
		t.Error("expected ResolveTemplate to error on unknown name, got nil")
	}
}

func TestResolveTemplate_RejectsPathSeparator(t *testing.T) {
	if _, err := ResolveTemplate("status.human_format", "/abs/path"); err == nil {
		t.Error("expected ResolveTemplate to reject value with slash, got nil")
	}
}

func TestSet_NameOnNonStringField(t *testing.T) {
	// Bool fields parse the value as bool — a name like "compact" fails.
	cfg := defaultConfig()
	err := Set(cfg, "notifications.macos", "compact")
	if err == nil {
		t.Error("expected Set non-bool value on bool field to error, got nil")
	}
}

func TestPresets_AllValidate(t *testing.T) {
	// Every shipped preset must validate via its bare name.
	for name := range statusHumanPresets {
		cfg := defaultConfig()
		cfg.Status.HumanFormat = name
		if err := validate(cfg); err != nil {
			t.Errorf("status.human_format preset %q fails validation: %v", name, err)
		}
	}
	for name := range statusTmuxPresets {
		cfg := defaultConfig()
		cfg.Status.TmuxFormat = name
		if err := validate(cfg); err != nil {
			t.Errorf("status.tmux_format preset %q fails validation: %v", name, err)
		}
	}
}
