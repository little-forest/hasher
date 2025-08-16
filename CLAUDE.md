# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building
```bash
# Local development build
cd src
goreleaser build -f ../.goreleaser.yml --clean --snapshot

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

### Dependencies
- Uses Go modules (`src/go.mod`)
- Dependency updates managed by Renovate
- Tool management via aqua (`aqua.yaml`)

## Architecture Overview

### Core Components

**Hash Calculation Engine (`src/core/`)**
- `hasher.go`: Main hash calculation logic with extended attributes (xattr) caching
- `hash.go`: Hash data structure with TSV/JSON output formats
- `hashalg.go`: Hash algorithm abstraction layer
- `hashstore.go`: In-memory hash storage with TSV import/export

**Command Layer (`src/cmd/`)**
- `root.go`: Cobra CLI root command setup
- Individual command files (`calc.go`, `update.go`, `compare.go`, etc.)
- Progress notification system (`*_progress_notifier.go`)

**Common Utilities (`src/common/`)**
- File type detection and validation
- Terminal color output utilities
- Path resolution and file walking

### Key Architectural Patterns

**Extended Attributes (xattr) Caching**
- Hash values stored as file extended attributes under `user.hasher.*`
- Attributes: hash value, file size, modification time, last check time
- Enables incremental hash updates by detecting file changes

**Worker Pool Pattern**
- Concurrent hash calculation using configurable worker pools
- Worker count auto-adjusted based on CPU cores (max: CPU-1)
- Progress tracking across workers with real-time notifications

**Hash Data Format**
- TSV format: `{path}\t{basename}\t{mtime}\t{algorithm}:{hash}`
- Supports multiple hash algorithms through abstraction layer
- Backwards compatible hash file loading

### File Organization

```
src/
├── main.go              # Entry point
├── cmd/                 # CLI command implementations
├── core/                # Core hashing logic
├── common/              # Shared utilities
└── go.mod              # Go module definition
```

### Data Flow

1. **File Discovery**: Recursive directory walking with symlink skipping
2. **Hash Validation**: Check xattr cache against file size/mtime
3. **Hash Calculation**: Parallel processing with progress tracking
4. **Storage**: Update xattr cache and optional TSV export
5. **Output**: Multiple formats (TSV, JSON, terminal display)

### Testing Strategy

- Unit tests in `*_test.go` files alongside source
- Helper functions in `helper_test.go`
- Focus on core hash calculation and file diff logic
- Mock progress notifiers for testing concurrent operations

### Build Configuration

- GoReleaser for cross-platform releases (Linux, macOS, Windows)
- CGO disabled for static binaries
- Version injection via ldflags
- Archive naming follows standard conventions

### Code Style

- Uses pre-commit hooks for consistency
- Go standard formatting via `go fmt`
- golangci-lint for static analysis
- Conventional commit messages preferred