package notifications

import (
	"fmt"
	"os/exec"
)

type macosChannel struct{}

func NewMacOSChannel() Channel { return macosChannel{} }

func (macosChannel) Name() string { return "macos" }
func (macosChannel) Dispatch(session, msg string) error {
	title := fmt.Sprintf("hive: %s", session)
	body := msg
	if body == "" {
		body = "Agent needs your attention"
	}
	script := fmt.Sprintf(`display notification %q with title %q sound name "Ping"`, body, title)
	return exec.Command("osascript", "-e", script).Run()
}
