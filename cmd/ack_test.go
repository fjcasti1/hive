package cmd

import (
	"strings"
	"testing"

	"github.com/fjcasti1/hive/internal/db"
)

// TestAckSession covers the ackSession unit: acking a queued session removes it
// and records history, while acking a missing session is a no-op returning found=false.
func TestAckSession(t *testing.T) {
	tests := []struct {
		name        string
		enqueue     bool
		session     string
		wantFound   bool
		wantHistory int
	}{
		{
			name:        "removes and records history",
			enqueue:     true,
			session:     "alpha",
			wantFound:   true,
			wantHistory: 1,
		},
		{
			name:        "missing returns false",
			enqueue:     false,
			session:     "ghost",
			wantFound:   false,
			wantHistory: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTest(t)

			if tc.enqueue {
				if err := db.Enqueue(database, tc.session, "tests passing", "%1"); err != nil {
					t.Fatalf("Enqueue: %v", err)
				}
			}

			found, err := ackSession(database, tc.session)
			if err != nil {
				t.Fatalf("ackSession: %v", err)
			}
			if found != tc.wantFound {
				t.Errorf("ackSession found = %v, want %v", found, tc.wantFound)
			}

			entries, err := db.List(database)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("queue should be empty after ack, got %d entries", len(entries))
			}

			hist, err := db.ListHistory(database)
			if err != nil {
				t.Fatalf("ListHistory: %v", err)
			}
			if len(hist) != tc.wantHistory {
				t.Fatalf("history should have %d entry(ies), got %d", tc.wantHistory, len(hist))
			}
			if tc.wantHistory > 0 {
				if hist[0].Session != tc.session || hist[0].Message != "tests passing" {
					t.Errorf("history entry = %+v, want session=%s message=%q", hist[0], tc.session, "tests passing")
				}
			}
		})
	}
}

// TestAckCommand covers the ackCmd cobra command behaviors: acking by index,
// out-of-range index errors, acking by session name, unknown-session reporting,
// and the no-args-outside-tmux error path.
func TestAckCommand(t *testing.T) {
	tests := []struct {
		name string
		// enqueue is the list of sessions to seed the queue with, in order.
		enqueue []string
		args    []string
		// wantErr / errContains describe the error-path expectation.
		wantErr     bool
		errContains string
		// wantOutput is a substring expected in stdout (success path only).
		wantOutput string
		// wantQueue is the expected session names remaining after the command.
		wantQueue []string
	}{
		{
			name:       "by index",
			enqueue:    []string{"alpha", "beta"},
			args:       []string{"1"},
			wantOutput: "alpha",
			wantQueue:  []string{"beta"},
		},
		{
			name:        "index out of range",
			enqueue:     []string{"alpha"},
			args:        []string{"5"},
			wantErr:     true,
			errContains: "out of range",
		},
		{
			name:       "by session name",
			enqueue:    []string{"alpha"},
			args:       []string{"alpha"},
			wantOutput: "Acknowledged",
			wantQueue:  nil,
		},
		{
			name:       "unknown session reports not found",
			enqueue:    nil,
			args:       []string{"nope"},
			wantOutput: "No waiting session",
		},
		{
			name:    "no args outside tmux",
			enqueue: nil,
			args:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTest(t) // setupTest clears TMUX

			// Two entries with the same-second created_at rely on insertion order
			// (id) to break ties, so List returns them in enqueue order.
			for _, s := range tc.enqueue {
				if err := db.Enqueue(database, s, "", ""); err != nil {
					t.Fatalf("Enqueue %s: %v", s, err)
				}
			}

			var runErr error
			out := captureStdout(t, func() {
				runErr = ackCmd.RunE(ackCmd, tc.args)
			})

			if tc.wantErr {
				if runErr == nil {
					t.Fatalf("expected an error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(runErr.Error(), tc.errContains) {
					t.Errorf("error = %q, want it to mention %q", runErr.Error(), tc.errContains)
				}
				return
			}

			if runErr != nil {
				t.Fatalf("ack RunE: %v", runErr)
			}
			if tc.wantOutput != "" && !strings.Contains(out, tc.wantOutput) {
				t.Errorf("output %q should contain %q", out, tc.wantOutput)
			}

			entries, err := db.List(database)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(entries) != len(tc.wantQueue) {
				t.Fatalf("queue has %d entries, want %d: %+v", len(entries), len(tc.wantQueue), entries)
			}
			for i, want := range tc.wantQueue {
				if entries[i].Session != want {
					t.Errorf("queue[%d].Session = %q, want %q", i, entries[i].Session, want)
				}
			}
		})
	}
}
