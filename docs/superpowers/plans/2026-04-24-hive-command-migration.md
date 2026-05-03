# Hive Command Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the hive CLI from the chezmoi dotfiles source into the standalone `fjcasti1/hive` repo one command at a time, starting with a Foundation PR that brings all internal packages, the `cmd/` restructure, DEVELOPMENT.md, and Makefile, followed by 12 individual command PRs.

**Architecture:** Each task produces a working, buildable binary with exactly one more command than the previous. The Foundation PR establishes the package structure and internal packages (config, db, notify, session) so every subsequent command PR only touches `cmd/` and nothing else. All files are copied verbatim from the chezmoi source at `/Users/kikocastillo/.local/share/chezmoi/hive/`, with one change: every import path `github.com/kikocastillo/hive` becomes `github.com/fjcasti1/hive`.

**Tech Stack:** Go 1.26.1, cobra v1.10.2, modernc.org/sqlite, gopkg.in/yaml.v3, GitHub Actions, GoReleaser

---

## File Map

### Foundation PR creates:
| File | Notes |
|------|-------|
| `internal/config/config.go` | Verbatim copy, import path updated |
| `internal/config/config_test.go` | Verbatim copy, import path updated |
| `internal/db/db.go` | Verbatim copy, import path updated |
| `internal/db/history.go` | Verbatim copy, import path updated |
| `internal/db/queue.go` | Verbatim copy, import path updated |
| `internal/db/settings.go` | Verbatim copy, import path updated |
| `internal/db/db_test.go` | Copy + fix: Enqueue calls need 4th arg (pane `""`) |
| `internal/notify/notify.go` | Verbatim copy, import path updated |
| `internal/notify/macos.go` | Verbatim copy, import path updated |
| `internal/notify/tmux.go` | Verbatim copy, import path updated |
| `internal/notify/notify_test.go` | Verbatim copy, import path updated |
| `internal/session/session.go` | Verbatim copy, import path updated |
| `internal/session/session_test.go` | Verbatim copy, import path updated |
| `cmd/root.go` | Verbatim copy, import path updated; init() only registers historyCmd |
| `cmd/util.go` | New file: extracts `timeAgo` shared by history, list, status |
| `cmd/history.go` | Verbatim copy, import path updated |
| `main.go` | Rewritten: calls `cmd.SetVersion(Version); cmd.Execute()` |
| `DEVELOPMENT.md` | New file (see spec) |
| `Makefile` | New file (see spec) |

### Each command PR modifies:
| File | Notes |
|------|-------|
| `cmd/<command>.go` | New file, verbatim copy, import path updated |
| `cmd/root.go` | Add one line to `init()` AddCommand list |

---

## Task 1: Foundation PR — packages, cmd/ restructure, history

**Files:**
- Modify: `go.mod`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `internal/db/db.go`
- Create: `internal/db/history.go`
- Create: `internal/db/queue.go`
- Create: `internal/db/settings.go`
- Create: `internal/db/db_test.go`
- Create: `internal/notify/notify.go`
- Create: `internal/notify/macos.go`
- Create: `internal/notify/tmux.go`
- Create: `internal/notify/notify_test.go`
- Create: `internal/session/session.go`
- Create: `internal/session/session_test.go`
- Rewrite: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/util.go`
- Create: `cmd/history.go`
- Create: `DEVELOPMENT.md`
- Create: `Makefile`

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout -b feat/foundation-packages-history
```

- [ ] **Step 2: Add missing dependencies to go.mod**

Run:
```bash
cd /Users/kikocastillo/coding/personal/hive
go get modernc.org/sqlite@latest
go get gopkg.in/yaml.v3@latest
go mod tidy
```

Expected: `go.mod` and `go.sum` updated. Verify with:
```bash
grep 'modernc.org/sqlite\|yaml.v3' go.mod
```

- [ ] **Step 3: Create `internal/config/config.go`**

```go
// internal/config/config.go
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

type Status struct {
	// Tokens: {session} {message} {age} {count} {extra}
	// ": {message}" and " | +{extra}" are collapsed when empty/zero
	Format string `yaml:"format"`
}

type List struct {
	WatchInterval int `yaml:"watch_interval"` // seconds between refreshes in --watch mode
}

type Snooze struct {
	DefaultDuration string `yaml:"default_duration"` // e.g. "10m", "1h"
}

type Pause struct {
	DefaultDuration string `yaml:"default_duration"` // empty = indefinite, otherwise e.g. "2h"
}

type History struct {
	RetentionDays int `yaml:"retention_days"` // entries older than this are purged on open
}

type Config struct {
	Notifications Notifications `yaml:"notifications"`
	Queue         Queue         `yaml:"queue"`
	Status        Status        `yaml:"status"`
	List          List          `yaml:"list"`
	History       History       `yaml:"history"`
	Snooze        Snooze        `yaml:"snooze"`
	Pause         Pause         `yaml:"pause"`
}

func DefaultConfig() Config {
	return Config{
		Notifications: Notifications{Macos: true, TmuxBell: true},
		Queue:         Queue{MaxMessageLength: 100},
		Status:        Status{Format: "{session}: {message} ({age}) | +{extra}"},
		List:          List{WatchInterval: 2},
		History:       History{RetentionDays: 7},
		Snooze:        Snooze{DefaultDuration: "10m"},
		Pause:         Pause{DefaultDuration: ""},
	}
}

func ConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "hive", "config.yaml")
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Create `internal/config/config_test.go`**

```go
// internal/config/config_test.go
package config_test

import (
	"os"
	"testing"

	"github.com/fjcasti1/hive/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if !cfg.Notifications.Macos {
		t.Error("expected Notifications.Macos=true by default")
	}
	if !cfg.Notifications.TmuxBell {
		t.Error("expected Notifications.TmuxBell=true by default")
	}
	if cfg.Queue.MaxMessageLength != 100 {
		t.Errorf("expected MaxMessageLength=100, got %d", cfg.Queue.MaxMessageLength)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error loading missing config: %v", err)
	}
	if cfg.Queue.MaxMessageLength != 100 {
		t.Error("expected defaults when config file is absent")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := config.DefaultConfig()
	cfg.Notifications.Macos = false
	cfg.Queue.MaxMessageLength = 50

	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Notifications.Macos != false {
		t.Error("expected Macos=false after round-trip")
	}
	if loaded.Queue.MaxMessageLength != 50 {
		t.Errorf("expected MaxMessageLength=50, got %d", loaded.Queue.MaxMessageLength)
	}
}

func TestConfigPath_XDGOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := config.ConfigPath()
	expected := dir + "/hive/config.yaml"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestConfigPath_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := os.Getenv("HOME")
	path := config.ConfigPath()
	expected := home + "/.config/hive/config.yaml"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}
```

- [ ] **Step 5: Run config tests**

```bash
cd /Users/kikocastillo/coding/personal/hive
go test ./internal/config/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 6: Create `internal/db/db.go`**

```go
// internal/db/db.go
package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const sqliteTimeLayout = "2006-01-02T15:04:05Z"

func DBPath() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(base, "hive", "hive.db")
}

func Open() (*sql.DB, error) {
	path := DBPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := migrate(database); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

func OpenMem() (*sql.DB, error) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	if err := migrate(database); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

func migrate(database *sql.DB) error {
	if _, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS queue (
			id         INTEGER PRIMARY KEY,
			session    TEXT NOT NULL UNIQUE,
			message    TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS history (
			id          INTEGER PRIMARY KEY,
			session     TEXT NOT NULL,
			message     TEXT,
			notified_at DATETIME,
			resolved_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT
		);
	`); err != nil {
		return err
	}
	// Add columns if missing (idempotent — ignore duplicate column errors)
	_, _ = database.Exec(`ALTER TABLE queue ADD COLUMN pane TEXT`)
	_, _ = database.Exec(`ALTER TABLE queue ADD COLUMN snoozed_until TEXT`)
	return nil
}
```

- [ ] **Step 7: Create `internal/db/history.go`**

```go
// internal/db/history.go
package db

