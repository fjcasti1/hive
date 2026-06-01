package cmd

import (
	"strings"
	"testing"

	"github.com/fjcasti1/hive/internal/db"
)

func TestListCommand(t *testing.T) {
	type entry struct {
		session string
		message string
		pane    string
	}

	tests := []struct {
		name         string
		entries      []entry
		wantContains []string
	}{
		{
			name:         "empty queue",
			entries:      nil,
			wantContains: []string{"no sessions waiting"},
		},
		{
			name: "with entries",
			entries: []entry{
				{session: "alpha", message: "tests passing", pane: "%1"},
				{session: "beta", message: "", pane: ""},
			},
			// beta has no message and no pane; both render as the "-" placeholder.
			wantContains: []string{"session", "alpha", "tests passing", "%1", "beta", "-"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTest(t)

			for _, e := range tc.entries {
				if err := db.Enqueue(database, e.session, e.message, e.pane); err != nil {
					t.Fatalf("Enqueue %s: %v", e.session, err)
				}
			}

			out := captureStdout(t, func() {
				if err := listCmd.RunE(listCmd, nil); err != nil {
					t.Fatalf("list RunE: %v", err)
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
