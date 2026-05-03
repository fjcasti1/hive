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

func TestLoadMissingFileWritesDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Queue.MaxMessageLength != 100 {
		t.Errorf("expected defaults when config file is absent, got %+v", cfg)
	}
	// Load should have materialized the defaults to disk so the user can find
	// and edit the file.
	if _, err := os.Stat(configPath()); err != nil {
		t.Errorf("expected Load to create %s, got stat error: %v", configPath(), err)
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

	if err := os.MkdirAll(filepath.Dir(configPath()), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	partial := []byte("queue:\n    max_message_length: 50\n")
	if err := os.WriteFile(configPath(), partial, 0o644); err != nil {
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
	got := configPath()
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
	if err := os.MkdirAll(filepath.Dir(configPath()), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	bad := []byte("history:\n    retention_days: -3\n")
	if err := os.WriteFile(configPath(), bad, 0o644); err != nil {
		t.Fatalf("write bad config: %v", err)
	}
	if _, err := Load(); err == nil {
		t.Error("expected Load to reject invalid file, got nil")
	}
}

func TestSetNonLeafKey(t *testing.T) {
	cfg := defaultConfig()
	err := Set(cfg, "notifications", "true")
	if err == nil {
		t.Fatal("expected error when assigning to a non-leaf key, got nil")
	}
}
