Before finishing a change:

1. Regenerate mocks if interfaces changed (`mockery`).
2. Build binary (`go build -o ksail .`) and/or `go build ./...`.
3. Run all tests (`go test ./...`).
4. Run `golangci-lint run --timeout 5m`; optionally run `mega-linter-runner -f go` for full validation.
5. Execute key CLI command helps (`./ksail --help`, etc.) when relevant modifications impact UX.
6. Ensure notifications/timing output read cleanly in command output when changes affect CLI messaging.
