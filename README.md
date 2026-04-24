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
| `hive next` | Switch to the next waiting session (FIFO) |
| `hive peek` | Show next waiting session without switching |
| `hive list` | Show all waiting sessions |
| `hive list --watch` | Live-refreshing view |
| `hive snooze [session] [duration]` | Hide a session for a while (e.g. `10m`, `1h`) |
| `hive pause [duration]` | Suppress all notifications (agents still queue) |
| `hive resume` | Re-enable notifications |
| `hive status` | tmux status bar output |
| `hive history` | Show recently resolved sessions |
| `hive config show` | Print current config as YAML |
| `hive config set <key> <value>` | Update a config value |
| `hive install` | Set up all integrations |
| `hive doctor` | Check all integrations |
| `hive --version` | Print version |

## Configuration

Config lives at `~/.config/hive/config.yaml`. Use `hive config set` to edit:

```bash
hive config set notifications.macos false        # disable macOS popups
hive config set notifications.tmux_bell false    # disable tmux bell
hive config set queue.max_message_length 80      # truncate long messages
hive config set status.format "{session} ({age}) | +{extra}"  # custom status bar
hive config set list.watch_interval 3            # refresh rate for --watch
hive config set history.retention_days 14        # how long to keep history
hive config set snooze.default_duration 15m      # default snooze length
hive config set pause.default_duration 2h        # default pause length (empty = indefinite)
```

### Status bar tokens

The `status.format` string supports these tokens:

| Token | Value |
|---|---|
| `{session}` | Session name |
| `{message}` | Agent's message (collapsed with separator when empty) |
| `{age}` | How long the session has been waiting |
| `{count}` | Total sessions in queue |
| `{extra}` | Sessions beyond the first (collapsed when 0) |

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

The status bar polls `hive status` every 5 seconds. When paused, it shows `⏸ N waiting` instead.

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
