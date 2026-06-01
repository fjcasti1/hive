package cmd

import (
	"os"
	"testing"
)

// TestResolveAgent covers the identity precedence rules that notify and ack
// share. tmux branches aren't exercised here (they shell out to tmux); TMUX is
// cleared so resolution falls through to the hook/flag/env sources.
func TestResolveAgent(t *testing.T) {
	tests := []struct {
		name        string
		idFlag      string
		labelFlag   string
		envID       string
		hook        *hookInput
		wantID      string
		wantLabel   string
		wantLocator string
		wantErr     bool
	}{
		{
			name:      "id flag wins over everything",
			idFlag:    "explicit",
			labelFlag: "label",
			envID:     "env-id",
			hook:      &hookInput{SessionID: "sess"},
			wantID:    "explicit",
			wantLabel: "label",
		},
		{
			name:      "env id when no flag",
			envID:     "env-id",
			wantID:    "env-id",
			wantLabel: "env-id",
		},
		{
			name:        "hook session_id and cwd",
			hook:        &hookInput{SessionID: "sid", Cwd: "/home/user/proj"},
			wantID:      "sid",
			wantLabel:   "proj",
			wantLocator: "cwd:/home/user/proj",
		},
		{
			name:      "label flag is the manual fallback id",
			labelFlag: "mysession",
			wantID:    "mysession",
			wantLabel: "mysession",
		},
		{
			name:    "no identity source errors",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TMUX", "")
			t.Setenv("TMUX_PANE", "")
			t.Setenv("HIVE_AGENT_ID", tc.envID)

			got, err := resolveAgent(tc.idFlag, tc.labelFlag, tc.hook)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveAgent: %v", err)
			}
			if got.ID != tc.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tc.wantID)
			}
			if got.Label != tc.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, tc.wantLabel)
			}
			if got.Locator != tc.wantLocator {
				t.Errorf("Locator = %q, want %q", got.Locator, tc.wantLocator)
			}
		})
	}
}

// TestReadHookInput verifies stdin parsing: a JSON payload on a pipe is parsed,
// while empty or absent input yields nil (so interactive use never blocks).
func TestReadHookInput(t *testing.T) {
	tests := []struct {
		name    string
		stdin   string
		wantNil bool
		wantID  string
	}{
		{name: "valid payload", stdin: `{"session_id":"abc","cwd":"/x","message":"hi"}`, wantID: "abc"},
		{name: "empty input", stdin: "", wantNil: true},
		{name: "garbage is ignored", stdin: "not json", wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			if _, err := w.WriteString(tc.stdin); err != nil {
				t.Fatalf("write: %v", err)
			}
			w.Close()
			orig := os.Stdin
			os.Stdin = r
			t.Cleanup(func() { os.Stdin = orig; r.Close() })

			got := readHookInput()
			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected a payload, got nil")
			}
			if got.SessionID != tc.wantID {
				t.Errorf("SessionID = %q, want %q", got.SessionID, tc.wantID)
			}
		})
	}
}
