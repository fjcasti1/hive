# Development Guide

## Prerequisites

- **Go 1.26+** — `go version`
- **tmux** — required at runtime for session commands
- **macOS** — required at runtime for `notify` (osascript); all other commands work on Linux too
- **gh** — GitHub CLI, used for opening PRs

## Getting Started

```bash
git clone git@github.com:fjcasti1/hive.git
cd hive
go mod download
```

Verify the build:

```bash
go build ./...
```

## Running Locally

### Fastest iteration — no compilation step

```bash
go run . --version
go run . notify --help
go run . notify -m "needs review"
```

### Build a local binary

```bash
go build -o hive .
./hive --version
rm hive  # clean up
```

### Install dev binary (shadows the Homebrew release)

```bash
go build -o ~/.local/bin/hive .
```

`~/.local/bin` is earlier in `$PATH` than `/opt/homebrew/bin`, so this shadows the Homebrew-installed `hive` binary. Useful when you want `hive` to refer to your local build across all terminal sessions.

To fall back to the Homebrew version:

```bash
rm ~/.local/bin/hive
```

### Build with an explicit version string

```bash
go build -ldflags "-X github.com/fjcasti1/hive/cmd.version=1.2.3-dev" -o hive .
./hive --version  # prints: hive 1.2.3-dev
```

In production GoReleaser injects the real tag via `-X main.Version={{.Version}}` automatically.

## Testing

```bash
go test ./...          # run all tests
go test ./internal/... # run only internal package tests
go test -v ./...       # verbose output
go test -run TestFoo ./internal/db/...  # run a specific test
```

Tests use an in-memory SQLite database opened by a private `openMem` helper in `internal/db/db_test.go`, so no real database is created or modified.

## Linting

```bash
go vet ./...
```

This is what CI runs. Fix all vet warnings before opening a PR.

## What CI Checks

Every PR must pass the `build` job in `.github/workflows/ci.yml`:

```
go build ./...
go vet ./...
```

Branch protection requires this check to pass before merging.

## Project Layout

```
main.go               # entrypoint — calls cmd.Execute()
cmd/
  root.go             # cobra root command, SetVersion, Execute
  notify.go           # `hive notify` subcommand
internal/
  config/             # config file (~/.hive/config/config.yaml)
  db/                 # SQLite queue + embedded migrations
```

## Adding a New Command

1. Create `cmd/<name>.go` with a `var <name>Cmd = &cobra.Command{...}` and its `init()` for flags.
2. Register it in `cmd/root.go` inside `init()`:
   ```go
   rootCmd.AddCommand(<name>Cmd)
   ```
3. Build and smoke-test:
   ```bash
   go build ./... && go run . <name> --help
   ```
4. Open a PR. The `build` CI check must pass before merging.

## Data Files

At runtime hive keeps everything under `~/.hive/`:

| File | Path |
|------|------|
| Config | `~/.hive/config/config.yaml` |
| Database | `~/.hive/db/hive.db` |

Paths are computed from `$HOME` in `internal/config/config.go` and `internal/db/db.go`. Tests use in-memory SQLite via a test-local `openMem` helper so they never touch real disk.

## Database & Migrations

Schema changes are managed by [pressly/goose](https://github.com/pressly/goose). Migration files live in `internal/db/migrations/` and are embedded into the binary at compile time via `//go:embed migrations/*.sql` in `internal/db/db.go`. They run automatically inside `db.Open()`, so any `hive` invocation that touches the DB also brings the schema up to date. Goose tracks applied migrations in its own `goose_db_version` table — you should not edit that table by hand.

### Adding a migration

Scaffold one with the goose CLI:

```bash
go run github.com/pressly/goose/v3/cmd/goose@latest \
    -dir internal/db/migrations create <name> sql
```

That creates `internal/db/migrations/<NNNNN>_<name>.sql` with `Up`/`Down` stubs already in place.

You can also hand-write the file. Convention: zero-padded sequence number, underscore, snake-cased description, `.sql`:

```
internal/db/migrations/00002_add_priority.sql
```

Each file has both directions, separated by goose's marker comments:

```sql
-- +goose Up
ALTER TABLE queue ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE queue DROP COLUMN priority;
```

Always fill in `Down`. It's what makes `goose down` and `goose redo` work during development, and forces you to think about reversibility while the change is fresh.

### Working with the goose CLI

Useful when iterating on a migration. All commands target the real DB at `~/.hive/db/hive.db`:

```bash
alias gx='go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/db/migrations sqlite3 ~/.hive/db/hive.db'

gx status   # applied vs pending
gx up       # apply all pending (also runs on hive startup)
gx down     # roll back the most recent migration
gx redo     # down + up — handy when tweaking the migration you just wrote
gx reset    # roll back everything (destructive)
```

You don't need the CLI to ship migrations — the embedded files run on app startup regardless. The CLI is purely for local inspection and rollback.

### Verifying a migration

`internal/db/db_test.go` opens an in-memory DB, runs migrations, and checks the schema. Run it on every migration change:

```bash
go test ./internal/db/
```

For richer checks (e.g., that a new column has the expected default), extend that test rather than writing a one-off script.

### Resetting your local DB

If the local DB lands in a weird state — half-applied migration, manual edits, mystery — nuke it:

```bash
rm ~/.hive/db/hive.db
```

The next `hive` run recreates it from migrations. Safe in dev: there's nothing in there but your own queue and history.

## Release Process

Releases are fully automated — you never tag manually.

1. Open a PR with [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore:`, etc.)
2. Merge to `main` → release-please opens a versioned release PR with an updated `CHANGELOG.md`
3. Merge the release PR → GoReleaser builds cross-platform binaries, publishes a GitHub release, and pushes the Homebrew formula to `fjcasti1/homebrew-hive`

### Installing the released binary

```bash
brew install fjcasti1/hive/hive  # tap-qualified (avoids conflict with Apache Hive)
hive --version
```
