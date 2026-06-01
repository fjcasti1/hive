package db

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
)

// TestMigration00004Backfill verifies that queue rows created under the old
// session/pane schema (v3) are carried into the new agent_id/label/locator
// columns when migration 00004 runs.
func TestMigration00004Backfill(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()

	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("dialect: %v", err)
	}

	// Migrate to the pre-decoupling schema and seed an old-shape row.
	if err := goose.UpTo(database, "migrations", 3); err != nil {
		t.Fatalf("UpTo(3): %v", err)
	}
	if _, err := database.Exec(
		`INSERT INTO queue (session, message, pane) VALUES (?, ?, ?)`,
		"alpha", "tests passing", "%5",
	); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Apply 00004.
	if err := goose.UpTo(database, "migrations", 4); err != nil {
		t.Fatalf("UpTo(4): %v", err)
	}

	var agentID, label, locator, message string
	if err := database.QueryRow(
		`SELECT agent_id, label, locator, message FROM queue WHERE agent_id = ?`, "alpha",
	).Scan(&agentID, &label, &locator, &message); err != nil {
		t.Fatalf("read migrated row: %v", err)
	}
	if agentID != "alpha" || label != "alpha" {
		t.Errorf("agent_id/label = %q/%q, want alpha/alpha", agentID, label)
	}
	if locator != "pane:%5" {
		t.Errorf("locator = %q, want pane:%%5", locator)
	}
	if message != "tests passing" {
		t.Errorf("message = %q, want %q", message, "tests passing")
	}
}
