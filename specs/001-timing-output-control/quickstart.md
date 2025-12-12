# Quickstart: Timing Output Control

## Goal

Enable timing output for a single KSail invocation using a CLI flag.

## Prereqs

- Go toolchain that matches `go.mod` (`go 1.25.4`).

## Run

### 1) Default behavior (timing off)

- Run any command without `--timing`.
- Expected: no timing block is printed.

Example:

- `ksail cluster init`

### 2) Enable timing output (timing on)

- Run the same command with `--timing`.
- Expected: after each `✔ ...` completion message, print:

```text
✔ completion message
⏲ current: <duration>
  total:  <duration>
```

Example:

- `ksail --timing cluster init`

## Validate locally

- `go test ./...`
- `go build ./...`
- `golangci-lint run --timeout 5m`
