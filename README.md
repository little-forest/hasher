# hasher

# Install

```bash
HASHER_VERSION=v0.1.2
INSTALL_DIR=~/bin

curl -sLo - https://github.com/little-forest/hasher/releases/download/${HASHER_VERSION}/hasher_linux_x86_64.tar.gz \
  | tar -C ${INSTALL_DIR} -zxv
```

# Use as a library

Hash calculation and the extended-attribute cache are published as a Go
package. It depends only on `github.com/pkg/xattr` — no color or progress
output is pulled into your application.

```
go get github.com/little-forest/hasher
```

```go
package main

import (
	_ "crypto/sha1" // register the algorithm; without it CalcHash fails
	"fmt"

	"github.com/little-forest/hasher/hashcore"
)

func main() {
	alg := hashcore.NewDefaultHashAlg() // sha1

	// Calculate without touching the file's attributes.
	hash, err := hashcore.CalcHash("a.txt", alg)
	if err != nil {
		panic(err)
	}
	fmt.Println(hash)

	// Or read through the xattr cache, recalculating only when the file
	// changed, and store the result back.
	changed, hash, err := hashcore.UpdateHashStrictly("a.txt", alg, false)
	if err != nil {
		panic(err)
	}
	fmt.Println(changed, hash)
}
```

The attribute names (`user.hasher.*`) are identical to the ones the CLI uses,
so both sides read each other's cache. Packages under `internal/` are not part
of the public interface. See `SPEC-LIB-001` in [SPECS.md](./SPECS.md).


# For developers

## Local build

Development tools (Go, GoReleaser, Task) are pinned in `aqua.yaml`.
Install [aqua](https://aquaproj.github.io/), run `aqua i -l`, then:

```
task build          # current platform only, symlinks ./hasher
task build-all      # all platforms
task test
task lint
```

## pre-commit

This repository allows code checking before committing locally by using [`pre\-commit`](https://pre-commit.com/).


1. [Install pre\-commit](https://pre-commit.com/#install)
2. Install hook script by pre-commit
```
pre-commit install
```
3. Install [golangci\-lint](https://github.com/golangci/golangci-lint)

`pre-commit` checks only staged files. If you want to check all files, please do the following.

```
pre-commit run -a
```

## cobra-cli

Install cobra-cli.

```
go install github.com/spf13/cobra-cli@latest
```

Make sub-command template.

```
cobra-cli add [SUB_COMMNAD_NAME]
```
