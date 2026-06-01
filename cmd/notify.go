package cmd

import (
	"fmt"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/notifications"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Add the current agent to the waiting queue and fire notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		idFlag, _ := cmd.Flags().GetString("id")
		labelFlag, _ := cmd.Flags().GetString("session")

		hook := readHookInput()
		agent, err := resolveAgent(idFlag, labelFlag, hook)
		if err != nil {
			return err
		}

		// A Claude Notification hook carries its own message; use it when one
		// wasn't passed explicitly.
		if msg == "" && hook != nil {
			msg = hook.Message
		}
		if len(msg) > cfg.Queue.MaxMessageLength {
			msg = msg[:cfg.Queue.MaxMessageLength]
		}

		if err := db.Enqueue(database, agent.ID, agent.Label, agent.Locator, msg); err != nil {
			return fmt.Errorf("queue error: %w", err)
		}

		// Fire notifications immediately after enqueueing, so that the user gets
		// feedback even if they don't have a tmux status line set up to show
		// the queue length.
		channels := notifications.Channels(cfg)
		notifications.Dispatch(channels, agent.Label, msg)

		return nil
	},
}

func init() {
	notifyCmd.Flags().StringP("message", "m", "", "Why the agent needs attention (max 100 chars)")
	notifyCmd.Flags().StringP("session", "s", "", "Display label for the agent (auto-detected from tmux if omitted)")
	notifyCmd.Flags().String("id", "", "Stable agent id (auto-detected from the Claude hook or tmux if omitted)")
}
