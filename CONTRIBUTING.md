# Contributing

Thanks for considering a contribution to devx.

## Setup

```bash
git clone https://github.com/gunwooko/devx.git
cd devx
go mod tidy
go test ./...
```

## Pull requests

- Keep changes focused.
- Add or update tests for behavior changes.
- Run `gofmt`, `go test ./...`, and `go vet ./...`.
- Explain user-visible behavior in the pull request description.

## Issues

Include:

- macOS or Linux version
- Go version
- tmux version
- `devx doctor` output
- Steps to reproduce
