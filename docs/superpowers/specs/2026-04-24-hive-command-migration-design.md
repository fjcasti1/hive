# Hive Command Migration Design

**Goal:** Incrementally migrate all hive CLI commands from the chezmoi dotfiles source into the standalone `fjcasti1/hive` repo, one command per PR, simplest to most complex. Includes a `cmd/` restructure, DEVELOPMENT.md guide, and Makefile.

**Architecture:** Foundation PR consolidates all internal packages and restructures `main.go` into a `cmd/` package alongside `history` (the first and simplest command). Each subsequent PR adds one command. Commands are ordered by number of packages touched and whether they read or write state.

**Source:** `/Users/kikocastillo/.local/share/chezmoi/hive/`
**Target:** `/Users/kikocastillo/coding/personal/hive/`

---

## Internal Packages

All four packages are ported in the Foundation PR:

| Package | Path | Responsibility |
|---------|------|---------------|
| config | `internal/config/` | Config file management |
| db | `internal/db/` | SQLite queue, history, settings |
| notify | `internal/notify/` | macOS and tmux notifications |
| session | `internal/session/` | Tmux session management |

---

## Foundation PR (PR 1)

Bundles the `cmd/` restructure with `history` — the first command — so the refactor is never shipped without a concrete command to justify it.

**Changes:**
- `main.go` becomes a minimal entrypoint that calls `cmd.Execute()`
- New `cmd/root.go` holds the cobra root command and version template
- New `cmd/history.go` implements the history command
- All four internal packages ported from chezmoi source
- `DEVELOPMENT.md` added at repo root
- `Makefile` added at repo root

---

## Command Migration Order

Each command is its own PR after the Foundation PR.

| PR | Command | Packages touched | Operations |
|----|---------|-----------------|------------|
| 2 | `notify` | notify | write (send notification) |
| 3 | `list` | db | read |
| 4 | `next` | db | read (queue pop) |
| 5 | `ack` | db | write |
| 6 | `snooze` | db | write + timing |
| 7 | `peek` | session | read |
| 8 | `status` | db + session | read |
| 9 | `pause` | session | write |
| 10 | `resume` | session | write + state transition |
| 11 | `config` | config | read/write |
| 12 | `install` | filesystem | setup (completions, config files) |
| 13 | `doctor` | all packages | diagnostic read |

**Rationale:** db-only commands first (testable without tmux), then session reads, then session writes, then meta commands that span multiple packages.

---

## DEVELOPMENT.md

Saved at repo root. Covers both personal use and contributors.

```markdown
# Development

## Prerequisites
- Go 1.26+
- tmux
- macOS (for notify features)

## Quick Start
make build          # compile to ./hive
make run            # go run . (fast iteration, no compile)
make install-dev    # install to ~/.local/bin/hive (shadows Homebrew)

## Dev vs Prod
- `make run` / `go run .` — fastest iteration, no binary
- `make install-dev` — installs dev binary at ~/.local/bin/hive, shadows brew-installed hive
- `make uninstall-dev` — removes dev binary, falls back to Homebrew version
- `brew install fjcasti1/hive/hive` — prod install

## Testing
make test           # go test ./...
make vet            # go vet ./...
make check          # vet + test together (what CI runs)

## Release Process
Releases are fully automated. To ship:
1. Open a PR with conventional commits (feat:, fix:, chore:)
2. Merge to main → release-please opens a release PR
3. Merge the release PR → GoReleaser tags and publishes binaries + Homebrew formula

## Project Layout
cmd/          # cobra subcommands, one file per command
internal/
  config/     # config file management
  db/         # sqlite queue, history, settings
  notify/     # macOS and tmux notifications
  session/    # tmux session management
```

---

## Makefile

```makefile
.PHONY: build run install-dev uninstall-dev test vet check clean

build:
	go build -o hive .

run:
	go run .

install-dev:
	go build -o ~/.local/bin/hive .

uninstall-dev:
	rm -f ~/.local/bin/hive

test:
	go test ./...

vet:
	go vet ./...

check: vet test

clean:
	rm -f hive
```

---

## Out of Scope

- Deleting `hive/` from the chezmoi dotfiles repo (Phase 3, after all commands migrated and validated)
- Updating the Homebrew tap README (separate cleanup task)
- Windows support
