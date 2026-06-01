# Status Templates

`hive status` renders queue state via two templated config keys:

- **`status.human_format`** — what `hive status` (or `--format=human`) prints to a terminal. ANSI bold/dim when stdout is a TTY, plain text when piped.
- **`status.tmux_format`** — what `hive status --format=tmux` produces, suitable for embedding in tmux's `status-right`.

Both are stored in `~/.hive/config.yaml` as a single bare name:

```yaml
status:
    human_format: default      # built-in preset
    tmux_format:  minimal      # built-in preset
    human_format: example      # custom template at ~/.hive/templates/example.tmpl
```

The resolver checks the preset library for the key first; if the name isn't a preset, it looks for `~/.hive/templates/<name>.tmpl`. JSON output (`--format=json`) is fixed-schema and not templated.

## Quick start

Pick a preset:

```bash
hive config set status.human_format compact
hive config set status.tmux_format  minimal
```

Author your own:

```bash
hive config template new mine            # creates + opens ~/.hive/templates/mine.tmpl in $EDITOR
hive config set status.human_format mine # point the key at it
```

Inspect:

```bash
hive config presets                              # list every preset name across both keys
hive config preset status.human_format default   # print one preset to stdout
hive config template list                        # list custom templates in ~/.hive/templates/
```

Edit later:

```bash
hive config template edit mine
```

## Built-in presets

Names are reserved across both keys — `template new compact` is rejected so a custom file can never silently shadow a preset.

### `status.human_format`

#### `default`

Multi-line summary with a marker for the next session and indented additional entries.

```gotmpl
{{- if eq .Count 0 -}}
🐝 No agents waiting
{{- else -}}
🐝 {{ bold .Count }} agent{{ if gt .Count 1 }}s{{ end }} waiting

  ▸ {{ bold .Next.Session }}{{ if .Next.Message }} — {{ .Next.Message }}{{ end }} {{ dim (printf "(%s)" .Next.Age) }}
{{- range slice .Queue 1 }}
    {{ .Session }}{{ if .Message }} — {{ .Message }}{{ end }} {{ dim (printf "(%s)" .Age) }}
{{- end }}
{{- end -}}
```

Renders:

```
🐝 3 agents waiting

  ▸ alpha — tests passing (2m)
    beta — blocked (1m)
    gamma (30s)
```

#### `compact`

One line. Useful when you have other status output you want to keep concise.

```gotmpl
{{- if eq .Count 0 -}}🐝 idle{{- else -}}🐝 {{ bold .Count }}{{ if .Next }} → {{ bold .Next.Session }}{{ if .Next.Message }}: {{ .Next.Message }}{{ end }} {{ dim (printf "(%s)" .Next.Age) }}{{ end }}{{- end -}}
```

Renders:

```
🐝 3 → alpha: tests passing (2m)
```

#### `verbose`

Shows every entry, every field, including pane id.

```gotmpl
{{- if eq .Count 0 -}}
🐝 No agents waiting
{{- else -}}
🐝 {{ bold .Count }} agent{{ if gt .Count 1 }}s{{ end }} waiting

{{ range $i, $e := .Queue }}{{ if eq $i 0 }}  ▸ {{ bold $e.Session }}{{ else }}    {{ $e.Session }}{{ end }}{{ if $e.Message }} — {{ $e.Message }}{{ end }} {{ dim (printf "(pane %s, %s)" $e.Pane $e.Age) }}
{{ end -}}
{{- end -}}
```

Renders:

```
🐝 3 agents waiting

  ▸ alpha — tests passing (pane %5, 2m)
    beta — blocked (pane %6, 1m)
    gamma (pane %7, 30s)
```

### `status.tmux_format`

Tmux templates use tmux's escape syntax (`#[fg=colour220,bold]`). They typically end with a trailing space so they separate cleanly from neighboring `status-right` widgets.

#### `default`

```gotmpl
{{- if .Next -}}#[fg=colour220,bold]🐝 {{ .Next.Session }}{{ if .Next.Message }}: {{ .Next.Message }}{{ end }} ({{ .Next.Age }}){{ if gt .Count 1 }} | +{{ len (slice .Queue 1) }}{{ end }}#[fg=default,nobold] {{ end -}}
```

Renders (in tmux): `🐝 alpha: tests passing (2m) | +2 ` (bold yellow bee + session, default-color message + age + extras count)

#### `minimal`

```gotmpl
{{- if .Next -}}🐝 {{ .Next.Session }}{{ if gt .Count 1 }} +{{ len (slice .Queue 1) }}{{ end }} {{ end -}}
```

Renders: `🐝 alpha +2 `

#### `verbose`

Adds dim color for the message and age, plus a `more` label for the extras count.

```gotmpl
{{- if .Next -}}#[fg=colour220,bold]🐝 {{ .Next.Session }}#[fg=default,nobold]{{ if .Next.Message }}: #[fg=colour245]{{ .Next.Message }}#[fg=default]{{ end }} #[fg=colour245]({{ .Next.Age }}){{ if gt .Count 1 }} | +{{ len (slice .Queue 1) }} more{{ end }}#[fg=default] {{ end -}}
```

