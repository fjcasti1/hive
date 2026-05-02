package cmd

import (
	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/tmux"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Add current session to the waiting queue and fire notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		sessionName, _ := cmd.Flags().GetString("session")

		if sessionName == "" {
			var err error
			sessionName, err = tmux.CurrentSession()
			if err != nil {
				return err
			}
		}

		if len(msg) > cfg.Queue.MaxMessageLength {
			msg = msg[:cfg.Queue.MaxMessageLength]
		}

		return db.Enqueue(database, sessionName, msg)
	},
}

func init() {
	notifyCmd.Flags().StringP("message", "m", "", "Why the agent needs attention (max 100 chars)")
	notifyCmd.Flags().StringP("session", "s", "", "tmux session name (auto-detected if omitted)")
}
