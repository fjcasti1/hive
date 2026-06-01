package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/fjcasti1/hive/internal/config"
	"github.com/spf13/cobra"
)

var configTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage status format template files in ~/.hive/templates/",
	Long: `Author and manage .tmpl files used by status format keys.

Template files live in ~/.hive/templates/. Once you've created one, point a
config key at it via 'hive config edit' or 'hive config set':

  hive config set status.human_format ~/.hive/templates/example.tmpl

Subcommands:
  new   create a new template file by name, open in $EDITOR
  edit  open an existing template file by name in $EDITOR
  list  list templates currently in ~/.hive/templates/
`,
}

var configTemplateNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new template file by name and open in $EDITOR",
	Long: `Create ~/.hive/templates/<name>.tmpl and open it in $EDITOR. Validates
on save (re-opens with an error comment if the template doesn't parse).

Errors if a file with that name already exists — use 'hive config template
edit <name>' to modify, or pick a different name.

Optional flag --from <key> seeds the new file with the current effective
template of a config key (e.g. --from status.human_format). Without --from
the file is created with a comment block listing the available data fields
and helper functions.

Note: this command only manages the file. To actually use it, point a
config key at it:

  hive config set status.human_format ~/.hive/templates/<name>.tmpl
  # or:
  hive config edit
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := templatePathForName(name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("template %s already exists at %s; use 'hive config template edit %s' to modify, or pick a different name", name, path, name)
		}

		fromKey, _ := cmd.Flags().GetString("from")
		seed, err := templateSeedContent(fromKey)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
			return err
		}
		if err := editTemplateFile(path, nil); err != nil {
			return err
		}
		// Strip .tmpl from the basename for the bare-name hint.
		bareName := strings.TrimSuffix(filepath.Base(path), ".tmpl")
		fmt.Printf("\nTo use this template, point a config key at it:\n  hive config set status.human_format %s\n", bareName)
		return nil
	},
}

var configTemplateEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Open an existing template file by name in $EDITOR",
	Long: `Open ~/.hive/templates/<name>.tmpl in $EDITOR. Validates on save and
re-opens with an error comment if the template doesn't parse.

Errors if no file with that name exists — use 'hive config template new
<name>' to create one, or 'hive config template list' to see what's
available.
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := templatePathForName(name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("template %s does not exist at %s; use 'hive config template new %s' to create it", name, path, name)
		}
		return editTemplateFile(path, nil)
	},
}

var configTemplateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List template files in ~/.hive/templates/",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := templatesDir()
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("(no templates yet — try 'hive config template new <name>')")
				return nil
			}
			return err
		}
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			if !strings.HasSuffix(n, ".tmpl") {
				continue
			}
			names = append(names, strings.TrimSuffix(n, ".tmpl"))
		}
		if len(names) == 0 {
			fmt.Println("(no templates yet — try 'hive config template new <name>')")
			return nil
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	},
}

func init() {
	configTemplateNewCmd.Flags().String("from", "", "Seed from another template name (preset or custom). Examples: --from compact, --from mine")
	configTemplateCmd.AddCommand(
		configTemplateNewCmd,
		configTemplateEditCmd,
		configTemplateListCmd,
	)
}

// templatePathForName builds the on-disk path for a template name. The name
// must be a simple basename (no slashes, no path traversal) and must not
// collide with a built-in preset name — the resolver would shadow the file
// with the preset, leaving the file silently unused. The .tmpl extension
// is added if missing.
func templatePathForName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("template name cannot be empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return "", fmt.Errorf("template name %q must not contain path separators", name)
	}
	if name == "." || name == ".." || strings.Contains(name, "..") {
		return "", fmt.Errorf("template name %q is not allowed", name)
	}
	base := name
	if strings.HasSuffix(base, ".tmpl") {
		base = strings.TrimSuffix(base, ".tmpl")
	}
	if config.IsReservedPresetName(base) {
		return "", fmt.Errorf("name %q is reserved (it's a built-in preset); pick a different name", base)
	}
	return filepath.Join(templatesDir(), base+".tmpl"), nil
}

