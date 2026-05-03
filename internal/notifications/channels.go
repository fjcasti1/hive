package notifications

import (
	"fmt"
	"os"

	"github.com/fjcasti1/hive/internal/config"
)

type Channel interface {
	Dispatch(session, msg string) error
	Name() string
}

func Channels(cfg *config.Config) []Channel {
	var chs []Channel
	if cfg.Notifications.Macos {
		chs = append(chs, NewMacOSChannel())
	}
	if cfg.Notifications.TmuxBell {
		chs = append(chs, NewTmuxChannel())
	}
	return chs
}

func Dispatch(channels []Channel, session, msg string) {
	for _, ch := range channels {
		if err := ch.Dispatch(session, msg); err != nil {
			fmt.Fprintf(os.Stderr, "notify[%s]: %v\n", ch.Name(), err)
		}
	}
}
