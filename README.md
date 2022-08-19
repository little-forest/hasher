# hasher

## lint

```
cd src
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v --disable-all --enable=govet,errcheck,staticcheck
```


