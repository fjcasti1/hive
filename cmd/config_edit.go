package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open ~/.hive/config.yaml in $EDITOR with validation",
	Long: `Open the hive config file in your $EDITOR (or 'vi' if unset). When
you save and exit, hive validates the result. If validation fails, the
error is prepended as a comment and the editor is reopened so you can
fix it. Loops until the file is valid or you decline to retry.

If the config file does not exist yet, the current effective configuration
(defaults plus any session overrides) is written to disk first so you have
a complete file to edit.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		// Always rewrite the file with the current effective config before
		// opening the editor. This auto-migrates the file to the current
		// schema each time the user edits — new fields added in a recent
		// version appear with their defaults so the user can see and tune
		// them. The user's customized values are preserved (Load + Save
		// round-trip keeps them).
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("write config before edit: %w", err)
		}

		for {
			if err := openInEditor(path); err != nil {
				return err
			}
			if _, err := config.Load(); err == nil {
				// Strip any leftover error comments from previous iterations.
				if err := stripErrorComments(path); err != nil {
					return err
				}
				fmt.Println("config saved.")
				return nil
			} else {
				if err := prependErrorComment(path, err); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "config invalid: %v\n", err)
				if !confirmRetry() {
					return fmt.Errorf("config edit aborted; %s left with error comment", path)
				}
			}
		}
	},
}

const errorCommentPrefix = "# ERROR: "

// openInEditor invokes $EDITOR (fallback "vi") on the given path,
// inheriting stdio so the user's terminal experience is preserved.
func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}
	return nil
}

// prependErrorComment writes the validation error as a YAML comment at the
// top of the file, replacing any prior error comment block. The user sees
// the error immediately on reopen.
func prependErrorComment(path string, err error) error {
	current, readErr := os.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	stripped := stripErrorCommentLines(string(current))
	header := errorCommentPrefix + err.Error() + "\n"
	return os.WriteFile(path, []byte(header+stripped), 0o644)
}

// stripErrorComments removes any lines starting with the errorCommentPrefix
// from the file. Called after a successful save so the file ends clean.
func stripErrorComments(path string) error {
	current, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	stripped := stripErrorCommentLines(string(current))
	if stripped == string(current) {
		return nil
	}
	return os.WriteFile(path, []byte(stripped), 0o644)
}

func stripErrorCommentLines(content string) string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, errorCommentPrefix) {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// confirmRetry prompts the user y/n on stdin. Default no.
func confirmRetry() bool {
	fmt.Fprint(os.Stderr, "retry edit? [y/N]: ")
	r := bufio.NewReader(os.Stdin)
	answer, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
