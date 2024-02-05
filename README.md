# hasher

# Install

```bash
HASHER_VERSION=v0.1.2
INSTALL_DIR=~/bin

curl -sLo - https://github.com/little-forest/hasher/releases/download/${HASHER_VERSION}/hasher_linux_x86_64.tar.gz \
  | tar -C ${INSTALL_DIR} -zxv
```

# For developers

## Local build

```
cd src
goreleaser build -f ../.goreleaser.yml --clean --snapshot
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
cd src
cobra-cli add [SUB_COMMNAD_NAME]
```
