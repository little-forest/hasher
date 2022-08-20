# hasher

# For developers

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