import (
	"database/sql"
	"fmt"
	"time"
)

type HistoryEntry struct {
	ID         int64
	Session    string
	Message    string
	NotifiedAt time.Time
	ResolvedAt time.Time
}

func MoveToHistory(database *sql.DB, session string) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		sess      string
		message   string
		createdAt string
	)
	err = tx.QueryRow(
		`SELECT session, COALESCE(message,''), created_at FROM queue WHERE session = ?`, session,
	).Scan(&sess, &message, &createdAt)
	if err == sql.ErrNoRows {
		return nil // idempotent — not in queue is not an error
	}
	if err != nil {
		return err
	}

	if _, err = tx.Exec(
		`INSERT INTO history (session, message, notified_at) VALUES (?, ?, ?)`,
		sess, message, createdAt,
	); err != nil {
		return err
	}

	if _, err = tx.Exec(`DELETE FROM queue WHERE session = ?`, session); err != nil {
		return err
	}

	return tx.Commit()
}

func PurgeHistory(database *sql.DB, retentionDays int) error {
	_, err := database.Exec(
		`DELETE FROM history WHERE resolved_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", retentionDays),
	)
	return err
}

func ListHistory(database *sql.DB) ([]HistoryEntry, error) {
	rows, err := database.Query(`
		SELECT id, session, COALESCE(message,''), COALESCE(notified_at, resolved_at), resolved_at
		FROM history ORDER BY resolved_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var (
			e          HistoryEntry
			notifiedAt string
			resolvedAt string
		)
		if err := rows.Scan(&e.ID, &e.Session, &e.Message, &notifiedAt, &resolvedAt); err != nil {
			return nil, err
		}
		e.NotifiedAt, _ = time.Parse(sqliteTimeLayout, notifiedAt)
		e.ResolvedAt, _ = time.Parse(sqliteTimeLayout, resolvedAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
```

- [ ] **Step 8: Create `internal/db/queue.go`**

```go
// internal/db/queue.go
package db

import (
	"database/sql"
	"time"
)

type Entry struct {
	ID        int64
	Session   string
	Message   string
	Pane      string // tmux pane ID (e.g. %23), empty if unknown
	CreatedAt time.Time
}

// Target returns the most specific tmux switch-client target available.
func (e Entry) Target() string {
	if e.Pane != "" {
		return e.Pane
	}
	return e.Session
}

func Enqueue(database *sql.DB, session, message, pane string) error {
	_, err := database.Exec(`
		INSERT INTO queue (session, message, pane, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(session) DO UPDATE SET
			message    = excluded.message,
			pane       = excluded.pane,
			created_at = excluded.created_at
	`, session, message, pane)
	return err
}

const activeFilter = `(snoozed_until IS NULL OR snoozed_until <= datetime('now'))`

func Snooze(database *sql.DB, session string, until time.Time) error {
	_, err := database.Exec(
		`UPDATE queue SET snoozed_until = ? WHERE session = ?`,
		until.UTC().Format(sqliteTimeLayout), session,
	)
	return err
}

func Next(database *sql.DB) (*Entry, error) {
	row := database.QueryRow(
		`SELECT id, session, COALESCE(message,''), COALESCE(pane,''), created_at FROM queue WHERE ` + activeFilter + ` ORDER BY created_at ASC LIMIT 1`,
	)
	return scanEntry(row)
}

func List(database *sql.DB) ([]Entry, error) {
	rows, err := database.Query(
		`SELECT id, session, COALESCE(message,''), COALESCE(pane,''), created_at FROM queue WHERE ` + activeFilter + ` ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var (
			e         Entry
			createdAt string
		)
		if err := rows.Scan(&e.ID, &e.Session, &e.Message, &e.Pane, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(sqliteTimeLayout, createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func scanEntry(row *sql.Row) (*Entry, error) {
	var (
		e         Entry
		createdAt string
	)
	err := row.Scan(&e.ID, &e.Session, &e.Message, &e.Pane, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.CreatedAt, _ = time.Parse(sqliteTimeLayout, createdAt)
	return &e, nil
}
```

- [ ] **Step 9: Create `internal/db/settings.go`**

```go
// internal/db/settings.go
package db

import (
	"database/sql"
	"time"
)

const pausedKey = "paused_until"

// indefinite sentinel: far-future date meaning "paused forever"
const indefinite = "9999-12-31T23:59:59Z"

func IsPaused(database *sql.DB) bool {
	var value string
	err := database.QueryRow(`SELECT value FROM settings WHERE key = ?`, pausedKey).Scan(&value)
	if err != nil {
		return false
	}
	if value == indefinite {
		return true
	}
	until, err := time.Parse(sqliteTimeLayout, value)
	if err != nil {
		return false
	}
	return time.Now().Before(until)
}

// SetPaused pauses notifications. Pass a zero time for indefinite.
func SetPaused(database *sql.DB, until time.Time) error {
	value := indefinite
	if !until.IsZero() {
		value = until.UTC().Format(sqliteTimeLayout)
	}
	_, err := database.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		pausedKey, value,
	)
	return err
}

func ClearPaused(database *sql.DB) error {
	_, err := database.Exec(`DELETE FROM settings WHERE key = ?`, pausedKey)
	return err
}
```

- [ ] **Step 10: Create `internal/db/db_test.go`**

Note: the chezmoi source has a bug — `Enqueue` calls pass only 3 args but the function takes 4 (session, message, pane). Fixed below with an empty pane `""`.

```go
// internal/db/db_test.go
package db_test

import (
	"testing"

	"github.com/fjcasti1/hive/internal/db"
)

func TestOpenMem(t *testing.T) {
	database, err := db.OpenMem()
	if err != nil {
		t.Fatalf("OpenMem failed: %v", err)
	}
	defer database.Close()

	for _, table := range []string{"queue", "history"} {
		var name string
		err := database.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", table, err)
		}
	}
}

func TestEnqueue(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	if err := db.Enqueue(database, "my-session", "hello", ""); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	entries, err := db.List(database)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Session != "my-session" {
		t.Errorf("expected session=my-session, got %s", entries[0].Session)
	}
	if entries[0].Message != "hello" {
		t.Errorf("expected message=hello, got %s", entries[0].Message)
	}
}

func TestEnqueueUpsert(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	db.Enqueue(database, "my-session", "first", "")
	db.Enqueue(database, "my-session", "second", "")

	entries, _ := db.List(database)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after upsert, got %d", len(entries))
	}
	if entries[0].Message != "second" {
		t.Errorf("expected message=second after upsert, got %s", entries[0].Message)
	}
}

func TestNextFIFO(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	db.Enqueue(database, "session-a", "first", "")
	db.Enqueue(database, "session-b", "second", "")

	entry, err := db.Next(database)
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.Session != "session-a" {
		t.Errorf("expected FIFO first=session-a, got %s", entry.Session)
	}
}

func TestNextEmpty(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	entry, err := db.Next(database)
	if err != nil {
		t.Fatalf("unexpected error on empty queue: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil for empty queue, got %+v", entry)
	}
}

func TestMoveToHistory(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	db.Enqueue(database, "done-session", "finished", "")
	if err := db.MoveToHistory(database, "done-session"); err != nil {
		t.Fatalf("MoveToHistory failed: %v", err)
	}

	entries, _ := db.List(database)
	if len(entries) != 0 {
		t.Errorf("expected empty queue after MoveToHistory, got %d entries", len(entries))
	}

	history, err := db.ListHistory(database)
	if err != nil {
		t.Fatalf("ListHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	if history[0].Session != "done-session" {
		t.Errorf("expected session=done-session, got %s", history[0].Session)
	}
	if history[0].Message != "finished" {
		t.Errorf("expected message=finished, got %s", history[0].Message)
	}
}

func TestMoveToHistoryIdempotent(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	if err := db.MoveToHistory(database, "nonexistent"); err != nil {
		t.Fatalf("MoveToHistory on missing session should not error: %v", err)
	}
}

func TestListHistoryOrder(t *testing.T) {
	database, _ := db.OpenMem()
	defer database.Close()

	db.Enqueue(database, "session-a", "first", "")
	db.Enqueue(database, "session-b", "second", "")
	db.MoveToHistory(database, "session-a")
	db.MoveToHistory(database, "session-b")

	history, _ := db.ListHistory(database)
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	if history[0].Session != "session-b" {
		t.Errorf("expected session-b first (most recent resolved), got %s", history[0].Session)
	}
}
```

- [ ] **Step 11: Run db tests**

```bash
cd /Users/kikocastillo/coding/personal/hive
go test ./internal/db/... -v
```

Expected: all tests PASS.

- [ ] **Step 12: Create `internal/notify/notify.go`**

```go
// internal/notify/notify.go
package notify

import (
	"fmt"
	"os"

	"github.com/fjcasti1/hive/internal/config"
)

type Channel interface {
	Send(session, message string) error
}

// Send fires all enabled notification channels with default implementations.
// Errors are printed to stderr but do not abort — the queue is the source of truth.
func Send(cfg config.Config, session, message string) {
	SendChannels(cfg, []Channel{MacOS{}}, []Channel{TmuxBell{}}, session, message)
}

// SendChannels is the testable core — accepts injectable channel slices.
func SendChannels(cfg config.Config, macosChannels, tmuxChannels []Channel, session, message string) {
	if cfg.Notifications.Macos {
		for _, ch := range macosChannels {
			if err := ch.Send(session, message); err != nil {
				fmt.Fprintf(os.Stderr, "hive: macos notification failed: %v\n", err)
			}
		}
	}
	if cfg.Notifications.TmuxBell {
		for _, ch := range tmuxChannels {
			if err := ch.Send(session, message); err != nil {
				fmt.Fprintf(os.Stderr, "hive: tmux bell failed: %v\n", err)
			}
		}
	}
}
```

- [ ] **Step 13: Create `internal/notify/macos.go`**

```go
// internal/notify/macos.go
package notify

import (
	"fmt"
	"os/exec"
)

type MacOS struct{}

func (MacOS) Send(session, message string) error {
	title := fmt.Sprintf("hive: %s", session)
	body := message
	if body == "" {
		body = "Agent needs your attention"
	}
	script := fmt.Sprintf(`display notification %q with title %q sound name "Ping"`, body, title)
	return exec.Command("osascript", "-e", script).Run()
}
```

- [ ] **Step 14: Create `internal/notify/tmux.go`**

```go
// internal/notify/tmux.go
package notify

import (
	"os"
	"os/exec"
	"strings"
)

type TmuxBell struct{}

// Send writes a BEL byte directly to the first pane's tty of the target session.
func (TmuxBell) Send(session, _ string) error {
	out, err := exec.Command("tmux", "list-panes", "-t", session, "-F", "#{pane_tty}").Output()
	if err != nil {
		return err
	}
	tty := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	f, err := os.OpenFile(tty, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte{0x07}) // BEL
	return err
}
```

- [ ] **Step 15: Create `internal/notify/notify_test.go`**

```go
// internal/notify/notify_test.go
package notify_test

import (
	"errors"
	"testing"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/notify"
)

type mockChannel struct {
	called bool
	err    error
}

func (m *mockChannel) Send(session, message string) error {
	m.called = true
	return m.err
}

func TestSendChannels_BothDisabled(t *testing.T) {
	macos := &mockChannel{}
	bell := &mockChannel{}
	cfg := config.Config{
		Notifications: config.Notifications{Macos: false, TmuxBell: false},
	}
	notify.SendChannels(cfg, []notify.Channel{macos}, []notify.Channel{bell}, "s", "m")
	if macos.called {
		t.Error("macOS channel should not fire when disabled")
	}
	if bell.called {
		t.Error("tmux bell channel should not fire when disabled")
	}
}

func TestSendChannels_BothEnabled(t *testing.T) {
	macos := &mockChannel{}
	bell := &mockChannel{}
	cfg := config.Config{
		Notifications: config.Notifications{Macos: true, TmuxBell: true},
	}
	notify.SendChannels(cfg, []notify.Channel{macos}, []notify.Channel{bell}, "s", "m")
	if !macos.called {
		t.Error("macOS channel should fire when enabled")
	}
	if !bell.called {
		t.Error("tmux bell channel should fire when enabled")
	}
}

func TestSendChannels_ErrorsNonFatal(t *testing.T) {
	failing := &mockChannel{err: errors.New("osascript not found")}
	cfg := config.Config{
		Notifications: config.Notifications{Macos: true, TmuxBell: false},
	}
	// Should not panic — notification failures are soft
	notify.SendChannels(cfg, []notify.Channel{failing}, nil, "s", "m")
}
```

- [ ] **Step 16: Run notify tests**

```bash
cd /Users/kikocastillo/coding/personal/hive
go test ./internal/notify/... -v
```

Expected: all 3 tests PASS.

- [ ] **Step 17: Create `internal/session/session.go`**

```go
// internal/session/session.go
package session

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

var ErrNotInTmux = errors.New("not running inside a tmux session; use --session to specify one")

func CurrentSession() (string, error) {
	if os.Getenv("TMUX") == "" {
		return "", ErrNotInTmux
	}
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
```

- [ ] **Step 18: Create `internal/session/session_test.go`**

```go
// internal/session/session_test.go
package session_test

import (
	"errors"
	"testing"

	"github.com/fjcasti1/hive/internal/session"
)

func TestCurrentSessionNotInTmux(t *testing.T) {
	t.Setenv("TMUX", "")
	_, err := session.CurrentSession()
	if !errors.Is(err, session.ErrNotInTmux) {
		t.Errorf("expected ErrNotInTmux when TMUX unset, got %v", err)
	}
}
```

- [ ] **Step 19: Run session tests**

```bash
cd /Users/kikocastillo/coding/personal/hive
go test ./internal/session/... -v
```

Expected: 1 test PASS.

- [ ] **Step 20: Create `cmd/util.go`**

This file extracts `timeAgo` which is shared by history, list, and status commands.

```go
// cmd/util.go
package cmd

import (
	"fmt"
	"time"
)

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
```

- [ ] **Step 21: Create `cmd/history.go`**

```go
// cmd/history.go
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show resolved notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := db.ListHistory(database)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("no history yet")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  session\tmessage\tresolved")
		for _, e := range entries {
			msg := e.Message
			if msg == "" {
				msg = "-"
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\n", e.Session, msg, timeAgo(e.ResolvedAt))
		}
		return w.Flush()
	},
}
```

- [ ] **Step 22: Create `cmd/root.go`**

Note: `init()` only registers `historyCmd` for now. Each subsequent PR will add its command here.

```go
// cmd/root.go
package cmd

import (
	"database/sql"
	"os"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var (
	cfg      config.Config
	database *sql.DB
)

var rootCmd = &cobra.Command{
	Use:   "hive",
	Short: "Manage multiple agentic tmux sessions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
		database, err = db.Open()
		if err != nil {
			return err
		}
		_ = db.PurgeHistory(database, cfg.History.RetentionDays)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if database != nil {
			database.Close()
		}
	},
}

func SetVersion(v string) {
	rootCmd.Version = v
	rootCmd.SetVersionTemplate("hive {{.Version}}\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		historyCmd,
	)
}
```

- [ ] **Step 23: Rewrite `main.go`**

```go
// main.go
package main

import "github.com/fjcasti1/hive/cmd"

// Version is set at build time via:
//
//	go build -ldflags "-X main.Version=1.0.0"
var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
```

- [ ] **Step 24: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . --version
go run . history
```

Expected:
- `go build ./...` exits 0 with no errors
- `go run . --version` prints `hive dev`
- `go run . history` prints `no history yet` (or a table if the local DB has entries)

- [ ] **Step 25: Create `DEVELOPMENT.md`**

```markdown
# Development

## Prerequisites
- Go 1.26+
- tmux
- macOS (for notify features)

## Quick Start

```
make build          # compile to ./hive
make run            # go run . (fast iteration, no compile)
make install-dev    # install to ~/.local/bin/hive (shadows Homebrew)
```

## Dev vs Prod

- `make run` / `go run .` — fastest iteration, no binary
- `make install-dev` — installs dev binary at `~/.local/bin/hive`, shadows brew-installed hive
- `make uninstall-dev` — removes dev binary, falls back to Homebrew version
- `brew install fjcasti1/hive/hive` — prod install

## Testing

```
make test           # go test ./...
make vet            # go vet ./...
make check          # vet + test together (what CI runs)
```

## Release Process

Releases are fully automated. To ship:

1. Open a PR with conventional commits (`feat:`, `fix:`, `chore:`)
2. Merge to main → release-please opens a release PR
3. Merge the release PR → GoReleaser tags and publishes binaries + Homebrew formula

## Project Layout

```
cmd/          # cobra subcommands, one file per command
internal/
  config/     # config file management
  db/         # sqlite queue, history, settings
  notify/     # macOS and tmux notifications
  session/    # tmux session management
```
```

- [ ] **Step 26: Create `Makefile`**

```makefile
.PHONY: build run install-dev uninstall-dev test vet check clean

build:
	go build -o hive .

run:
	go run .

install-dev:
	go build -o ~/.local/bin/hive .

uninstall-dev:
	rm -f ~/.local/bin/hive

test:
	go test ./...

vet:
	go vet ./...

check: vet test

clean:
	rm -f hive
```

- [ ] **Step 27: Run full test suite**

```bash
cd /Users/kikocastillo/coding/personal/hive
make check
```

Expected: `go vet ./...` exits 0, all tests PASS.

- [ ] **Step 28: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add internal/ cmd/ main.go go.mod go.sum DEVELOPMENT.md Makefile
git commit -m "feat: add internal packages, cmd/ structure, history command, DEVELOPMENT.md, Makefile"
git push -u origin feat/foundation-packages-history
gh pr create --title "feat: foundation packages, cmd/ restructure, history command" --body "$(cat <<'EOF'
## Summary
- Ports all four internal packages (config, db, notify, session) from chezmoi source
- Restructures main.go into cmd/ package with root.go, util.go, history.go
- Adds DEVELOPMENT.md and Makefile
- First real command: `hive history`

## Test plan
- [ ] `make check` passes (go vet + go test ./...)
- [ ] `go run . --version` prints `hive dev`
- [ ] `go run . history` runs without error
EOF
)"
```

---

## Task 2: notify command

**Files:**
- Create: `cmd/notify.go`
- Modify: `cmd/root.go` (add notifyCmd to init)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-notify
```

- [ ] **Step 2: Create `cmd/notify.go`**

```go
// cmd/notify.go
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/notify"
	"github.com/fjcasti1/hive/internal/session"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Add current session to the waiting queue and fire notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		sessionName, _ := cmd.Flags().GetString("session")

		if sessionName == "" {
			var err error
			sessionName, err = session.CurrentSession()
			if errors.Is(err, session.ErrNotInTmux) {
				return fmt.Errorf("not in a tmux session; use --session <name>")
			}
			if err != nil {
				return err
			}
		}

		if len(msg) > cfg.Queue.MaxMessageLength {
			msg = msg[:cfg.Queue.MaxMessageLength]
		}

		pane := os.Getenv("TMUX_PANE")
		if err := db.Enqueue(database, sessionName, msg, pane); err != nil {
			return fmt.Errorf("queue error: %w", err)
		}

		if !db.IsPaused(database) {
			notify.Send(cfg, sessionName, msg)
		}
		return nil
	},
}

func init() {
	notifyCmd.Flags().StringP("message", "m", "", "Why the agent needs attention (max 100 chars)")
	notifyCmd.Flags().StringP("session", "s", "", "tmux session name (auto-detected if omitted)")
}
```

- [ ] **Step 3: Register notifyCmd in `cmd/root.go`**

In `cmd/root.go`, find the `init()` function and add `notifyCmd`:

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . notify --help
go run . notify --session test-session --message "hello"
go run . history
```

Expected:
- `go build ./...` exits 0
- `notify --help` shows usage with `--message` and `--session` flags
- `notify --session test-session --message "hello"` enqueues the session (no tmux required due to `--session` flag); macOS notification fires if enabled
- `history` now shows no entries (notify enqueues, not histories)

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/notify.go cmd/root.go
git commit -m "feat: add notify command"
git push -u origin feat/cmd-notify
gh pr create --title "feat: add notify command" --body "$(cat <<'EOF'
## Summary
- Adds `hive notify` — enqueues current session and fires macOS/tmux notifications
- Respects `--session` flag and `IsPaused` state

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive notify --session smoke-test -m "hello"` exits 0, fires macOS notification
- [ ] `hive list` shows smoke-test in queue
EOF
)"
```

---

## Task 3: list command

**Files:**
- Create: `cmd/list.go`
- Modify: `cmd/root.go` (add listCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-list
```

- [ ] **Step 2: Create `cmd/list.go`**

```go
// cmd/list.go
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

const clearScreen = "\033[H\033[2J"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all waiting sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		namesOnly, _ := cmd.Flags().GetBool("names")
		watch, _ := cmd.Flags().GetBool("watch")
		interval := cfg.List.WatchInterval
		if i, _ := cmd.Flags().GetInt("interval"); i > 0 {
			interval = i
		}

		for {
			entries, err := db.List(database)
			if err != nil {
				return err
			}
			if namesOnly {
				if watch {
					fmt.Print(clearScreen)
				}
				for _, e := range entries {
					fmt.Printf("%s\t%s\n", e.Target(), e.Session)
				}
			} else {
				if watch {
					fmt.Print(clearScreen)
					fmt.Printf("hive watch  (every %ds)  %s\n\n", interval, time.Now().Format("15:04:05"))
				}
				if len(entries) == 0 {
					fmt.Println("no sessions waiting")
				} else {
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
					fmt.Fprintln(w, "  #\tsession\tpane\tmessage\twaiting")
					for i, e := range entries {
						msg := e.Message
						if msg == "" {
							msg = "-"
						}
						pane := e.Pane
						if pane == "" {
							pane = "-"
						}
						fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n", i+1, e.Session, pane, msg, timeAgo(e.CreatedAt))
					}
					w.Flush()
				}
			}
			if !watch {
				break
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("names", false, "print only session names, one per line")
	listCmd.Flags().BoolP("watch", "w", false, "refresh the list on an interval (see list.watch_interval config)")
	listCmd.Flags().Int("interval", 0, "override watch interval in seconds")
}
```

- [ ] **Step 3: Register listCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . list
go run . list --names
```

Expected:
- `go build ./...` exits 0
- `list` shows a table of waiting sessions (or `no sessions waiting`)
- `list --names` prints bare session names one per line

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/list.go cmd/root.go
git commit -m "feat: add list command"
git push -u origin feat/cmd-list
gh pr create --title "feat: add list command" --body "$(cat <<'EOF'
## Summary
- Adds `hive list` — shows all waiting sessions in a table
- Supports `--names` (bare output for scripting) and `--watch` (auto-refresh)

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive list` shows table or "no sessions waiting"
- [ ] `hive list --names` shows bare session names
EOF
)"
```

---

## Task 4: next command

**Files:**
- Create: `cmd/next.go`
- Modify: `cmd/root.go` (add nextCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-next
```

- [ ] **Step 2: Create `cmd/next.go`**

```go
// cmd/next.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next waiting session (FIFO)",
	RunE: func(cmd *cobra.Command, args []string) error {
		printOnly, _ := cmd.Flags().GetBool("print")

		entry, err := db.Next(database)
		if err != nil {
			return err
		}
		if entry == nil {
			fmt.Fprintln(os.Stderr, "hive: no sessions waiting")
			return nil
		}

		if printOnly {
			fmt.Println(entry.Session)
			return nil
		}

		return exec.Command("tmux", "switch-client", "-t", entry.Target()).Run()
	},
}

func init() {
	nextCmd.Flags().BoolP("print", "p", false, "Print session name instead of switching")
}
```

- [ ] **Step 3: Register nextCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . next --print
```

Expected:
- `go build ./...` exits 0
- `next --print` prints the next session name (or prints nothing to stderr if queue is empty)

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/next.go cmd/root.go
git commit -m "feat: add next command"
git push -u origin feat/cmd-next
gh pr create --title "feat: add next command" --body "$(cat <<'EOF'
## Summary
- Adds `hive next` — switches to the next waiting tmux session (FIFO)
- `--print` flag prints the session name instead of switching (useful for scripting)

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive next --print` prints session name or exits cleanly with nothing waiting
EOF
)"
```

---

## Task 5: ack command

**Files:**
- Create: `cmd/ack.go`
- Modify: `cmd/root.go` (add ackCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-ack
```

- [ ] **Step 2: Create `cmd/ack.go`**

```go
// cmd/ack.go
package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/session"
	"github.com/spf13/cobra"
)

var ackCmd = &cobra.Command{
	Use:   "ack [session-or-index]",
	Short: "Acknowledge a session — mark feedback given and move to history",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName, _ := cmd.Flags().GetString("session")

		if len(args) > 0 {
			arg := args[0]
			if idx, err := strconv.Atoi(arg); err == nil {
				entries, err := db.List(database)
				if err != nil {
					return err
				}
				if idx < 1 || idx > len(entries) {
					return fmt.Errorf("index %d out of range (1-%d)", idx, len(entries))
				}
				sessionName = entries[idx-1].Session
			} else {
				sessionName = arg
			}
		}

		if sessionName == "" {
			var err error
			sessionName, err = session.CurrentSession()
			if errors.Is(err, session.ErrNotInTmux) {
				return fmt.Errorf("not in a tmux session; use --session <name>")
			}
			if err != nil {
				return err
			}
		}

		return db.MoveToHistory(database, sessionName)
	},
}

func init() {
	ackCmd.Flags().StringP("session", "s", "", "tmux session name (auto-detected if omitted)")
}
```

- [ ] **Step 3: Register ackCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . notify --session ack-test -m "needs ack"
go run . list
go run . ack ack-test
go run . list
go run . history
```

Expected: after `ack ack-test`, list shows empty queue and history shows the acked session.

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/ack.go cmd/root.go
git commit -m "feat: add ack command"
git push -u origin feat/cmd-ack
gh pr create --title "feat: add ack command" --body "$(cat <<'EOF'
## Summary
- Adds `hive ack` — moves a session from the queue to history
- Accepts session name, list index, or auto-detects from TMUX env

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive notify --session s -m "test"` + `hive ack s` moves it to history
- [ ] `hive ack 1` acks by index
EOF
)"
```

---

## Task 6: snooze command

**Files:**
- Create: `cmd/snooze.go`
- Modify: `cmd/root.go` (add snoozeCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-snooze
```

- [ ] **Step 2: Create `cmd/snooze.go`**

```go
// cmd/snooze.go
package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/session"
	"github.com/spf13/cobra"
)

var snoozeCmd = &cobra.Command{
	Use:   "snooze [session-or-index] [duration]",
	Short: "Hide a session from the queue for a duration (e.g. 10m, 1h)",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName, durStr := "", cfg.Snooze.DefaultDuration

		switch len(args) {
		case 2:
			sessionName, durStr = args[0], args[1]
		case 1:
			durStr = args[0]
		}

		// Resolve session name or index
		if sessionName != "" {
			if idx, err := strconv.Atoi(sessionName); err == nil {
				entries, err := db.List(database)
				if err != nil {
					return err
				}
				if idx < 1 || idx > len(entries) {
					return fmt.Errorf("index %d out of range (1-%d)", idx, len(entries))
				}
				sessionName = entries[idx-1].Session
			}
		} else {
			var err error
			sessionName, err = session.CurrentSession()
			if errors.Is(err, session.ErrNotInTmux) {
				return fmt.Errorf("not in a tmux session; use: hive snooze [session] <duration>")
			}
			if err != nil {
				return err
			}
		}

		d, err := time.ParseDuration(durStr)
		if err != nil {
			return fmt.Errorf("invalid duration %q — use Go format e.g. 10m, 1h, 30s", durStr)
		}

		if err := db.Snooze(database, sessionName, time.Now().Add(d)); err != nil {
			return err
		}
		fmt.Printf("snoozed %s for %s\n", sessionName, durStr)
		return nil
	},
}
```

- [ ] **Step 3: Register snoozeCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . notify --session snooze-test -m "snooze me"
go run . list
go run . snooze snooze-test 30s
go run . list
```

Expected: after snooze, `list` shows empty queue (session hidden for 30s).

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/snooze.go cmd/root.go
git commit -m "feat: add snooze command"
git push -u origin feat/cmd-snooze
gh pr create --title "feat: add snooze command" --body "$(cat <<'EOF'
## Summary
- Adds `hive snooze` — hides a session from the queue for a duration
- Accepts session name, list index, or auto-detects; duration defaults to config value

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive notify --session s -m "t"` + `hive snooze s 30s` hides it; `hive list` shows empty
EOF
)"
```

---

## Task 7: peek command

**Files:**
- Create: `cmd/peek.go`
- Modify: `cmd/root.go` (add peekCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-peek
```

- [ ] **Step 2: Create `cmd/peek.go`**

```go
// cmd/peek.go
package cmd

import (
	"fmt"
	"os"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var peekCmd = &cobra.Command{
	Use:   "peek",
	Short: "Show the next waiting session without switching",
	RunE: func(cmd *cobra.Command, args []string) error {
		entry, err := db.Next(database)
		if err != nil {
			return err
		}
		if entry == nil {
			fmt.Fprintln(os.Stderr, "hive: no sessions waiting")
			return nil
		}
		if entry.Message != "" {
			fmt.Printf("%s — %s\n", entry.Session, entry.Message)
		} else {
			fmt.Println(entry.Session)
		}
		return nil
	},
}
```

- [ ] **Step 3: Register peekCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . notify --session peek-test -m "hello peek"
go run . peek
```

Expected: prints `peek-test — hello peek`.

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/peek.go cmd/root.go
git commit -m "feat: add peek command"
git push -u origin feat/cmd-peek
gh pr create --title "feat: add peek command" --body "$(cat <<'EOF'
## Summary
- Adds `hive peek` — shows the next waiting session without switching to it

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive notify --session s -m "msg"` + `hive peek` prints `s — msg`
EOF
)"
```

---

## Task 8: status command

**Files:**
- Create: `cmd/status.go`
- Modify: `cmd/root.go` (add statusCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-status
```

- [ ] **Step 2: Create `cmd/status.go`**

```go
// cmd/status.go
package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print compact status for tmux status bar",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, _ = config.Load()
		database, _ = db.Open()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if database == nil {
			return nil
		}
		if db.IsPaused(database) {
			entries, _ := db.List(database)
			if len(entries) == 0 {
				fmt.Printf("#[fg=colour245]⏸ paused#[fg=default] ")
			} else {
				fmt.Printf("#[fg=colour245]⏸ %d waiting#[fg=default] ", len(entries))
			}
			return nil
		}
		entries, err := db.List(database)
		if err != nil || len(entries) == 0 {
			return nil
		}
		e := entries[0]
		extra := len(entries) - 1
		label := renderStatusLabel(cfg.Status.Format, e.Session, e.Message, timeAgo(e.CreatedAt), extra)
		fmt.Printf("#[fg=colour220,bold]🐝 %s#[fg=default,nobold] ", label)
		return nil
	},
}

func renderStatusLabel(format, session, message, age string, extra int) string {
	if message == "" {
		format = strings.ReplaceAll(format, ": {message}", "")
		format = strings.ReplaceAll(format, "{message}: ", "")
		format = strings.ReplaceAll(format, " {message}", "")
		format = strings.ReplaceAll(format, "{message}", "")
	}
	if extra == 0 {
		format = strings.ReplaceAll(format, " | +{extra}", "")
		format = strings.ReplaceAll(format, "| +{extra}", "")
		format = strings.ReplaceAll(format, " +{extra}", "")
		format = strings.ReplaceAll(format, "{extra}", "0")
	}
	r := strings.NewReplacer(
		"{session}", session,
		"{message}", message,
		"{age}", age,
		"{extra}", strconv.Itoa(extra),
		"{count}", strconv.Itoa(extra+1),
	)
	return r.Replace(format)
}
```

- [ ] **Step 3: Register statusCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
		statusCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . notify --session status-test -m "checking status"
go run . status
```

Expected: prints tmux-formatted status line like `#[fg=colour220,bold]🐝 status-test: checking status (0s ago)#[fg=default,nobold] `.

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/status.go cmd/root.go
git commit -m "feat: add status command"
git push -u origin feat/cmd-status
gh pr create --title "feat: add status command" --body "$(cat <<'EOF'
## Summary
- Adds `hive status` — prints tmux-formatted compact status for the status bar
- Uses PersistentPreRun override so a broken DB doesn't abort the status render

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive status` prints nothing when queue is empty
- [ ] `hive notify --session s -m "m"` + `hive status` prints bee emoji with session info
EOF
)"
```

---

## Task 9: pause command

**Files:**
- Create: `cmd/pause.go`
- Modify: `cmd/root.go` (add pauseCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-pause
```

- [ ] **Step 2: Create `cmd/pause.go`**

```go
// cmd/pause.go
package cmd

import (
	"fmt"
	"time"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause [duration]",
	Short: "Suppress notifications (agents still queue). Pass a duration e.g. 2h to auto-resume.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		durStr := cfg.Pause.DefaultDuration
		if len(args) > 0 {
			durStr = args[0]
		}

		var until time.Time // zero = indefinite
		if durStr != "" {
			d, err := time.ParseDuration(durStr)
			if err != nil {
				return fmt.Errorf("invalid duration %q — use Go format e.g. 2h, 30m", durStr)
			}
			until = time.Now().Add(d)
		}

		if err := db.SetPaused(database, until); err != nil {
			return err
		}

		if until.IsZero() {
			fmt.Println("hive paused — run 'hive resume' to re-enable notifications")
		} else {
			fmt.Printf("hive paused until %s\n", until.Format("15:04:05"))
		}
		return nil
	},
}
```

- [ ] **Step 3: Register pauseCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
		statusCmd,
		pauseCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . pause
go run . status
go run . notify --session paused-test -m "while paused"
```

Expected:
- `pause` prints "hive paused — run 'hive resume' to re-enable notifications"
- `status` shows paused indicator
- `notify` enqueues but fires no macOS notification (because paused)

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/pause.go cmd/root.go
git commit -m "feat: add pause command"
git push -u origin feat/cmd-pause
gh pr create --title "feat: add pause command" --body "$(cat <<'EOF'
## Summary
- Adds `hive pause` — suppresses notifications while keeping the queue active
- Optional duration arg for timed auto-resume

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive pause` + `hive status` shows paused indicator
- [ ] `hive pause 1s` + wait 2s + `hive status` shows normal (auto-resumed)
EOF
)"
```

---

## Task 10: resume command

**Files:**
- Create: `cmd/resume.go`
- Modify: `cmd/root.go` (add resumeCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-resume
```

- [ ] **Step 2: Create `cmd/resume.go`**

```go
// cmd/resume.go
package cmd

import (
	"fmt"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Re-enable notifications after a pause",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.ClearPaused(database); err != nil {
			return err
		}
		fmt.Println("hive resumed — notifications re-enabled")
		return nil
	},
}
```

- [ ] **Step 3: Register resumeCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
		statusCmd,
		pauseCmd,
		resumeCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . pause
go run . status
go run . resume
go run . status
```

Expected: status shows paused after `pause`, then normal after `resume`.

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/resume.go cmd/root.go
git commit -m "feat: add resume command"
git push -u origin feat/cmd-resume
gh pr create --title "feat: add resume command" --body "$(cat <<'EOF'
## Summary
- Adds `hive resume` — clears the paused state and re-enables notifications

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive pause` + `hive resume` + `hive status` shows normal (not paused)
EOF
)"
```

---

## Task 11: config command

**Files:**
- Create: `cmd/config.go`
- Modify: `cmd/root.go` (add configCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-config
```

- [ ] **Step 2: Create `cmd/config.go`**

```go
// cmd/config.go
package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify hive configuration",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, _ = config.Load()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current configuration as YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value using dot-notation key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if err := applyConfigValue(&cfg, key, value); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("set %s = %s\n", key, value)
		return nil
	},
}

func applyConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "notifications.macos":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool for %s: %q (use true/false)", key, value)
		}
		cfg.Notifications.Macos = v
	case "notifications.tmux_bell":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool for %s: %q (use true/false)", key, value)
		}
		cfg.Notifications.TmuxBell = v
	case "queue.max_message_length":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int for %s: %q", key, value)
		}
		cfg.Queue.MaxMessageLength = v
	case "status.format":
		cfg.Status.Format = value
	case "list.watch_interval":
		v, err := strconv.Atoi(value)
		if err != nil || v < 1 {
			return fmt.Errorf("invalid int for %s: %q (must be >= 1)", key, value)
		}
		cfg.List.WatchInterval = v
	case "history.retention_days":
		v, err := strconv.Atoi(value)
		if err != nil || v < 1 {
			return fmt.Errorf("invalid int for %s: %q (must be >= 1)", key, value)
		}
		cfg.History.RetentionDays = v
	case "snooze.default_duration":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid duration for %s: %q (e.g. 10m, 1h, 30s)", key, value)
		}
		cfg.Snooze.DefaultDuration = value
	case "pause.default_duration":
		if value != "" {
			if _, err := time.ParseDuration(value); err != nil {
				return fmt.Errorf("invalid duration for %s: %q (e.g. 2h, 30m) or empty for indefinite", key, value)
			}
		}
		cfg.Pause.DefaultDuration = value
	default:
		return fmt.Errorf(
			"unknown config key %q\n\nValid keys:\n  notifications.macos\n  notifications.tmux_bell\n  queue.max_message_length\n  status.format\n  list.watch_interval\n  history.retention_days\n  snooze.default_duration\n  pause.default_duration",
			key,
		)
	}
	return nil
}

