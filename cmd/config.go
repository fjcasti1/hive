package cmd

import (
	"fmt"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify hive configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the current effective configuration as YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value and persist it to ~/.hive/config.yaml.

Available keys:
  notifications.macos        bool    enable macOS notification popups
  notifications.tmux_bell    bool    enable tmux bell on notify
  queue.max_message_length   int     maximum length of agent message
  history.retention_days     int     days to keep history (0 disables)
  status.human_format        string  Go text/template for default output
  status.tmux_format         string  Go text/template for --format=tmux

For string-typed keys, prefix the value with "@" to use a built-in
preset, e.g. "@compact" or "@verbose":

  hive config set status.human_format @compact
  hive config set status.tmux_format @minimal

Examples:
  hive config set notifications.macos false
  hive config set queue.max_message_length 200
  hive config set status.human_format @verbose
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if err := config.Set(cfg, key, value); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("set %s = %s\n", key, value)
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset <key>",
	Short: "Reset a configuration key to its default value",
	Long: `Reset one configuration key to its built-in default. Useful when a
custom value (especially a template) is broken or no longer wanted, without
losing the rest of your customizations.

Examples:
  hive config reset notifications.macos
  hive config reset status.human_format
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if err := config.Reset(cfg, key); err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("reset %s to default\n", key)
		return nil
	},
}

func init() {
	configCmd.AddCommand(
		configShowCmd,
		configSetCmd,
		configResetCmd,
		configEditCmd,
	)
}