## Template syntax

Templates are [Go `text/template`](https://pkg.go.dev/text/template) strings. Three concepts cover almost everything:

### Actions: `{{ ... }}`

Anything between `{{` and `}}` is an action — substitution, conditional, iteration, or function call. Outside the braces is literal text.

| Form | What it does |
|---|---|
| `{{ .Field }}` | Print a field of the current context |
| `{{ if EXPR }}A{{ else }}B{{ end }}` | Conditional |
| `{{ range .Slice }}...{{ end }}` | Iterate |
| `{{ functionName arg1 arg2 }}` | Call a function (prefix style) |

### The dot: `.`

`.` is the current context — a `*statusData` struct at the top level. Inside a `range`, `.` rebinds to the current element. To reach the outer scope from inside a `range`, use `$`.

### Whitespace trimming: `{{-` and `-}}`

By default every space and newline in the source is preserved. Trim with hyphens:

| Form | Effect |
|---|---|
| `{{- ` | Trim whitespace *before* this action |
| ` -}}` | Trim whitespace *after* this action |

Use these to suppress newlines around control-flow actions so the output doesn't have stray blank lines.

## Data fields

Templates have access to a single root context with these fields:

| Field | Type | Description |
|---|---|---|
| `.Count` | int | Total queue size |
| `.Next` | object \| nil | Head of the queue; `nil` when `.Count == 0` |
| `.Next.Session` | string | Session name |
| `.Next.Message` | string | Agent's message (may be empty) |
| `.Next.Pane` | string | tmux pane id (e.g., `%5`) |
| `.Next.Age` | string | Pre-formatted age (e.g., `2m`, `30s`) |
| `.Queue` | `[]Entry` | Every entry in the queue, including the head |

Each entry in `.Queue` has the same shape as `.Next` (Session, Message, Pane, Age).

## Helper functions

In addition to text/template's built-ins, hive registers:

| Function | Signature | Effect |
|---|---|---|
| `add` | `add a b` (ints) | Returns `a + b`. Useful for 1-indexed display: `{{ add $i 1 }}`. |
| `bold` | `bold v` | Wraps `v` in ANSI bold escape codes when stdout is a TTY; passes through unchanged otherwise. |
| `dim` | `dim v` | Wraps `v` in ANSI dim escape codes when stdout is a TTY; passes through otherwise. |

text/template built-ins you'll commonly use:

- `eq`, `ne`, `lt`, `le`, `gt`, `ge` — comparisons
- `and`, `or`, `not` — boolean
- `len` — length of a slice/string
- `slice` — `slice .Queue 1` returns `.Queue[1:]`
- `index` — `index .Queue 0` returns `.Queue[0]`
- `printf` — string formatting
- `range`, `if`, `else`, `with`, `end` — control flow

## Common patterns

### Collapse surrounding punctuation when empty

The trick: wrap both the separator and the field in the same `if`:

```gotmpl
{{ .Next.Session }}{{ if .Next.Message }} — {{ .Next.Message }}{{ end }}
```

Output is `alpha — tests passing` when the message is set, `alpha` when it isn't. No stray ` — `.

### "X more in queue"

Without arithmetic operators, use `slice` + `len`:

```gotmpl
{{ if gt .Count 1 }}{{ len (slice .Queue 1) }} more{{ end }}
```

(Equivalent to `.Count - 1` when `.Count > 1`.)

### 1-indexed iteration

```gotmpl
{{ range $i, $e := .Queue }}
[{{ add $i 1 }}] {{ $e.Session }}
{{- end }}
```

### Singular/plural

```gotmpl
{{ .Count }} agent{{ if gt .Count 1 }}s{{ end }} waiting
```

## Authoring a custom template

```bash
# 1. Pick a name (anything that isn't a reserved preset name).
hive config template new mine

# 2. $EDITOR opens with a comment block listing the data fields and helpers.
#    Replace it with your template. Save.

# 3. Wire it up.
hive config set status.human_format mine

# 4. Test.
hive status
```

Variations:

```bash
# Seed from a built-in preset:
hive config template new mine --from compact

# Seed from another custom template:
hive config template new other --from mine

# Re-open later:
hive config template edit mine
```

If your template doesn't parse on save, `hive config template new`/`edit` re-opens it with the error as a `# ERROR:` comment at the top. Fix and save again.

## Troubleshooting

**`config: status.human_format: template "X" not found ...`** — `X` isn't a preset for that key, and no `~/.hive/templates/X.tmpl` exists. Either `hive config template new X` or pick a preset name from the error's hint.

**`name "default" is reserved (it's a built-in preset)`** — pick a name that isn't any of `default`, `compact`, `verbose`, `minimal`. Reserved names are listed by `hive config presets`.

**`--from "default" is ambiguous`** — `default` is a preset for both human and tmux formats. The error message tells you the workaround: pipe a specific preset to a file via `hive config preset <key> default > ~/.hive/templates/mine.tmpl` and then `hive config template edit mine`.

**Reset to ship defaults** — `hive config reset status.human_format` restores the bare-name reference (e.g., `default`).