func init() {
	configCmd.AddCommand(configShowCmd, configSetCmd)
}
```

- [ ] **Step 3: Register configCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
		statusCmd,
		pauseCmd,
		resumeCmd,
		configCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . config show
go run . config set notifications.macos false
go run . config show
go run . config set notifications.macos true
```

Expected: `config show` prints YAML; `config set` updates the value and confirms with `set notifications.macos = false`.

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/config.go cmd/root.go
git commit -m "feat: add config command"
git push -u origin feat/cmd-config
gh pr create --title "feat: add config command" --body "$(cat <<'EOF'
## Summary
- Adds `hive config show` — prints current config as YAML
- Adds `hive config set <key> <value>` — sets a dot-notation config key

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive config show` prints YAML with defaults
- [ ] `hive config set notifications.macos false` + `hive config show` reflects the change
- [ ] `hive config set unknown.key val` prints a clear error with valid keys listed
EOF
)"
```

---

## Task 12: install command

**Files:**
- Create: `cmd/install.go`
- Modify: `cmd/root.go` (add installCmd)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-install
```

- [ ] **Step 2: Create `cmd/install.go`**

```go
// cmd/install.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up hive integrations (Claude hooks, tmux snippet, shell completion)",
	RunE: func(cmd *cobra.Command, args []string) error {
		doClaude, _ := cmd.Flags().GetBool("claude")
		doTmux, _ := cmd.Flags().GetBool("tmux")
		doCompletion, _ := cmd.Flags().GetBool("completion")
		shell, _ := cmd.Flags().GetString("shell")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			fmt.Println("dry-run: no changes will be written\n")
		}

		if doClaude {
			fmt.Println("Claude Code hooks (~/.claude/settings.json):")
			if err := installClaudeHooks(dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			}
			fmt.Println()
		}

		if doCompletion {
			if shell == "" {
				shell = filepath.Base(os.Getenv("SHELL"))
			}
			fmt.Printf("Shell completion (%s):\n", shell)
			if err := installShellCompletion(shell, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			}
			fmt.Println()
		}

		if doTmux {
			fmt.Println("tmux config — add to ~/.tmux.conf and reload:")
			fmt.Println(tmuxSnippet())
		}

		if !dryRun {
			fmt.Println("Done! Run 'hive doctor' to verify the setup.")
		}
		return nil
	},
}

