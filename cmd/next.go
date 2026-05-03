package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Switch to the next waiting session (FIFO)",
	RunE: func(cmd *cobra.Command, args []string) error {
		showOnly, _ := cmd.Flags().GetBool("show")

		entry, err := db.Next(database)
		if err != nil {
			return err
		}
		if entry == nil {
			fmt.Fprintln(os.Stderr, "hive: no sessions waiting")
			return nil
		}

		if showOnly {
			fmt.Println(entry.Target())
			return nil
		}

		return exec.Command("tmux", "switch-client", "-t", entry.Target()).Run()
	},
}

func init() {
	nextCmd.Flags().Bool("show", false, "Show session name instead of switching")
}
