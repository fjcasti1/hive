package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show resolved notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := db.ListHistory(database)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("no history yet")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  session\tmessage\tnotified\tresolved")
		for _, e := range entries {
			msg := e.Message
			if msg == "" {
				msg = "-"
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", e.Session, msg, timeAgo(e.NotifiedAt), timeAgo(e.ResolvedAt))
		}
		return w.Flush()
	},
}
