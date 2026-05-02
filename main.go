package main

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

func main() {
	cmd := &cobra.Command{
		Use:     "hive",
		Short:   "Manage multiple AI agent sessions in tmux",
		Version: Version,
	}
	cmd.SetVersionTemplate("hive {{.Version}}\n")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
