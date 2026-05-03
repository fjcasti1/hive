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
  notifications.macos        bool  enable macOS notification popups
  notifications.tmux_bell    bool  enable tmux bell on notify
  queue.max_message_length   int   maximum length of agent message

Examples:
  hive config set notifications.macos false
  hive config set queue.max_message_length 200`,
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

func init() {
	configCmd.AddCommand(configShowCmd, configSetCmd)
}
