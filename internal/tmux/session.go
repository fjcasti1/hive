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
