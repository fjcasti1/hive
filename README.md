# hive

Manage multiple AI agent sessions in tmux. Agents call `hive notify` when they need attention; you get an alert and a queue entry in the status bar. Navigate to the waiting session, give feedback, and the queue clears automatically.

Designed for [sesh](https://github.com/joshmedeski/sesh) users running Claude Code, Codex, or any other agent in separate tmux sessions.

## How it works

```
Agent finishes a task
  → calls hive notify
    → macOS notification + tmux bell fires
    → session added to queue
      → status bar shows 🐝 my-session: tests done (2m)
        → you press prefix+A to jump there (or prefix+a for a picker)
          → you reply or grant permission
            → hive ack fires automatically
              → session removed from queue
```

## Installation

**Homebrew (recommended):**
```bash
brew tap fjcasti1/hive
brew install hive
```

**Go:**
```bash
go install github.com/fjcasti1/hive@latest
```

Then set up integrations:
```bash
hive install   # Claude hooks, tmux snippet, shell completion
hive doctor    # verify everything is wired up
```

## Commands

| Command | Description |
|---|---|
| `hive notify [-m "message"]` | Add current session to queue and send notifications |
| `hive ack [session-or-index]` | Acknowledge — mark feedback given, move to history |
| `hive next` | Switch to the next waiting session (FIFO); `--show` prints it without switching, `--ack` acknowledges after switching |
| `hive status` | Show queue state (default human; `--format=tmux` or `--format=json` available) |
| `hive list` | _Deprecated — use `hive status`_ |
| `hive snooze [session] [duration]` | Hide a session for a while (e.g. `10m`, `1h`) |
| `hive pause [duration]` | Suppress all notifications (agents still queue) |
| `hive resume` | Re-enable notifications |
| `hive history` | Show recently resolved sessions |
| `hive config show` | Print current config as YAML |
| `hive config set <key> <value>` | Update a config value |
| `hive install` | Set up all integrations |
| `hive doctor` | Check all integrations |
| `hive --version` | Print version |

## Configuration

Config lives at `~/.hive/config.yaml`. The file appears the first time you customize something — running `hive` with no config in place uses the built-in defaults silently. Four ways to manage it:

```bash
hive config show                                 # print the current effective configuration
hive config set <key> <value>                    # set one key, persist
hive config reset <key>                          # restore one key to its default
hive config edit                                 # open the file in $EDITOR with validation
```

Common settings:

```bash
hive config set notifications.macos false        # disable macOS popups
hive config set notifications.tmux_bell false    # disable tmux bell
hive config set queue.max_message_length 80      # truncate long messages
hive config set history.retention_days 14        # days to keep history (0 disables history)
```

`history.retention_days` controls how long resolved sessions are retained. The purge runs on every `hive` invocation that opens the database, deleting entries with a `resolved_at` older than the cutoff. Setting it to `0` makes the cutoff "now," so every invocation wipes the history table — effectively disabling history.

### Status format templates

`status.human_format` and `status.tmux_format` are Go [`text/template`](https://pkg.go.dev/text/template) strings rendered against the queue. The default human format produces ANSI-bold/dim output when `hive status` is connected to a terminal; pipes and redirects automatically receive plain text.

For the common cases, use a built-in **preset** instead of writing your own template — prefix the value with `@`:

```bash
hive config set status.human_format @compact     # one-line summary
hive config set status.human_format @verbose     # multi-line with pane info
hive config set status.human_format @default     # back to the shipped default
hive config set status.tmux_format  @minimal     # tighter status-bar output
```

`hive config set status.human_format @bogus` will fail and list the available preset names.

If you want to write your own, available template fields are:

| Field | Type | Description |
|---|---|---|
| `.Count` | int | Total queue size |
| `.Next` | object \| nil | Head of the queue; nil when `.Count == 0` |
| `.Next.Session` | string | Session name |
| `.Next.Message` | string | Agent's message (may be empty) |
| `.Next.Pane` | string | tmux pane id (e.g., `%5`) |
| `.Next.Age` | string | Pre-formatted age (e.g., `2m`) |
| `.Queue` | array | Full queue, each entry with the same shape as `.Next` |

Helper functions registered on top of `text/template` built-ins: `add a b`, `bold v`, `dim v`. (`bold` and `dim` emit ANSI when stdout is a TTY, plain text otherwise.) The `slice`, `len`, `printf`, `if`/`else`, and `range` you'd expect from text/template all work too. Use conditionals to collapse punctuation when a field is empty, e.g. `{{ if .Next.Message }}: {{ .Next.Message }}{{ end }}`.

If you break a template, `hive config edit` reopens the file with the validation error as a comment so you can fix it. As a last resort, `hive config reset status.human_format` restores the shipped default.

JSON output (`hive status --format=json`) is fixed-schema and not configurable.

## Claude Code integration

`hive install` writes these hooks into `~/.claude/settings.json`:

| Event | Action |
|---|---|
| `Stop`, `StopFailure`, `Notification`, `Elicitation` | `hive notify` |
| `UserPromptSubmit`, `PostToolUse`, `ElicitationResult`, `SessionEnd` | `hive ack` |

Agents can also pass context:

```bash
hive notify -m "tests passing, ready for review"
hive notify -m "blocked: need API key"
```

## tmux integration

`hive install` prints a config snippet. Key bindings:

| Binding | Action |
|---|---|
| `prefix + A` | Jump to next waiting session |
| `prefix + a` | Picker — fzf list of waiting sessions, Enter to jump |

The status bar polls `hive status --format=tmux` every 5 seconds. The default tmux template produces output like `🐝 my-session: tests done (2m ago) | +1`; customize via `hive config set status.tmux_format "..."`.

## Contributing

Commits should follow the [Conventional Commits](https://www.conventionalcommits.org) format — this drives automatic changelog generation and version bumping via release-please:

```
feat: add snooze command
fix: correctly parse RFC3339 timestamps from sqlite
chore: update dependencies
docs: improve README installation steps
```

Releases are automated: merge a Release PR created by release-please → tag is created → GoReleaser builds binaries and updates the Homebrew formula.

## Building from source

```bash
git clone https://github.com/fjcasti1/hive
cd hive
go build -ldflags "-X main.Version=0.1.0" -o ~/.local/bin/hive .
```
