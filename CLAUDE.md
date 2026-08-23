# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Design Documents — read these first

Two documents are the canonical record of this project. **Read the relevant one before changing behavior**, and do not restate their contents here.

| Document | Covers |
| --- | --- |
| [SPECS.md](./SPECS.md) | **Externally visible behavior** — CLI contract (commands, flags, output formats, exit codes), TSV/JSON formats, the `user.hasher.*` extended-attribute interface, accepted limitations. Items carry stable `SPEC-<AREA>-<nnn>` IDs |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | **Internal implementation** — package layout, core types, data flow, xattr caching, worker pool, diff algorithm, and the reasoning behind them |

Rules when working with them:

- Reference spec items by **ID** (e.g. `SPEC-XATTR-005`), not by paraphrase. IDs are never reused after deletion, so they stay greppable.
- **The same fact must not appear in both documents.** External behavior belongs to SPECS.md; how it is achieved belongs to ARCHITECTURE.md.
- If a change alters externally visible behavior, SPECS.md must be updated in the same change.
- `ARCHITECTURE.md` section 11 tracks known defects as **B1–B18**. Several are load-bearing (see "Known pitfalls" below). When fixing one, remove its entry and update anything in SPECS.md marked "⚠️ 既知の不具合".

Both documents are written in Japanese and describe the implementation **as it actually behaves**, not as it was intended to behave.

## Development Commands

### Building
```bash
# Local development build
cd src
goreleaser build -f ../.goreleaser.yml --clean --snapshot

# Plain build (fastest for iteration)
cd src
go build -o /tmp/hasher .

# Production build (releases)
# Uses .goreleaser.yml configuration for cross-platform builds
```

### Code Quality
```bash
# Run pre-commit hooks (includes go fmt, go mod tidy, golangci-lint)
pre-commit run -a

# Individual linting
cd src
go fmt ./...
go mod tidy
golangci-lint run
```

### Testing
```bash
cd src
go test ./...

# Run specific test
go test ./core -run TestSpecificFunction

# Run tests with verbose output
go test -v ./...
```

Tests exist **only in `src/core/`**; `cmd/` and `common/` have none. Coverage is concentrated on the diff algorithm (`dirdiff_test.go` covers all `DiffStatus` values). See ARCHITECTURE.md section 12 for what is untested — notably the recursive tree diff, TSV parsing, the worker pool, and the entire CLI layer.

Test helpers live in `helper_test.go`. Note that `copyFile` deliberately restores the source's mtime, because diff comparison depends on it.

### Dependencies
- Uses Go modules (`src/go.mod`)
- Dependency updates managed by Renovate
- Tool management via aqua (`aqua.yaml`)

## Orientation

Three packages, one-directional dependencies:

```
main ─> cmd ─> core ─> common
         └──────────────┘
```

| Package | Responsibility |
| --- | --- |
| `src/cmd/` | Cobra command definitions, flag handling, screen output |
| `src/core/` | Hash logic: xattr caching, worker pool, diff algorithms |
| `src/common/` | File type detection, color constants, directory walking |

The central idea is that **hash values live in the file's own extended attributes** rather than a side database, so they survive moves and renames. Everything else follows from that. See ARCHITECTURE.md section 1.

## Known pitfalls when changing this code

These are consequences of B1–B18 that will mislead you if you assume the code works as it reads:

- **Hashing is effectively single-threaded.** The worker pool and its CPU-based sizing logic are implemented, but the caller hardcodes 1 worker (`cmd/update.go:97`). Do not assume concurrency is exercised (B14).
- **Symbolic links are not actually skipped.** File-type detection uses `os.Stat`, which follows links, so the "symlink" branch is unreachable; `os.Lstat` appears nowhere in the repo (B2, B3).
- **`dirdiff` never compares files directly under the given roots** — only files inside subdirectories (B11).
- **`update` without `-r` swallows failures**, returning exit 0 with no message, due to a shadowed `err` (B1). The two `// nolint:govet` markers there are suppressing the warning that would have caught it.
- **Directory walking is split across three implementations** with differing symlink and error handling (ARCHITECTURE.md section 10). Check which one a code path uses before changing walk behavior.
- **SHA-1 is the only usable algorithm.** The algorithm abstraction supports sha256/sha512 by name, but no CLI flag selects them, and only `crypto/sha1` is registered via blank import in `main.go`.
- **`Json()` is unreachable from the CLI** and does no escaping (B18). Do not treat it as a supported output format.
- **TSV loading is not backward compatible** — column 4 must contain `:`; a bare hash is rejected (`core/hashstore.go:116-119`).
- **There is no TTY detection.** ANSI escapes are emitted even when piped or redirected, and `NO_COLOR` is unsupported (B16, `SPEC-LIMIT-003`).

When fixing any of these, add a regression test — most of them survived because the area has no test coverage.

## Build Configuration

- GoReleaser for cross-platform releases: linux/darwin/windows × amd64/arm64
- CGO disabled for static binaries
- Version info injected via ldflags into `cmd` package variables (`version`, `revision`, `date`, `osArch`)
- Archive names use a capitalized OS and `x86_64` for amd64, e.g. `hasher_Linux_x86_64.tar.gz`
- `go.mod` declares Go 1.21 while CI and releases build with 1.26.6 (B12)
- CI only triggers on changes under `src/`, `.github/workflows/`, and `aqua.yaml` — so `.goreleaser.yml` changes are not validated by PRs

## Code Style

- Uses pre-commit hooks for consistency
- Go standard formatting via `go fmt`
- golangci-lint for static analysis (`govet` with all checks, `errcheck`, `staticcheck`, `misspell`)
- Conventional commit messages preferred
- Prefer adding a `// nolint` only with a comment explaining why; existing ones have hidden real bugs
