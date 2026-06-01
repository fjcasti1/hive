package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/db"
)

// setupTest wires the package globals (database, cfg) to an isolated,
// fully-migrated SQLite database backed by a per-test temp HOME. It also
// clears TMUX so tmux.CurrentSession reports "not in tmux" deterministically.
// Globals and the connection are torn down via t.Cleanup. Because it calls
// t.Setenv, tests using it cannot run in parallel — which is what we want,
// since they share the package-level globals.
func setupTest(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TMUX", "")

	d, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	c, err := config.Load()
	if err != nil {
		d.Close()
		t.Fatalf("config.Load: %v", err)
	}

	database = d
	cfg = c
	t.Cleanup(func() {
		d.Close()
		database = nil
		cfg = nil
	})
}

// captureStdout redirects os.Stdout for the duration of fn and returns
// everything written to it. The command RunE functions render through
// os.Stdout (directly or via a tabwriter constructed at call time), so the
// swap must happen before fn runs.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = orig
	return <-done
}
