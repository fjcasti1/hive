package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var ackCmd = &cobra.Command{
	Use:   "ack [agent-or-index]",
	Short: "Acknowledge an agent — mark feedback given and move to history",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idFlag, _ := cmd.Flags().GetString("id")
		labelFlag, _ := cmd.Flags().GetString("session")

		// key is what the user (or the resolver) names the agent by: an
		// agent_id or a label. It is resolved to a concrete agent_id below.
		var key string
		switch {
		case idFlag != "":
			key = idFlag
		case labelFlag != "":
			key = labelFlag
		case len(args) > 0:
			arg := args[0]
			if idx, err := strconv.Atoi(arg); err == nil {
				entries, err := db.List(database)
				if err != nil {
					return err
				}
				if idx < 1 || idx > len(entries) {
					return fmt.Errorf("index %d out of range (1-%d)", idx, len(entries))
				}
				key = entries[idx-1].AgentID
			} else {
				key = arg
			}
		default:
			agent, err := resolveAgent(idFlag, labelFlag, readHookInput())
			if err != nil {
				return err
			}
			key = agent.ID
		}

		agentID, err := resolveAckKey(database, key)
		if err != nil {
			return err
		}

		label, found, err := ackAgent(database, agentID)
		if err != nil {
			return fmt.Errorf("ack agent=%q: %w", agentID, err)
		}
		if !found {
			fmt.Printf("No waiting session found for %q\n", key)
		} else {
			fmt.Printf("Acknowledged session %q\n", label)
		}
		return nil
	},
}

func init() {
	ackCmd.Flags().StringP("session", "s", "", "Agent label (auto-detected from tmux if omitted)")
	ackCmd.Flags().String("id", "", "Agent id (auto-detected from the Claude hook or tmux if omitted)")
}

// resolveAckKey maps a user-supplied key (an agent_id or a label) to a concrete
// agent_id by matching against the queue. If nothing matches, the key is
// returned unchanged so the caller's delete cleanly reports "not found".
func resolveAckKey(database *sql.DB, key string) (string, error) {
	entries, err := db.List(database)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.AgentID == key || e.Label == key {
			return e.AgentID, nil
		}
	}
	return key, nil
}

// ackAgent removes the queue entry for the given agent and inserts the
// corresponding history row in a single transaction. It returns the agent's
// label and whether a row was acknowledged (false if no queue entry existed).
func ackAgent(database *sql.DB, agentID string) (string, bool, error) {
	var label string
	var found bool
	err := db.WithTx(database, func(tx *sql.Tx) error {
		deleted, err := db.Delete(tx, agentID)
		if err != nil || deleted == nil {
			return err
		}
		found = true
		label = deleted.Label
		return db.AddHistory(tx, deleted.Label, deleted.Message, deleted.NotifiedAt)
	})
	return label, found, err
}
