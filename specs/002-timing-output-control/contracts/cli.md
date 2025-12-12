# CLI Contract: Timing Output Control

## Flag

- Name: `--timing`
- Type: boolean
- Scope: global/root persistent flag (available on all commands and subcommands)
- Default: `false`

## Output Contract

### Default behavior (`--timing` not set)

- No timing output is printed.
- Existing output for all commands remains unchanged.

### Timing enabled (`--timing` set)

For each existing success/completion message printed with the `✔` symbol, print a timing block immediately after it:

```text
✔ <completion message>
⏲ current: <duration>
  total:  <duration>
```

Where:

- `<duration>` is formatted using Go `time.Duration` string formatting.
- `current` represents the elapsed time of the most recently completed activity.
- `total` represents the accumulated elapsed time across completed activities within the same command run.

### Error behavior

- Timing output must not be printed as part of error messages.
- If timing is enabled and a command fails, timing output should not be emitted for the failing activity.
