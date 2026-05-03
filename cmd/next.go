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

		entry, err := db.Peek(database)
		if err != nil {
			return err
		}
		if entry == nil {
			fmt.Println("no sessions waiting")
			return nil
		}

		if err := tmux.SwitchTo(entry.Target()); err != nil {
			return err
		}

		if doAck {
			sessionName := entry.Session
			found, err := ackSession(database, sessionName)
			if err != nil {
				return fmt.Errorf("ack session=%q: %w", sessionName, err)
			}
			if found {
				fmt.Printf("Acknowledged session %q\n", sessionName)
			}
		}
		return nil
	},
}

func init() {
	nextCmd.Flags().Bool("ack", false, "Acknowledge the session after switching")
}