// templatesDir returns the directory where named templates live.
func templatesDir() string {
	return filepath.Join(os.Getenv("HOME"), ".hive", "templates")
}

// templateSeedContent returns the initial content for a new template file.
// `from` is a template name (preset or custom). Empty `from` returns the
// default field-reference comment block.
//
// Resolution order for `from`:
//
//  1. Search all preset libraries. If a preset with that name is found in
//     exactly one library, use it. If it appears in multiple libraries
//     (e.g., "default" exists for both human_format and tmux_format),
//     return a disambiguation error.
//  2. Otherwise, look up `~/.hive/templates/<from>.tmpl`.
//  3. Otherwise, error.
func templateSeedContent(from string) (string, error) {
	if from == "" {
		return defaultTemplateSeedComment, nil
	}
	if err := config.ValidateTemplateName(from); err != nil {
		return "", fmt.Errorf("--from %q: %w", from, err)
	}
	content, foundIn, presetErr := config.LookupPresetByName(from)
	if presetErr == nil {
		if len(foundIn) > 1 {
			return "", fmt.Errorf("--from %q is ambiguous: a preset named %q exists in %s; create the template with a key-specific source via 'hive config preset <key> %s > ~/.hive/templates/<name>.tmpl' followed by 'hive config template edit <name>'", from, from, strings.Join(foundIn, " and "), from)
		}
		return content, nil
	}
	// Try as a custom template
	path := filepath.Join(templatesDir(), from+".tmpl")
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}
	if os.IsNotExist(err) {
		return "", fmt.Errorf("--from %q: not a known preset and no template at %s", from, path)
	}
	return "", fmt.Errorf("--from %q: read %s: %w", from, path, err)
}

const defaultTemplateSeedComment = `{{- /*
  Hive status template — Go text/template syntax.
  https://pkg.go.dev/text/template

  Data fields:
    .Count (int)              total queue size
    .Next (object | nil)      head of the queue; nil when .Count == 0
      .Next.Session (string)
      .Next.Message (string)
      .Next.Pane    (string)
      .Next.Age     (string)
    .Queue (array)            all entries; same shape as .Next per entry

  Helper functions:
    add a b                   integer addition
    bold v                    ANSI bold (no-op when piped to file/pipe)
    dim  v                    ANSI dim  (no-op when piped to file/pipe)
    plus text/template built-ins: if / range / eq / gt / len / slice / printf
*/ -}}
`

// editTemplateFile opens path in $EDITOR, validates the parsed template on
// save, retries on parse error. If onSuccess is non-nil it's called after
// the first successful save.
func editTemplateFile(path string, onSuccess func() error) error {
	for {
		if err := openInEditor(path); err != nil {
			return err
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		_, parseErr := template.New("test").Funcs(config.TemplateFuncs()).Parse(string(content))
		if parseErr == nil {
			if err := stripErrorComments(path); err != nil {
				return err
			}
			if onSuccess != nil {
				if err := onSuccess(); err != nil {
					return err
				}
			}
			fmt.Printf("template saved: %s\n", path)
			return nil
		}
		if err := prependErrorComment(path, parseErr); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "template invalid: %v\n", parseErr)
		if !confirmRetry() {
			return fmt.Errorf("aborted; %s left as-is", path)
		}
	}
}

// currentStringValue reads the current value of a templated string key
// from cfg.
func currentStringValue(cfg *config.Config, key string) string {
	switch key {
	case "status.human_format":
		return cfg.Status.HumanFormat
	case "status.tmux_format":
		return cfg.Status.TmuxFormat
	}
	return ""
}

// homeRelativePath converts an absolute path under $HOME to its ~/...
// equivalent, so command output and yaml stay portable.
func homeRelativePath(p string) string {
	home := os.Getenv("HOME")
	if home == "" || !strings.HasPrefix(p, home+string(filepath.Separator)) {
		return p
	}
	return filepath.Join("~", p[len(home)+1:])
}