func init() {
	installCmd.Flags().Bool("claude", true, "install Claude Code hooks")
	installCmd.Flags().Bool("tmux", true, "print tmux config snippet")
	installCmd.Flags().Bool("completion", true, "install shell completion")
	installCmd.Flags().String("shell", "", "target shell for completion (default: $SHELL)")
	installCmd.Flags().Bool("dry-run", false, "show what would change without writing")
}

// --- Claude hooks ---

var hiveNotifyHooks = []string{"Stop", "StopFailure", "Notification", "Elicitation"}
var hiveAckHooks = []string{"UserPromptSubmit", "PostToolUse", "ElicitationResult", "SessionEnd"}

func installClaudeHooks(dryRun bool) error {
	home := os.Getenv("HOME")
	settingsPath := filepath.Join(home, ".claude", "settings.json")

	raw := make(map[string]interface{})
	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("cannot parse %s: %w", settingsPath, err)
		}
	}

	hooks, _ := raw["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	for _, event := range hiveNotifyHooks {
		printHookStatus(event, "hive notify 2>/dev/null; exit 0",
			mergeHiveHook(hooks, event, "hive notify 2>/dev/null; exit 0"))
	}
	for _, event := range hiveAckHooks {
		printHookStatus(event, "hive ack 2>/dev/null; exit 0",
			mergeHiveHook(hooks, event, "hive ack 2>/dev/null; exit 0"))
	}

	if dryRun {
		return nil
	}

	raw["hooks"] = hooks
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}

