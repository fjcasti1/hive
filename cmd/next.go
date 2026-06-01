package cmd

import (
	"fmt"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/tmux"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next waiting session (FIFO)",
	RunE: func(cmd *cobra.Command, args []string) error {
		doAck, _ := cmd.Flags().GetBool("ack")
		doShow, _ := cmd.Flags().GetBool("show")

		entry, err := db.Show(database)
		if err != nil {
			return err
		}
		if entry == nil {
			fmt.Println("no sessions waiting")
			return nil
		}

		session, message := entry.Session, entry.Message

		// --show prints the front-of-queue entry and exits without switching.
		// It takes precedence over --ack, to avoid acknowledging a session
		// without switching to it.
		if doShow {
			fmt.Println(formatNextLine(session, message))
			return nil
		}

		if err := tmux.SwitchTo(entry.Target()); err != nil {
			return err
		}

		if doAck {
			found, err := ackSession(database, session)
			if err != nil {
				return fmt.Errorf("ack session=%q: %w", session, err)
			}
			if found {
				fmt.Printf("Acknowledged session %q\n", session)
			}
		}
		return nil
	},
}

func init() {
	nextCmd.Flags().Bool("ack", false, "Acknowledge the session after switching")
	nextCmd.Flags().Bool("show", false, "Show the next waiting session without switching")
}

// formatNextLine renders the line printed when `next` reports the front-of-queue
// entry without switching (the --show path, and the empty-queue case). found is
// false when the queue is empty.
func formatNextLine(session, message string) string {
	if message != "" {
		return fmt.Sprintf("%s — %s", session, message)
	}
	return session
}
