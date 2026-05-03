package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
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

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Queue.MaxMessageLength != 100 {
		t.Errorf("expected defaults when config file is absent, got %+v", cfg)
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := DefaultConfig()
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

func TestConfigPath(t *testing.T) {
	t.Setenv("HOME", "/test/home")
	got := ConfigPath()
	want := filepath.Join("/test/home", ".hive", "config.yaml")
	if got != want {
		t.Errorf("ConfigPath = %q, want %q", got, want)
	}
}

func TestSetBool(t *testing.T) {
	cfg := DefaultConfig()
	if err := Set(&cfg, "notifications.macos", "false"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if cfg.Notifications.Macos {
		t.Error("expected Macos=false after Set")
	}
}

func TestSetInt(t *testing.T) {
	cfg := DefaultConfig()
	if err := Set(&cfg, "queue.max_message_length", "250"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, want := cfg.Queue.MaxMessageLength, 250; got != want {
		t.Errorf("MaxMessageLength = %d, want %d", got, want)
	}
}

func TestSetUnknownKey(t *testing.T) {
	cfg := DefaultConfig()
	err := Set(&cfg, "foo.bar", "baz")
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
}

func TestSetUnknownNestedKey(t *testing.T) {
	cfg := DefaultConfig()
	err := Set(&cfg, "notifications.unknown_field", "true")
	if err == nil {
		t.Fatal("expected error for unknown nested key, got nil")
	}
}

func TestSetWrongType(t *testing.T) {
	cfg := DefaultConfig()
	err := Set(&cfg, "notifications.macos", "maybe")
	if err == nil {
		t.Fatal("expected error for non-bool value, got nil")
	}
}

func TestSetNonLeafKey(t *testing.T) {
	cfg := DefaultConfig()
	err := Set(&cfg, "notifications", "true")
	if err == nil {
		t.Fatal("expected error when assigning to a non-leaf key, got nil")
	}
}
