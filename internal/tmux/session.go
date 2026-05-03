package tmux

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

var ErrNotInTmux = errors.New("not running inside a tmux session; use --session to specify one")

func CurrentSession() (string, error) {
	if os.Getenv("TMUX") == "" {
		return "", ErrNotInTmux
	}
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// SwitchTo switches the current tmux client to the given target. The target
// can be a session name, a pane id (e.g. %5), or any value tmux's switch-client
// accepts via -t. Subprocess stderr is forwarded so any tmux error is visible.
func SwitchTo(target string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", target)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
