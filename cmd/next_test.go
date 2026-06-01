package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/fjcasti1/hive/internal/db"
)

func TestFormatNextLine(t *testing.T) {
	tests := []struct {
		name    string
		session string
		message string
		want    string
	}{
		{
			name:    "with message",
			session: "alpha",
			message: "tests passing",
			want:    "alpha — tests passing",
		},
		{
			name:    "without message",
			session: "gamma",
			message: "",
			want:    "gamma",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatNextLine(tc.session, tc.message)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// runNextShow drives the `--show` path of the next command, which prints the
// head of the queue without switching tmux (so it needs no tmux server).
func runNextShow(t *testing.T) string {
	t.Helper()
	if err := nextCmd.Flags().Set("show", "true"); err != nil {
		t.Fatalf("set show flag: %v", err)
	}
	t.Cleanup(func() { nextCmd.Flags().Set("show", strconv.FormatBool(false)) })

	return captureStdout(t, func() {
		if err := nextCmd.RunE(nextCmd, nil); err != nil {
			t.Fatalf("next RunE: %v", err)
		}
	})
}

func TestNextCommand(t *testing.T) {
	type entry struct {
		session string
		message string
		pane    string
	}

	tests := []struct {
		name          string
		seed          []entry
		show          bool   // use the --show path; otherwise capture RunE directly
		wantOutput    string // empty means assert via wantContains
		wantContains  string
		wantRemaining int
	}{
		{
			name:         "empty queue",
			seed:         nil,
			show:         false,
			wantContains: "no sessions waiting",
		},
		{
			name: "show with message does not consume",
			seed: []entry{
				{session: "alpha", message: "tests passing", pane: "%1"},
			},
			show:          true,
			wantOutput:    "alpha — tests passing",
			wantRemaining: 1,
		},
		{
			name: "show without message",
			seed: []entry{
				{session: "gamma", message: "", pane: ""},
			},
			show:          true,
			wantOutput:    "gamma",
			wantRemaining: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setupTest uses t.Setenv, so subtests must not run in parallel.
			setupTest(t)

			for _, e := range tc.seed {
				locator := ""
				if e.pane != "" {
					locator = "pane:" + e.pane
				}
				if err := db.Enqueue(database, e.session, e.session, locator, e.message); err != nil {
					t.Fatalf("Enqueue: %v", err)
				}
			}

			var out string
			if tc.show {
				out = strings.TrimSpace(runNextShow(t))
			} else {
				out = captureStdout(t, func() {
					if err := nextCmd.RunE(nextCmd, nil); err != nil {
						t.Fatalf("next RunE: %v", err)
					}
				})
			}

			if tc.wantContains != "" {
				if !strings.Contains(out, tc.wantContains) {
					t.Errorf("output %q should contain %q", out, tc.wantContains)
				}
			} else if out != tc.wantOutput {
				t.Errorf("got %q, want %q", out, tc.wantOutput)
			}

			// --show must not consume entries.
			if tc.show {
				entries, err := db.List(database)
				if err != nil {
					t.Fatalf("List: %v", err)
				}
				if len(entries) != tc.wantRemaining {
					t.Errorf("--show should leave the queue intact, got %d entries, want %d", len(entries), tc.wantRemaining)
				}
			}
		})
	}
}