// mergeHiveHook adds the hive hook command for an event if not already present.
// Returns true if added, false if it was already there.
func mergeHiveHook(hooks map[string]interface{}, event, command string) bool {
	entries, _ := hooks[event].([]interface{})
	for _, entry := range entries {
		m, _ := entry.(map[string]interface{})
		for _, h := range toSlice(m["hooks"]) {
			hm, _ := h.(map[string]interface{})
			if hm["command"] == command {
				return false
			}
		}
	}
	hooks[event] = append(entries, map[string]interface{}{
		"matcher": "",
		"hooks":   []interface{}{map[string]interface{}{"type": "command", "command": command}},
	})
	return true
}

func toSlice(v interface{}) []interface{} {
	s, _ := v.([]interface{})
	return s
}

func printHookStatus(event, command string, added bool) {
	if added {
		fmt.Printf("  + %-20s → %s\n", event, command)
	} else {
		fmt.Printf("  ✓ %-20s (already configured)\n", event)
	}
}

// --- Shell completion ---

func installShellCompletion(shell string, dryRun bool) error {
	home := os.Getenv("HOME")

	targets := map[string][]string{
		"zsh":  {filepath.Join(home, ".zsh", "completions", "_hive"), filepath.Join(home, ".local", "share", "zsh", "site-functions", "_hive")},
		"bash": {filepath.Join(home, ".bash_completion.d", "hive"), "/etc/bash_completion.d/hive"},
		"fish": {filepath.Join(home, ".config", "fish", "completions", "hive.fish")},
	}

	paths, ok := targets[shell]
	if !ok {
		return fmt.Errorf("unsupported shell %q (supported: zsh, bash, fish)", shell)
	}

	// Pick first writable parent dir
	target := paths[0]
	for _, p := range paths {
		if _, err := os.Stat(filepath.Dir(p)); err == nil {
			target = p
			break
		}
	}

	fmt.Printf("  → %s\n", target)
	if dryRun {
		return nil
	}

	out, err := exec.Command("hive", "completion", shell).Output()
	if err != nil {
		return fmt.Errorf("completion generation failed: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, out, 0644)
}

// --- tmux snippet ---

func tmuxSnippet() string {
	return strings.TrimRight(`
  # hive — agent queue integration
  set -g focus-events on
  set -g status-interval 5

  # Status bar — choose ONE of the two options below:

  # Option A: plain status-right (works with any theme)
  set -g status-right "#(hive status)  %H:%M"

  # Option B: jimeh/tmux-themepack slot override
  # set -gu @theme-status-right
  # set -g @themepack-status-right-area-left-format "#(hive status)  %H:%M"
  # set -g @theme-status-interval 5

  # Keybindings
  bind A run-shell 'hive next'
  bind a display-popup -E "hive list --names 2>/dev/null | fzf --delimiter='\\t' --with-nth=2 --prompt '🐝  ' --no-sort | cut -f1 | xargs -r tmux switch-client -t"

  # Auto-ack when pane closes
  set-hook -g pane-exited "run-shell 'hive ack --session \#{session_name} 2>/dev/null'"
`, "\n")
}
```

- [ ] **Step 3: Register installCmd in `cmd/root.go`**

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
		statusCmd,
		pauseCmd,
		resumeCmd,
		configCmd,
		installCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . install --dry-run
go run . install --tmux=false --completion=false --dry-run
```

Expected:
- `go build ./...` exits 0
- `install --dry-run` shows what would be written without making changes
- `install --tmux=false --completion=false --dry-run` shows only Claude hook changes

- [ ] **Step 5: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/install.go cmd/root.go
git commit -m "feat: add install command"
git push -u origin feat/cmd-install
gh pr create --title "feat: add install command" --body "$(cat <<'EOF'
## Summary
- Adds `hive install` — wires Claude Code hooks, prints tmux config snippet, installs shell completion
- Supports `--dry-run` to preview changes before writing

## Test plan
- [ ] `go build ./...` passes
- [ ] `hive install --dry-run` shows planned changes without writing
- [ ] `hive install --tmux=false --completion=false` updates ~/.claude/settings.json with hive hooks
EOF
)"
```

---

## Task 13: doctor command

**Files:**
- Create: `cmd/doctor.go`
- Modify: `cmd/root.go` (add doctorCmd — final list)

- [ ] **Step 1: Create feature branch**

```bash
cd /Users/kikocastillo/coding/personal/hive
git checkout main && git pull
git checkout -b feat/cmd-doctor
```

- [ ] **Step 2: Create `cmd/doctor.go`**

```go
// cmd/doctor.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check hive integrations and report any issues",
	// Override root PersistentPreRunE so a broken DB doesn't abort the check.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, _ = config.Load()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		failures := 0

		// --- Database ---
		section("Database")
		dbPath := db.DBPath()
		database, err := db.Open()
		if err != nil {
			fail("cannot open %s: %v", dbPath, err)
			failures++
		} else {
			entries, _ := db.List(database)
			history, _ := db.ListHistory(database)
			pass("%s (%d queued, %d history)", dbPath, len(entries), len(history))
			database.Close()
		}
		fmt.Println()

		// --- Config ---
		section("Config")
		cfgPath := config.ConfigPath()
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			warn("%s not found — using defaults (run 'hive config show' to inspect)", cfgPath)
		} else if err != nil {
			fail("cannot read %s: %v", cfgPath, err)
			failures++
		} else {
			pass("%s", cfgPath)
		}
		fmt.Println()

		// --- Claude Code hooks ---
		section("Claude Code hooks (~/.claude/settings.json)")
		claudePath := filepath.Join(os.Getenv("HOME"), ".claude", "settings.json")
		data, err := os.ReadFile(claudePath)
		if err != nil {
			fail("%s: %v — run 'hive install --tmux=false --completion=false'", claudePath, err)
			failures++
		} else {
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				fail("cannot parse %s: %v", claudePath, err)
				failures++
			} else {
				hooks, _ := raw["hooks"].(map[string]interface{})
				allEvents := append(hiveNotifyHooks, hiveAckHooks...)
				for _, event := range allEvents {
					if isHiveHookPresent(hooks, event) {
						pass("%s", event)
					} else {
						fail("%s missing — run 'hive install --tmux=false --completion=false'", event)
						failures++
					}
				}
			}
		}
		fmt.Println()

		// --- tmux ---
		section("tmux")
		if _, err := exec.LookPath("tmux"); err != nil {
			fail("tmux not found in PATH")
			failures++
		} else {
			sr, _ := tmuxOption("status-right")
			tf, _ := tmuxOption("@themepack-status-right-area-left-format")
			if strings.Contains(sr, "hive status") || strings.Contains(tf, "hive status") {
				pass("status bar includes hive status")
			} else {
				warn("status bar does not include hive status — add to ~/.tmux.conf (see 'hive install')")
			}

			if si, _ := tmuxOption("status-interval"); si != "" {
				n, _ := strconv.Atoi(si)
				if n > 0 && n <= 10 {
					pass("status-interval = %s", si)
				} else {
					warn("status-interval = %s — recommend ≤ 10 for timely updates", si)
				}
			}

			keys, _ := tmuxListKeys()
			if strings.Contains(keys, "hive next") {
				pass("keybinding: hive next")
			} else {
				warn("hive next keybinding not found — add to ~/.tmux.conf (see 'hive install')")
			}
			if strings.Contains(keys, "hive list") {
				pass("keybinding: hive list (popup)")
			} else {
				warn("hive list keybinding not found — add to ~/.tmux.conf (see 'hive install')")
			}

			hooks, _ := tmuxOption("pane-exited")
			if strings.Contains(hooks, "hive ack") {
				pass("pane-exited hook: hive ack")
			} else {
				warn("pane-exited hook not set — closing a pane won't auto-ack")
			}
		}
		fmt.Println()

		// --- Shell completion ---
		section("Shell completion")
		shell := filepath.Base(os.Getenv("SHELL"))
		completionPaths := map[string][]string{
			"zsh":  {filepath.Join(os.Getenv("HOME"), ".zsh", "completions", "_hive"), filepath.Join(os.Getenv("HOME"), ".local", "share", "zsh", "site-functions", "_hive")},
			"bash": {filepath.Join(os.Getenv("HOME"), ".bash_completion.d", "hive")},
			"fish": {filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions", "hive.fish")},
		}
		found := false
		for _, p := range completionPaths[shell] {
			if _, err := os.Stat(p); err == nil {
				pass("%s", p)
				found = true
				break
			}
		}
		if !found {
			warn("completion not installed for %s — run 'hive install --claude=false --tmux=false'", shell)
		}
		fmt.Println()

		// --- Notifications ---
		section("Notifications")
		if _, err := exec.LookPath("osascript"); err != nil {
			warn("osascript not found — macOS notifications unavailable")
		} else if cfg.Notifications.Macos {
			pass("macOS notifications enabled")
		} else {
			warn("macOS notifications disabled in config (notifications.macos = false)")
		}
		if cfg.Notifications.TmuxBell {
			pass("tmux bell enabled")
		} else {
			warn("tmux bell disabled in config (notifications.tmux_bell = false)")
		}
		fmt.Println()

		if failures > 0 {
			fmt.Printf("%d failure(s) found — run 'hive install' to fix\n", failures)
			os.Exit(1)
		} else {
			fmt.Println("All checks passed.")
		}
		return nil
	},
}

