# Research: Timing Output Control

## Decision 1: Control via global `--timing` flag only

- Decision: Add a root persistent boolean flag `--timing` (default false) to enable timing output for a single CLI invocation.
- Rationale: Matches the spec’s opt-in requirement and keeps default output stable.
- Alternatives considered:
  - Config file / `ksail.yaml` option: rejected by requirement (CLI-only).
  - Per-command flags: rejected (must be global/persistent).

## Decision 2: Reuse existing `timer.Timer` (`GetTiming`) as the source of truth

- Decision: Use the existing `pkg/ui/timer.Timer` interface and its `GetTiming()` return values.
  - `current` := stage duration
  - `total` := total duration
- Rationale: Avoids redefining what is timed and keeps changes scoped to output/rendering.
- Alternatives considered:
  - New timing accumulator types: unnecessary (YAGNI) given existing abstraction.

## Decision 3: Render timing as a multi-line block after `✔` messages

- Decision: Change timing rendering to a block printed immediately after an existing success/completion line.

  Output contract when timing is enabled:

  ```text
  ✔ <completion message>
  ⏲ current: <duration>
    total:  <duration>
  ```

  Where `<duration>` is Go `time.Duration` string format.

- Rationale: Aligns with the clarified spec and preserves the existing `✔` line unchanged.
- Alternatives considered:
  - Inline timing suffix (current behavior): rejected (format mismatch).
  - JSON / machine-readable timing: out of scope.

## Decision 4: Test strategy

- Decision: Cover behavior via public APIs only:
  - `pkg/ui/notify` tests assert formatting via exported functions/types.
  - CLI/command tests assert flag-driven output differences and keep default output stable.
- Rationale: Complies with constitution (black-box tests only) and protects user-visible output.
