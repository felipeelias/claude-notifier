# Contributing

Thanks for your interest in contributing.

## Reporting bugs

Open an issue. Include what you expected, what happened instead, and your OS/Go version. Config snippets help (redact tokens).

## Adding a plugin

This is the most useful kind of contribution. The process:

1. Fork the repo and create a branch
2. Add your plugin in `plugins/<name>/<name>.go`
3. Implement the `notifier.Notifier` interface (`Name()` and `Send()`)
4. Add `toml` struct tags for config fields
5. Implement `SampleConfig()` so `claude-notifier init` includes your plugin
6. Register with `cli.Registry.Register()` in an `init()` function
7. Add the blank import in `main.go`
8. Write tests (use `httptest` for HTTP-based plugins, see `plugins/ntfy/` for reference)
9. Open a PR

If you're unsure whether a plugin fits, open an issue first to discuss it.

## Development

```bash
# install go 1.24+ via asdf
asdf install

# run unit tests only (faster)
go test ./... -short -v -race

# run all tests including integration
go test ./... -v -race

# build
go build -o claude-notifier .

# test manually
echo '{"message":"hello","title":"test"}' | ./claude-notifier --config path/to/config.toml
```

## Style

- `go vet` and `go test -race` must pass
- `markdownlint-cli2` must pass on all markdown files
- Conventional commits (`feat:`, `fix:`, `test:`, `docs:`, etc.)
- Keep plugins self-contained in their own package
- Don't add dependencies unless you really need them

## Pull requests

Keep PRs focused on one thing. A plugin PR shouldn't also refactor the config loader. Tests are required for new code. The CI runs `go test -race` and `go vet` on every PR.
