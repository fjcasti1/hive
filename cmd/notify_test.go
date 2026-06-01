package cmd

import (
	"testing"

	"github.com/fjcasti1/hive/internal/db"
)

// runNotify sets the message/session flags, invokes notify's RunE, and resets
// the flags afterwards so state does not leak between tests.
func runNotify(t *testing.T, message, session string) error {
	t.Helper()
	if err := notifyCmd.Flags().Set("message", message); err != nil {
		t.Fatalf("set message flag: %v", err)
	}
	if err := notifyCmd.Flags().Set("session", session); err != nil {
		t.Fatalf("set session flag: %v", err)
	}
	t.Cleanup(func() {
		notifyCmd.Flags().Set("message", "")
		notifyCmd.Flags().Set("session", "")
	})
	return notifyCmd.RunE(notifyCmd, nil)
}

// disableNotificationChannels turns off the macOS and tmux-bell channels so the
// command does not spawn osascript/tmux subprocesses during tests.
func disableNotificationChannels() {
	cfg.Notifications.Macos = false
	cfg.Notifications.TmuxBell = false
}

func TestNotifyCommand(t *testing.T) {
	type notifyCall struct {
		message string
		session string
	}

	tests := []struct {
		name string
		// maxMessageLength overrides cfg.Queue.MaxMessageLength after setup
		// when non-zero (0 leaves the default in place).
		maxMessageLength int
		// notifies is the sequence of notify calls to perform.
		notifies []notifyCall
		// wantErr asserts that the final notify call returns an error.
		wantErr bool
		// wantCount/wantSession/wantMessage assert the final queue state.
		wantCount   int
		wantSession string
		wantMessage string
	}{
		{
			name:        "enqueues entry",
			notifies:    []notifyCall{{message: "tests passing", session: "alpha"}},
			wantCount:   1,
			wantSession: "alpha",
			wantMessage: "tests passing",
		},
		{
			name:             "truncates message",
			maxMessageLength: 5,
			notifies:         []notifyCall{{message: "abcdefghij", session: "beta"}},
			wantCount:        1,
			wantSession:      "beta",
			wantMessage:      "abcde",
		},
		{
			name: "re-enqueue resets message",
			notifies: []notifyCall{
				{message: "first", session: "alpha"},
				{message: "second", session: "alpha"},
			},
			wantCount:   1,
			wantSession: "alpha",
			wantMessage: "second",
		},
		{
			name:     "no session outside tmux errors",
			notifies: []notifyCall{{message: "msg", session: ""}},
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setupTest uses t.Setenv, so subtests must not run in parallel.
			setupTest(t) // clears TMUX
			disableNotificationChannels()
			if tc.maxMessageLength != 0 {
				cfg.Queue.MaxMessageLength = tc.maxMessageLength
			}

			var lastErr error
			for i, n := range tc.notifies {
				lastErr = runNotify(t, n.message, n.session)
				// Only the final call's error is asserted; earlier calls in a
				// success sequence must not fail.
				if !tc.wantErr && lastErr != nil {
					t.Fatalf("notify RunE (call %d): %v", i, lastErr)
				}
			}

			if tc.wantErr {
				if lastErr == nil {
					t.Fatal("expected an error from notify, got nil")
				}
				return
			}

			entries, err := db.List(database)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(entries) != tc.wantCount {
				t.Fatalf("expected %d queued entry(ies), got %d", tc.wantCount, len(entries))
			}
			if tc.wantCount > 0 {
				if entries[0].Session != tc.wantSession {
					t.Errorf("session = %q, want %q", entries[0].Session, tc.wantSession)
				}
				if entries[0].Message != tc.wantMessage {
					t.Errorf("message = %q, want %q", entries[0].Message, tc.wantMessage)
				}
			}
		})
	}
}
