# Quickstart: Timing Output Control

## Goal

Enable per-activity timing output on demand for a single CLI run.

## Usage

- Default (timing off):

  ```bash
  ksail <command> [args]
  ```

- Enable timing output:

  ```bash
  ksail --timing <command> [args]
  # or
  ksail <command> --timing [args]
  ```

## What you’ll see

When `--timing` is enabled and an activity completes successfully, the CLI prints the usual completion line followed by a timing block:

```text
✔ <completion message>
⏲ current: <duration>
  total:  <duration>
```

Durations use Go `time.Duration` string formatting (examples: `12ms`, `1.2s`, `3m4.5s`).

## Notes

- Timing output is per-invocation: the next run without `--timing` prints no timing.
- Timing output is intended to stay out of default output to avoid noise and snapshot churn.
