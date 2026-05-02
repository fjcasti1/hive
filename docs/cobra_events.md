# Cobra Lifecycle and Events

## Overview

A Cobra program runs in two distinct phases:

1. **Go package init** — every package's `init()` function runs once, in import-dependency order, before `main()` starts. This is plain Go, not Cobra. In a Cobra app it is where you build the command tree: register subcommands with `rootCmd.AddCommand(...)` and define flags with `cmd.Flags().StringVar(...)`. No flag values are available yet, no command has been resolved, and `Execute()` has not been called.

2. **Cobra runtime lifecycle** — begins when `main()` calls `rootCmd.Execute()`. Cobra parses `os.Args`, walks the command tree to resolve the target command, validates flags, and then fires a fixed sequence of hooks around the command's `Run` function. This is the part most "lifecycle" docs are talking about.

Understanding the order matters because it dictates **where** to put each kind of work:

| Concern | Where it belongs |
| --- | --- |
| Wiring subcommands and flags | `init()` |
| Reading config files / env vars | `cobra.OnInitialize` |
| Opening shared resources (DB, HTTP client) | `PersistentPreRunE` on root |
| The actual command's work | `RunE` |
| Tearing those resources down | `PersistentPostRunE` on root, **plus** `defer` for error paths |

> ⚠️ Never call `rootCmd.Execute()` from `init()`. `init()` runs at package-load time, before `main()`. If `Execute()` is in there, your CLI starts running before `main()` does — and runs again when `main()` calls it. Keep `init()` for tree construction only.

## The hook reference

There are two registration sites for lifecycle code:

- **Package-level**: `cobra.OnInitialize(fn ...func())`. Functions registered here are global to the program — they fire once per `Execute()`, before any command's hooks. They take no arguments, so you have no `*cobra.Command` context (no flag values via `cmd.Flags()`, no args). Best for things that don't depend on which command was invoked: loading a config file, setting up a logger.

- **Per-command**: fields on `*cobra.Command`. Each has a non-error and an error-returning variant; the E suffix returns `error`, which Cobra propagates out of `Execute()`. **If both `Foo` and `FooE` are set on the same command, only `FooE` runs.** Pick one form per command.

| Field | Inherits to children? | When it fires |
| --- | --- | --- |
| `PersistentPreRun(E)` | Yes | Before `PreRun` of the resolved command |
| `PreRun(E)` | No (only the exact command) | Right before `Run` |
| `Run(E)` | No | The command's actual work |
| `PostRun(E)` | No | Right after `Run` (only on success — see caveat) |
| `PersistentPostRun(E)` | Yes | After `PostRun` |

### "Nearest one wins" for persistent hooks

Persistent hooks **do not chain up the tree**. When a leaf command runs, Cobra walks from that command toward the root looking for the first non-nil `PersistentPreRunE`, runs it, and stops. Same for `PersistentPostRunE`.

```
root      PersistentPreRunE = A
└─ sub    PersistentPreRunE = B
   └─ leaf   (no hook)
```

Running `app sub leaf` fires `B` only. `A` is shadowed. If you want both, `B` has to call `A` explicitly:

```go
subCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
        return err
    }
    return loadSubResources()
}
```

This is the most common Cobra footgun. If a subtree's hook isn't firing, check whether an ancestor hook is shadowing it — or vice versa.

## Execution order, end to end

For `hive notify --foo=bar`:

1. Go runtime runs every package's `init()` (subcommands registered, flags defined).
2. `main()` calls `cmd.Execute()` which calls `rootCmd.Execute()`.
3. Cobra parses `os.Args`, walks the tree to resolve `notifyCmd`, and validates flags.
4. **`cobra.OnInitialize` callbacks** fire, in the order they were registered. No `*cobra.Command` context.
5. **`PersistentPreRunE`** — nearest non-nil walking from `notifyCmd` toward root.
6. **`PreRunE`** — only if set on `notifyCmd`.
7. **`RunE`** — `notifyCmd.RunE`. If it returns an error, jump to step 10.
8. **`PostRunE`** — only if set on `notifyCmd`.
9. **`PersistentPostRunE`** — nearest non-nil walking from `notifyCmd` toward root.
10. `Execute()` returns. If any `*E` hook returned an error, that error is the return value; otherwise `nil`.

## Error semantics — the cleanup gotcha

If `RunE` (or any `*PreRunE`) returns an error, **`PostRunE` and `PersistentPostRunE` are skipped**. The error short-circuits the rest of the lifecycle.

This means **post-run hooks are not safe places to put cleanup that must always run**. If you open a database in `PersistentPreRunE` and close it in `PersistentPostRunE`, an error in `RunE` leaks the connection until process exit (which usually doesn't matter for a CLI, but matters for daemons, tests, and anything writing files).

Two reliable patterns for guaranteed cleanup:

```go
// Option 1: defer inside RunE
RunE: func(cmd *cobra.Command, args []string) error {
    db, err := openDB()
    if err != nil { return err }
    defer db.Close()
    return doWork(db)
}

// Option 2: defer in main, around Execute
func main() {
    defer cleanup()
    cmd.Execute()
}
```

Use `PersistentPostRunE` for things that are nice-to-have on the success path (printing a summary, flushing telemetry on clean exit) — not for releasing resources whose lifetime must match the process.

## Practical guidance for this project

- **Config**: `cobra.OnInitialize(loadConfig)` in `cmd/root.go`. Runs once, doesn't need `*cobra.Command`.
- **DB / clients**: `rootCmd.PersistentPreRunE`. You have access to `cmd.Flags()` so config overrides via flags work. Inherits to every subcommand by default.
- **Skipping setup for some subcommands**: commands like `version`, `config print`, or anything that shouldn't touch the DB need to override the inherited setup. Set their own `PersistentPreRunE` to a no-op (or one that does only the subset of work they need). `nil` is not enough — Cobra walks up to the next non-nil ancestor.
- **Cleanup**: prefer `defer` inside `RunE` or in `main()` around `Execute()`. Use `PersistentPostRunE` only for success-path-only work.
- **Errors**: prefer the `*E` variants everywhere. Returning errors lets Cobra exit with the right code and lets callers test hooks directly.
