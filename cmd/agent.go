package cmd

// agent.go resolves the identity of the agent invoking hive. A single resolver
// is shared by `notify` (which enqueues) and `ack` (which dequeues) so the two
// always agree on the key for a given agent — if they disagreed, acks would
// miss and entries would leak.
//
// Identity has three parts, each resolved independently from the best source
// available:
//
//	agent_id — the stable dedup key: --id > $HIVE_AGENT_ID > hook session_id >
//	           tmux session name > --session. Errors if none is available.
//	label    — the human-readable display name: --session > tmux session name >
//	           hook cwd basename > agent_id.
//	locator  — how to reach the agent: "pane:%N" in tmux, else "cwd:/path" when
//	           a hook reported one, else empty (not navigable).
//
// hive is assumed to run against a single tmux server, or none — never multiple
// servers — so a pane id alone is an unambiguous switch target.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/fjcasti1/hive/internal/tmux"
)

// errNoIdentity is returned when no identity source is available: not in tmux,
// no hook payload on stdin, and no explicit --id/--session given.
var errNoIdentity = errors.New("cannot determine agent identity: run inside tmux, pipe a Claude hook payload, or pass --id/--session")

// agentIdentity is the resolved identity of the invoking agent.
type agentIdentity struct {
	ID      string
	Label   string
	Locator string
}

// hookInput is the subset of a Claude Code hook's JSON stdin payload that hive
// uses to identify the agent. All Claude hooks (Stop, Notification, etc.) carry
// session_id and cwd; only some carry message.
type hookInput struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	Message   string `json:"message"`
}

// readHookInput reads and parses a Claude hook JSON payload from stdin, or
// returns nil when there is none. It only reads when stdin is a pipe (not a
// terminal or /dev/null), so an interactive `hive notify` never blocks waiting
// for input that will not come.
func readHookInput() *hookInput {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice != 0 {
		return nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	var h hookInput
	if err := json.Unmarshal(data, &h); err != nil {
		return nil
	}
	return &h
}

// resolveAgent derives the invoking agent's identity from explicit flags, the
// environment, an optional Claude hook payload, and the tmux context.
func resolveAgent(idFlag, labelFlag string, hook *hookInput) (agentIdentity, error) {
	var tmuxSession, tmuxPane string
	if os.Getenv("TMUX") != "" {
		tmuxSession, _ = tmux.CurrentSession()
		tmuxPane = os.Getenv("TMUX_PANE")
	}

	id := firstNonEmpty(idFlag, os.Getenv("HIVE_AGENT_ID"))
	if id == "" && hook != nil {
		id = hook.SessionID
	}
	id = firstNonEmpty(id, tmuxSession, labelFlag)
	if id == "" {
		return agentIdentity{}, errNoIdentity
	}

	label := firstNonEmpty(labelFlag, tmuxSession)
	if label == "" && hook != nil && hook.Cwd != "" {
		label = filepath.Base(hook.Cwd)
	}
	label = firstNonEmpty(label, id)

	var locator string
	switch {
	case tmuxPane != "":
		locator = "pane:" + tmuxPane
	case hook != nil && hook.Cwd != "":
		locator = "cwd:" + hook.Cwd
	}

	return agentIdentity{ID: id, Label: label, Locator: locator}, nil
}

// firstNonEmpty returns the first non-empty string in vals, or "".
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
