package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

// statusEntry is the shape exposed to templates and JSON output. Field tags
// drive the JSON schema; field names drive the template token names.
type statusEntry struct {
	Session    string    `json:"session"`
	Message    string    `json:"message"`
	Pane       string    `json:"pane"`
	Age        string    `json:"age"`
	NotifiedAt time.Time `json:"notified_at"`
}

// statusData is the root context passed to templates and the JSON serializer.
// Count,  and Next are derivable from Queue and are exposed for template
// convenience only — JSON consumers compute them from the queue array
// themselves.
type statusData struct {
	Count int           `json:"count"`
	Next  *statusEntry  `json:"next"`
	Queue []statusEntry `json:"queue"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show queue state for humans, tmux, or programmatic consumers",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		return runStatus(database, cfg, format)
	},
}

func init() {
	statusCmd.Flags().String("format", "human", "output format: human, tmux, json")
}

// runStatus is the rendering entry point shared with the deprecated `list`
// alias.
func runStatus(database *sql.DB, cfg *config.Config, format string) error {
	data, err := buildStatusData(database)
	if err != nil {
		return err
	}
	switch format {
	case "human":
		if err := execTemplate(os.Stdout, cfg.Status.HumanFormat, data); err != nil {
			return err
		}
		fmt.Println()
		return nil
	case "tmux":
		return execTemplate(os.Stdout, cfg.Status.TmuxFormat, data)
	case "json":
		return json.NewEncoder(os.Stdout).Encode(data)
	default:
		return fmt.Errorf("unknown format %q (valid: human, tmux, json)", format)
	}
}

// buildStatusData reads the queue and shapes it for rendering.
func buildStatusData(database *sql.DB) (*statusData, error) {
	entries, err := db.List(database)
	if err != nil {
		return nil, err
	}

	data := &statusData{
		Count: len(entries),
		Queue: make([]statusEntry, 0, len(entries)),
	}
	for _, e := range entries {
		data.Queue = append(data.Queue, statusEntry{
			Session:    e.Session,
			Message:    e.Message,
			Pane:       e.Pane,
			Age:        timeAgo(e.CreatedAt),
			NotifiedAt: e.CreatedAt,
		})
	}
	if data.Count > 0 {
		data.Next = &data.Queue[0]
	}
	return data, nil
}

// execTemplate parses tmplStr and writes its rendering of data to w.
// Uses only text/template built-ins — no custom function map.
func execTemplate(w io.Writer, tmplStr string, data *statusData) error {
	tmpl, err := template.New("hive_status").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(w, data)
}
