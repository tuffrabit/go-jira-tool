# Contributing

Thanks for your interest in contributing to `go-jira-tool`.

## Development setup

Requires Go 1.25+.

```bash
git clone https://github.com/tuffrabit/go-jira-tool.git
cd go-jira-tool
go build
```

You will need a Jira Cloud API token to test against a real instance — see the README for setup. Do not commit your `config.json`.

## Build and verify

```bash
go vet ./...
go build ./...
go test ./...
```

The `build.sh` / `build.bat` scripts produce cross-platform release binaries under `bin/`.

## Submitting changes

1. Open an issue first for non-trivial changes so we can align on direction.
2. Fork, branch, and open a PR against `main`.
3. CI must pass.
4. Keep changes focused — separate refactors from feature work.

## Reporting issues

Please include:

- The command you ran (with credentials redacted).
- The output (stdout and stderr).
- Your Go version (`go version`) and OS.
- Whether your Jira is Cloud or Server/Data Center (this tool is currently Cloud-only).

## Licensing

By submitting a contribution, you agree that it is licensed under the same Apache License 2.0 as the rest of the project.
