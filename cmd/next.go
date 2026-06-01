package cmd

import (
	"fmt"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/fjcasti1/hive/internal/tmux"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next waiting agent (FIFO)",
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

		// --show prints the front-of-queue entry and exits without switching.
		// It takes precedence over --ack, to avoid acknowledging an agent
		// without switching to it.
		if doShow {
			fmt.Println(formatNextLine(entry.Label, entry.Message))
			return nil
		}

		// An agent with no tmux pane (e.g. a bare process tracked only by cwd)
		// can't be switched to. Report where it lives and leave it queued
		// rather than silently acking something we couldn't reach.
		target := entry.Target()
		if target == "" {
			fmt.Printf("%s is not in tmux — cannot switch", entry.Label)
			if entry.Locator != "" {
				fmt.Printf(" (%s)", entry.Locator)
			}
			fmt.Println()
			return nil
		}

		if err := tmux.SwitchTo(target); err != nil {
			return err
		}

		if doAck {
			label, found, err := ackAgent(database, entry.AgentID)
			if err != nil {
				return fmt.Errorf("ack agent=%q: %w", entry.AgentID, err)
			}
			if found {
				fmt.Printf("Acknowledged session %q\n", label)
			}
		}
		return nil
	},
}

func init() {
	nextCmd.Flags().Bool("ack", false, "Acknowledge the agent after switching")
	nextCmd.Flags().Bool("show", false, "Show the next waiting agent without switching")
}

// formatNextLine renders the line printed when `next` reports the front-of-queue
// entry without switching (the --show path, and the empty-queue case).
func formatNextLine(label, message string) string {
	if message != "" {
		return fmt.Sprintf("%s — %s", label, message)
	}
	return label
}
