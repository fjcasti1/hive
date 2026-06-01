# Changelog

## [1.6.0](https://github.com/fjcasti1/hive/compare/v1.5.0...v1.6.0) (2026-06-01)


### 🎁 New Features

* Add --show flag to next command ([#37](https://github.com/fjcasti1/hive/issues/37)) ([8ece0fe](https://github.com/fjcasti1/hive/commit/8ece0fed1dbafe5e94fcb82f11e8aa62e9dbceee))
* Add `hive config reset <key>` to restore a single key to its ([d0bfbfd](https://github.com/fjcasti1/hive/commit/d0bfbfdfb0f93c82a4ece61b25836a48a0574051))
* Add hive status command with configurable templates, presets, and config edit/reset ([#34](https://github.com/fjcasti1/hive/issues/34)) ([d0bfbfd](https://github.com/fjcasti1/hive/commit/d0bfbfdfb0f93c82a4ece61b25836a48a0574051))
* Create config file with defaults on first run ([#39](https://github.com/fjcasti1/hive/issues/39)) ([0ad65f8](https://github.com/fjcasti1/hive/commit/0ad65f8f3739c06eb0a5fb03b855cbd8d173e736))
* Ship built-in template presets selectable with an `[@name](https://github.com/name)` prefix ([d0bfbfd](https://github.com/fjcasti1/hive/commit/d0bfbfdfb0f93c82a4ece61b25836a48a0574051))


### 🐛 Bug Fixes

* Change config `Load` to no longer auto-write a defaults file on ([d0bfbfd](https://github.com/fjcasti1/hive/commit/d0bfbfdfb0f93c82a4ece61b25836a48a0574051))
* JSON output is a fixed, non-templated schema (`count`, `next`, ([d0bfbfd](https://github.com/fjcasti1/hive/commit/d0bfbfdfb0f93c82a4ece61b25836a48a0574051))
* Reject unknown config keys with strict YAML decoding ([#40](https://github.com/fjcasti1/hive/issues/40)) ([5fd4a80](https://github.com/fjcasti1/hive/commit/5fd4a8016538b3095654734a2f11cf24afdaf2c8))


### 📚 Documentation

* **internal:** add godocs ([8ece0fe](https://github.com/fjcasti1/hive/commit/8ece0fed1dbafe5e94fcb82f11e8aa62e9dbceee))


### 💫 Code Refactoring

* **db:** rename internal method Peek -&gt; Show ([8ece0fe](https://github.com/fjcasti1/hive/commit/8ece0fed1dbafe5e94fcb82f11e8aa62e9dbceee))


### 🧪 Tests

* Add cmd-package tests ([#38](https://github.com/fjcasti1/hive/issues/38)) ([4d2919a](https://github.com/fjcasti1/hive/commit/4d2919af7a6eb61ac386b3d5e513fa3167225be5))

## [1.5.0](https://github.com/fjcasti1/hive/compare/v1.4.0...v1.5.0) (2026-05-03)


### 🎁 New Features

* activate history purge with configuration field ([298a59a](https://github.com/fjcasti1/hive/commit/298a59a5ff54f5af8e394802ccb6d9ffaf2afd9b))
* **config:** add `hive config` commands: `set` & `show` ([#32](https://github.com/fjcasti1/hive/issues/32)) ([298a59a](https://github.com/fjcasti1/hive/commit/298a59a5ff54f5af8e394802ccb6d9ffaf2afd9b))


### 🐛 Bug Fixes

* write default configuration to file on first execution ([298a59a](https://github.com/fjcasti1/hive/commit/298a59a5ff54f5af8e394802ccb6d9ffaf2afd9b))


### 📚 Documentation

* Update development and readme guide ([298a59a](https://github.com/fjcasti1/hive/commit/298a59a5ff54f5af8e394802ccb6d9ffaf2afd9b))


### 💫 Code Refactoring

* Change `Config` to a pointer type ([298a59a](https://github.com/fjcasti1/hive/commit/298a59a5ff54f5af8e394802ccb6d9ffaf2afd9b))
* Flatten config and db paths ([298a59a](https://github.com/fjcasti1/hive/commit/298a59a5ff54f5af8e394802ccb6d9ffaf2afd9b))

## [1.4.0](https://github.com/fjcasti1/hive/compare/v1.3.0...v1.4.0) (2026-05-03)


### 🎁 New Features

* add next command to switch to oldest waiting session ([#30](https://github.com/fjcasti1/hive/issues/30)) ([78de2b7](https://github.com/fjcasti1/hive/commit/78de2b7c28dee650a650a8a4a9f5e8f78855816e))


### 🐛 Bug Fixes

* Stabilize queue ordering with `id ASC` tiebreaker when multiple ([78de2b7](https://github.com/fjcasti1/hive/commit/78de2b7c28dee650a650a8a4a9f5e8f78855816e))


### 💫 Code Refactoring

* Add `tmux.SwitchTo` helper that wraps `tmux switch-client -t ([78de2b7](https://github.com/fjcasti1/hive/commit/78de2b7c28dee650a650a8a4a9f5e8f78855816e))
* Extract `ackSession` helper from `ack.go` so both `ack` and ([78de2b7](https://github.com/fjcasti1/hive/commit/78de2b7c28dee650a650a8a4a9f5e8f78855816e))


### 🧪 Tests

* Add four table-driven DB tests for `Peek`: empty queue, single ([78de2b7](https://github.com/fjcasti1/hive/commit/78de2b7c28dee650a650a8a4a9f5e8f78855816e))

## [1.3.0](https://github.com/fjcasti1/hive/compare/v1.2.0...v1.3.0) (2026-05-03)


### 🎁 New Features

* add history command and persist acked notifications ([#28](https://github.com/fjcasti1/hive/issues/28)) ([0a0d7e7](https://github.com/fjcasti1/hive/commit/0a0d7e7f077bcf4a48bc180ab0690e4da6af4c63))
* Enable macOS and tmux-bell notifications by default in ([596e0bb](https://github.com/fjcasti1/hive/commit/596e0bba24dac09cee03aac5bc26d2e14ad185e6))
* track tmux pane in queue entries ([#26](https://github.com/fjcasti1/hive/issues/26)) ([596e0bb](https://github.com/fjcasti1/hive/commit/596e0bba24dac09cee03aac5bc26d2e14ad185e6))


### 🐛 Bug Fixes

* use RFC3339 for SQLite timestamps and add notified column to history ([#29](https://github.com/fjcasti1/hive/issues/29)) ([908c2c8](https://github.com/fjcasti1/hive/commit/908c2c8724ffa6cbd7d0c7139a58f6f303e3f191))


### 📚 Documentation

* Remove stray trailing `## Changelog` header from CHANGELOG ([596e0bb](https://github.com/fjcasti1/hive/commit/596e0bba24dac09cee03aac5bc26d2e14ad185e6))


### 💫 Code Refactoring

* `db.Delete` to use `RETURNING` and return the deleted row ([0a0d7e7](https://github.com/fjcasti1/hive/commit/0a0d7e7f077bcf4a48bc180ab0690e4da6af4c63))
* Introduce `db.Querier` interface so DB functions accept either ([0a0d7e7](https://github.com/fjcasti1/hive/commit/0a0d7e7f077bcf4a48bc180ab0690e4da6af4c63))
* Unexport `QueueEntry` → `queueEntry` since it's only used ([0a0d7e7](https://github.com/fjcasti1/hive/commit/0a0d7e7f077bcf4a48bc180ab0690e4da6af4c63))

## [1.2.0](https://github.com/fjcasti1/hive/compare/v1.1.0...v1.2.0) (2026-05-02)


### 🎁 New Features

* add macOS and tmux-bell notification channels ([#24](https://github.com/fjcasti1/hive/issues/24)) ([4247060](https://github.com/fjcasti1/hive/commit/424706062e10f3e12056e8648a1cd23f09afc641))
* auto-detect tmux session when --session is omitted ([#22](https://github.com/fjcasti1/hive/issues/22)) ([3322744](https://github.com/fjcasti1/hive/commit/332274406d619fc1a94154f6fa0ef5993a1153ec))


### ❔ Miscellaneous Chores

* Add sections to CHANGELOG ([#25](https://github.com/fjcasti1/hive/issues/25)) ([c5e9db0](https://github.com/fjcasti1/hive/commit/c5e9db0376b161819c0806b9c1154545f712f00d))

## [1.1.0](https://github.com/fjcasti1/hive/compare/v1.0.0...v1.1.0) (2026-05-02)


### Features

* add `notify` command for adding sessions to the queue and throw ([1587c74](https://github.com/fjcasti1/hive/commit/1587c74efb26b82bd6c3abe63bb54def917976e1))
* add core internal structural tooling ([#20](https://github.com/fjcasti1/hive/issues/20)) ([1587c74](https://github.com/fjcasti1/hive/commit/1587c74efb26b82bd6c3abe63bb54def917976e1))

## 1.0.0 (2026-04-24)


### Features

* initial hive CLI with version stub ([39f678f](https://github.com/fjcasti1/hive/commit/39f678fb5ac502adf69fd9215d669370cd8c0d32))
* initial hive CLI with version stub ([8232485](https://github.com/fjcasti1/hive/commit/8232485eb2c6aafcbe86411f6653dc37bc097eb8))


### Bug Fixes

* revert homebrew key to brews ([#11](https://github.com/fjcasti1/hive/issues/11)) ([b2edcad](https://github.com/fjcasti1/hive/commit/b2edcad14d40275ca439dd5ada6196feb57eaae4))
* switch to homebrew key and explicit token in goreleaser config ([#6](https://github.com/fjcasti1/hive/issues/6)) ([781ec0a](https://github.com/fjcasti1/hive/commit/781ec0a4c00aa00771bd5cc260a41e9ba7320dc6))
* use RELEASE_PLEASE_TOKEN so release PRs trigger CI ([#5](https://github.com/fjcasti1/hive/issues/5)) ([0537024](https://github.com/fjcasti1/hive/commit/053702490894c1136546cd5bdcf6da9a0624cc8d))
