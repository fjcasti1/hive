// Package db is the SQLite persistence layer for hive. It manages the queue of
// pending session messages and the history of resolved notifications, applying
// any pending schema migrations when the database is opened.
package db

// db.go holds the database lifecycle: opening the connection, running schema
// migrations, and the transaction/Querier primitives shared across the package.

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Querier is the subset of *sql.DB and *sql.Tx that db functions use.
// Atomic primitives accept a Querier so callers can run them standalone
// or compose them inside a transaction via WithTx.
type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// WithTx runs fn inside a transaction. Commits on nil error, rolls back otherwise.
func WithTx(database *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// DBPath returns the filesystem path to the SQLite database, located at
// ~/.hive/hive.db.
func DBPath() string {
	return filepath.Join(os.Getenv("HOME"), ".hive", "hive.db")
}

// Open opens a connection to the SQLite database, creating it and its parent directories if needed.
// It also runs any pending migrations before returning the database connection.
// It should only be called once per process, and callers should defer Close() on the returned *sql.DB.
// DO NOT call Open() for individual commands, only the root command should call Open() in its
// PersistentPreRunE, and the resulting *sql.DB should be shared with subcommands via closure or context.
func Open() (*sql.DB, error) {
	path := DBPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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

func migrate(database *sql.DB) error {
	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	if err := goose.Up(database, "migrations"); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
