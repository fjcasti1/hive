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

Config lives at `~/.hive/config.yaml`. The file appears the first time you customize something — running `hive` with no config in place uses the built-in defaults silently.

```bash
hive config show                                 # print the current effective configuration
hive config set <key> <value>                    # set one key, persist
hive config reset <key>                          # restore one key to its default
hive config edit                                 # open the YAML file in $EDITOR with validation
hive config presets                              # list built-in template presets
hive config preset <key> <name>                  # print one preset to stdout

hive config template new <name>                  # create ~/.hive/templates/<name>.tmpl, open in $EDITOR
hive config template edit <name>                 # open ~/.hive/templates/<name>.tmpl in $EDITOR
hive config template list                        # list templates in ~/.hive/templates/
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

`status.human_format` and `status.tmux_format` are stored in YAML as **bare names** — either a built-in preset or a custom template at `~/.hive/templates/<name>.tmpl`:

```yaml
status:
    human_format: default      # built-in preset
    tmux_format:  minimal      # built-in preset
    human_format: example      # custom template at ~/.hive/templates/example.tmpl
```

```bash
hive config set status.human_format compact     # built-in compact preset
hive config template new mine                   # create + edit a custom template
hive config set status.human_format mine        # use the custom template
```

For preset listings, custom-template authoring, the available data fields, helper functions, and template syntax: **see [TEMPLATES.md](./TEMPLATES.md)**.

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