func section(name string) { fmt.Printf("=== %s\n", name) }
func pass(f string, a ...interface{})   { fmt.Printf("  ✓ "+f+"\n", a...) }
func warn(f string, a ...interface{})   { fmt.Printf("  ⚠ "+f+"\n", a...) }
func fail(f string, a ...interface{})   { fmt.Printf("  ✗ "+f+"\n", a...) }

func tmuxOption(key string) (string, error) {
	out, err := exec.Command("tmux", "show-option", "-gv", key).Output()
	return strings.TrimSpace(string(out)), err
}

func tmuxListKeys() (string, error) {
	out, err := exec.Command("tmux", "list-keys").Output()
	return string(out), err
}

func isHiveHookPresent(hooks map[string]interface{}, event string) bool {
	entries, _ := hooks[event].([]interface{})
	for _, entry := range entries {
		m, _ := entry.(map[string]interface{})
		for _, h := range toSlice(m["hooks"]) {
			hm, _ := h.(map[string]interface{})
			cmd, _ := hm["command"].(string)
			if strings.HasPrefix(cmd, "hive notify") || strings.HasPrefix(cmd, "hive ack") {
				return true
			}
		}
	}
	return false
}
```

- [ ] **Step 3: Register doctorCmd in `cmd/root.go`** — this is the final complete list

```go
func init() {
	rootCmd.AddCommand(
		historyCmd,
		notifyCmd,
		listCmd,
		nextCmd,
		ackCmd,
		snoozeCmd,
		peekCmd,
		statusCmd,
		pauseCmd,
		resumeCmd,
		configCmd,
		installCmd,
		doctorCmd,
	)
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
cd /Users/kikocastillo/coding/personal/hive
go build ./...
go run . doctor
```

Expected: doctor runs all checks and prints a section-by-section report. Warnings are fine; failures indicate something needs fixing (e.g., Claude hooks not installed yet).

- [ ] **Step 5: Run full test suite one final time**

```bash
cd /Users/kikocastillo/coding/personal/hive
make check
```

Expected: all tests PASS, go vet exits 0.

- [ ] **Step 6: Commit and push PR**

```bash
cd /Users/kikocastillo/coding/personal/hive
git add cmd/doctor.go cmd/root.go
git commit -m "feat: add doctor command"
git push -u origin feat/cmd-doctor
gh pr create --title "feat: add doctor command" --body "$(cat <<'EOF'
## Summary
- Adds `hive doctor` — runs all integration checks (DB, config, Claude hooks, tmux, shell completion, notifications)
- Exits 1 if any failures found; exits 0 if all checks pass (warnings are non-fatal)
- All 13 commands now ported — migration complete

## Test plan
- [ ] `go build ./...` passes
- [ ] `make check` passes (all tests green)
- [ ] `hive doctor` runs without panicking; all sections rendered
EOF
)"
```
