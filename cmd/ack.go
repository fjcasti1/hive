package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/tmux"
	"github.com/spf13/cobra"
)

var ackCmd = &cobra.Command{
	Use:   "ack [session-or-index]",
	Short: "Acknowledge a session — mark feedback given and move to history",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionName, _ := cmd.Flags().GetString("session")

		if sessionName == "" {
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
			} else {
				var err error
				sessionName, err = tmux.CurrentSession()
				if err != nil {
					return err
				}
			}
		}

		found, err := ackSession(database, sessionName)
		if err != nil {
			return fmt.Errorf("ack session=%q: %w", sessionName, err)
		}
		if !found {
			fmt.Printf("No waiting session found for %q\n", sessionName)
		} else {
			fmt.Printf("Acknowledged session %q\n", sessionName)
		}
		return nil
	},
}

func init() {
	ackCmd.Flags().StringP("session", "s", "", "tmux session name (auto-detected if omitted)")
}

// ackSession removes the queue entry for the given session and inserts the
// corresponding history row in a single transaction. Returns true if a row
// was acknowledged, false if no queue entry existed for that session.
func ackSession(database *sql.DB, session string) (bool, error) {
	var found bool
	err := db.WithTx(database, func(tx *sql.Tx) error {
		deleted, err := db.Delete(tx, session)
		if err != nil || deleted == nil {
			return err
		}
		found = true
		return db.AddHistory(tx, deleted.Session, deleted.Message, deleted.NotifiedAt)
	})
	return found, err
}
