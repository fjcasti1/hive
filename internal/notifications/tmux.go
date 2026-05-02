package notifications

import (
	"os"
	"os/exec"
	"strings"
)

type tmuxBellChannel struct{}

func NewTmuxChannel() Channel { return tmuxBellChannel{} }

func (tmuxBellChannel) Name() string { return "tmux-bell" }
func (tmuxBellChannel) Dispatch(session, _ string) error {
	// existing pane-tty + BEL body
	out, err := exec.Command("tmux", "list-panes", "-t", session, "-F", "#{pane_tty}").Output()
	if err != nil {
		return err
	}
	tty := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	f, err := os.OpenFile(tty, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte{0x07}) // BEL
	return err
}
