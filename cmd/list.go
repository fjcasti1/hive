package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all waiting sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		queueEntries, err := db.List(database)
		if err != nil {
			return err
		}
		if len(queueEntries) == 0 {
			fmt.Println("no sessions waiting")
		} else {
			w := tabwriter.NewWriter(
				os.Stdout,
				0,
				0,
				2,
				' ',
				0,
			)
			fmt.Fprintln(w, "  #\tsession\tpane\tmessage\twaiting")
			for i, e := range queueEntries {
				msg := e.Message
				if msg == "" {
					msg = "-"
				}
				pane := e.Target()
				if pane == "" {
					pane = "-"
				}
				fmt.Fprintf(
					w,
					"  %d\t%s\t%s\t%s\t%s\n",
					i+1,
					e.Label,
					pane,
					msg,
					timeAgo(e.CreatedAt),
				)
			}
			w.Flush()
		}
		return nil
	},
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
