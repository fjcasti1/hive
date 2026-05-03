package cmd

import (
	"database/sql"
	"os"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/fjcasti1/hive/internal/db"
	"github.com/spf13/cobra"
)

// NOTE: Cobra lifecycle order is important for understanding when to load config and initialize resources.
// ----------
// - cobra.OnInitialize(fn ...func()) -- A package-level registration. The functions you
// pass run once, very early, as part of rootCmd.Execute() — after Cobra has parsed flags
// but before it dispatches to any command's Run/PreRun.
// - PersistentPreRunE (and its non-E and non-Persistent siblings) A field on a *cobra.Command.
// Cobra invokes it as part of the command lifecycle for that command and its descendants.
// ----------
// Lifecycle order for myapp serve --port 9000
// - 1. cobra.OnInitialize callbacks       ← config loading, no cmd context
// - 2. PersistentPreRunE  (root → ... → serve, nearest one wins)
// - 3. PreRunE            (serve only)
// - 4. RunE               (serve only)
// - 5. PostRunE           (serve only)
// - 6. PersistentPostRunE (nearest one wins)

var (
	version  = "dev"
	cfg      config.Config
	database *sql.DB

	rootCmd = &cobra.Command{
		Use:     "hive",
		Short:   "Manage multiple agentic tmux sessions",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.Load()
			if err != nil {
				return err
			}
			database, err = db.Open()
			if err != nil {
				return err
			}
			err = db.PurgeHistory(database, cfg.History.RetentionDays)
			if err != nil {
				return err
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if database != nil {
				database.Close()
			}
		},
	}
)

func init() {
	rootCmd.SetVersionTemplate("hive {{.Version}}\n")
	rootCmd.AddCommand(
		notifyCmd,
		listCmd,
		ackCmd,
		historyCmd,
		nextCmd,
		configCmd,
	// 	snoozeCmd,
	// 	pauseCmd,
	// 	resumeCmd,
	// 	statusCmd,
	// 	installCmd,
	// 	doctorCmd,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
