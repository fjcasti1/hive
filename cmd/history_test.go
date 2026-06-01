package cmd

import (
	"strings"
	"testing"

	"github.com/fjcasti1/hive/internal/db"
)

func TestHistoryCommand(t *testing.T) {
	type seedEntry struct {
		session string
		message string
	}

	tests := []struct {
		name         string
		seed         []seedEntry
		wantContains []string
	}{
		{
			name:         "empty history",
			seed:         nil,
			wantContains: []string{"no history yet"},
		},
		{
			name: "with entries",
			seed: []seedEntry{
				{session: "alpha", message: "shipped"},
			},
			wantContains: []string{"session", "alpha", "shipped", "notified", "resolved"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setupTest uses t.Setenv, so this subtest must not run in parallel.
			setupTest(t)

			// Queue then acknowledge so each entry lands in history.
			for _, e := range tc.seed {
				if err := db.Enqueue(database, e.session, e.session, "pane:%1", e.message); err != nil {
					t.Fatalf("Enqueue(%q): %v", e.session, err)
				}
				if _, _, err := ackAgent(database, e.session); err != nil {
					t.Fatalf("ackAgent(%q): %v", e.session, err)
				}
			}

			out := captureStdout(t, func() {
				if err := historyCmd.RunE(historyCmd, nil); err != nil {
					t.Fatalf("history RunE: %v", err)
				}
			})

			for _, want := range tc.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("output should contain %q\n--- output ---\n%s", want, out)
				}
			}
		})
	}
}
