package db

import (
	"database/sql"
	"testing"
)

func openMem() (*sql.DB, error) {
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

func TestMigrateUp(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	for _, table := range []string{"queue", "goose_db_version"} {
		var name string
		err := database.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}

	var version int64
	if err := database.QueryRow(
		`SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1`,
	).Scan(&version); err != nil {
		t.Fatalf("read goose version: %v", err)
	}
	if version != 4 {
		t.Errorf("want schema version 4, got %d", version)
	}

	// Idempotency: running migrate again should be a no-op.
	if err := migrate(database); err != nil {
		t.Fatalf("re-run migrate: %v", err)
	}
}

func TestPeekEmpty(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	entry, err := Show(database)
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil entry on empty queue, got %+v", entry)
	}
}

func TestPeekSingleEntry(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	if err := Enqueue(database, "alpha", "alpha", "pane:%1", "msg-a"); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	entry, err := Show(database)
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if got, want := entry.Label, "alpha"; got != want {
		t.Errorf("Label = %q, want %q", got, want)
	}
	if got, want := entry.Message, "msg-a"; got != want {
		t.Errorf("Message = %q, want %q", got, want)
	}
	if got, want := entry.Target(), "%1"; got != want {
		t.Errorf("Target = %q, want %q", got, want)
	}
}

func TestPeekTiebreaksOnID(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	// Two rows with the SAME created_at — only the id distinguishes them.
	// Without an id tiebreaker the ordering would be non-deterministic.
	const sameTime = "2026-01-01 10:00:00"
	if _, err := database.Exec(
		`INSERT INTO queue (agent_id, label, locator, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		"alpha", "alpha", "pane:%1", "first", sameTime,
	); err != nil {
		t.Fatalf("insert alpha: %v", err)
	}
	if _, err := database.Exec(
		`INSERT INTO queue (agent_id, label, locator, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		"beta", "beta", "pane:%2", "second", sameTime,
	); err != nil {
		t.Fatalf("insert beta: %v", err)
	}

	entry, err := Show(database)
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if got, want := entry.Label, "alpha"; got != want {
		t.Errorf("Label = %q, want %q (lower id wins on tie)", got, want)
	}
}

func TestPurgeHistory_EmptyTable(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	if err := PurgeHistory(database, 7); err != nil {
		t.Errorf("PurgeHistory on empty table: %v", err)
	}
}

func TestPurgeHistory_DeletesRowsOlderThanCutoff(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	// Two rows: one old (10 days ago), one fresh (now).
	insert := `INSERT INTO history (session, message, notified_at, resolved_at)
		VALUES (?, ?, ?, ?)`
	if _, err := database.Exec(insert,
		"old", "stale", "2025-01-01 00:00:00", "2025-01-01 00:00:00",
	); err != nil {
		t.Fatalf("insert old: %v", err)
	}
	if _, err := database.Exec(insert,
		"fresh", "recent", "2099-01-01 00:00:00", "2099-01-01 00:00:00",
	); err != nil {
		t.Fatalf("insert fresh: %v", err)
	}

	if err := PurgeHistory(database, 7); err != nil {
		t.Fatalf("PurgeHistory: %v", err)
	}

	var sessions []string
	rows, err := database.Query(`SELECT session FROM history ORDER BY session`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("scan: %v", err)
		}
		sessions = append(sessions, s)
	}
	if len(sessions) != 1 || sessions[0] != "fresh" {
		t.Errorf("after purge expected only [fresh], got %v", sessions)
	}
}

func TestPurgeHistory_ZeroWipesEverything(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	insert := `INSERT INTO history (session, message, notified_at, resolved_at)
		VALUES (?, ?, ?, ?)`
	for _, sess := range []string{"a", "b", "c"} {
		if _, err := database.Exec(insert,
			sess, "msg", "2025-01-01 00:00:00", "2025-01-01 00:00:00",
		); err != nil {
			t.Fatalf("insert %s: %v", sess, err)
		}
	}

	if err := PurgeHistory(database, 0); err != nil {
		t.Fatalf("PurgeHistory(0): %v", err)
	}

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM history`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("retention_days=0 should wipe all rows, got %d remaining", count)
	}
}

func TestPeekReturnsOldest(t *testing.T) {
	database, err := openMem()
	if err != nil {
		t.Fatalf("openMem: %v", err)
	}
	defer database.Close()

	// Insert directly so we can control created_at and avoid SQLite's
	// second-precision CURRENT_TIMESTAMP, which would race two consecutive
	// Enqueue calls.
	if _, err := database.Exec(
		`INSERT INTO queue (agent_id, label, locator, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		"alpha", "alpha", "pane:%1", "first", "2026-01-01 10:00:00",
	); err != nil {
		t.Fatalf("insert alpha: %v", err)
	}
	if _, err := database.Exec(
		`INSERT INTO queue (agent_id, label, locator, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		"beta", "beta", "pane:%2", "second", "2026-01-01 10:00:01",
	); err != nil {
		t.Fatalf("insert beta: %v", err)
	}

	entry, err := Show(database)
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if got, want := entry.Label, "alpha"; got != want {
		t.Errorf("Label = %q, want %q (oldest first)", got, want)
	}
}
