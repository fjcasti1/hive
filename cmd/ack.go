package cmd

import (
	"fmt"
	"strconv"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var ackCmd = &cobra.Command{
	Use:   "ack [session-or-index]",
	Short: "Acknowledge a session — mark feedback given and move to history",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName, _ := cmd.Flags().GetString("session")

		if len(args) > 0 {
			arg := args[0]
			if idx, err := strconv.Atoi(arg); err == nil {
				entries, err := db.List(database)
				if err != nil {
					return err
				}
				if idx < 1 || idx > len(entries) {
					return fmt.Errorf("index %d out of range (1-%d)", idx, len(entries))
				}
				sessionName = entries[idx-1].Session
			} else {
				sessionName = arg
			}
		}

		wasDeleted, err := db.Delete(database, sessionName)
		if err != nil {
			return err
		}
		if wasDeleted {
			fmt.Printf("Acknowledged session %q\n", sessionName)
		} else {
			fmt.Printf("No waiting session found for %q\n", sessionName)
		}
		return nil
	},
}

func init() {
	ackCmd.Flags().StringP("session", "s", "", "tmux session name (auto-detected if omitted)")
}
