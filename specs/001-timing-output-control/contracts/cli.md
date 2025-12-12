# Contract: CLI Timing Output

## Flag

- Name: `--timing`
- Type: boolean
- Scope: root-level persistent flag (applies to all subcommands)
- Default: off

## Output Contract

When `--timing` is enabled, after each timed activity completion message (`✔ ...`), the CLI prints a timing block:

```text
✔ completion message
⏲ current: <duration>
  total:  <duration>
```

### Definitions

- `current`: duration of the most recently completed timed activity.
- `total`: accumulated duration across timed activities in the current invocation.

### Notes

- This contract intentionally avoids config-file support.
- Duration formatting should be consistent across commands (use Go duration formatting unless otherwise specified).
