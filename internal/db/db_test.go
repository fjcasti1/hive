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
	if version != 1 {
		t.Errorf("want schema version 1, got %d", version)
	}

	// Idempotency: running migrate again should be a no-op.
	if err := migrate(database); err != nil {
		t.Fatalf("re-run migrate: %v", err)
	}
}
