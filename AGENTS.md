# Repository Guidelines

## Project Structure & Module Organization
- `cmd/wt/` is the main entrypoint for the CLI binary.
- `internal/cli/` wires Cobra commands for `wt add`, `wt list`, and other subcommands.
- `internal/git/`, `internal/config/`, `internal/setup/`, `internal/tui/`, and `internal/util/` hold worktree operations, config parsing, setup copy/link logic, TUI models, and helpers.
- `scripts/` contains install scripts referenced in the README.
- `.github/workflows/` and `.goreleaser.yml` drive tagged releases.

## Build, Test, and Development Commands
- `make build` builds `bin/wt` with embedded version info.
- `make run` builds then runs the local binary.
- `make test` runs `go test -v ./...` across all packages.
- `make tidy` syncs `go.mod` and `go.sum`.
- `make install` copies `bin/wt` to `~/.local/bin/`.

## Coding Style & Naming Conventions
- Use standard Go formatting (`gofmt`) and idiomatic Go patterns.
- Keep CLI command files in `internal/cli/` with descriptive names (for example, `add.go`, `list.go`).
- File names are lower_snake_case; tests use `_test.go`.
- Avoid exporting identifiers unless they are used across packages.

## Testing Guidelines
- Tests are standard Go tests placed adjacent to the code under `internal/...`.
- There are currently no committed test files; add tests for new logic or bug fixes.
- Run the full suite with `make test` before opening a PR.

## Commit & Pull Request Guidelines
- Follow the existing conventional commit style: `feat:`, `fix:`, `chore:`, `ci:` with short, imperative summaries.
- PRs should include a concise description, relevant command output (for example, `make test`), and the platform(s) validated on.

## Release & Configuration Notes
- Tagging `v*` releases triggers GoReleaser via GitHub Actions; update `.goreleaser.yml` if packaging changes.
- `.wt.json` is the repo’s sample config; update it cautiously if behavior or defaults change.
