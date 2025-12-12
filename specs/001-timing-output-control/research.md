# Research: Timing Output Control

## Context (existing behavior)

- KSail-Go already has `pkg/ui/timer` providing `(total, stage)` durations and `pkg/ui/notify` rendering timing as `"[stage: X]"` or `"[stage: X|total: Y]"` appended to success lines.
- Several commands (and helpers like `pkg/cmd/lifecycle_helpers.go`) build timing strings manually using `notify.FormatTiming(...)`.

## Decisions

### Decision 1 — Timing output is opt-in via `--timing`

- **Decision**: Add a root-level (persistent) boolean flag `--timing` that enables timing output for the current invocation.
- **Rationale**: Matches the spec’s “flag-only” scope and avoids introducing config persistence.
- **Alternatives considered**:
  - Always-on timing (current behavior in some commands) → rejected because it adds noise by default.
  - `ksail.yaml` config → rejected by spec clarification.

### Decision 2 — Reuse existing `pkg/ui/timer` semantics

- **Decision**: Treat `Timer.GetTiming()`’s `stage` as **current** and `total` as **total**.
- **Rationale**: This matches the spec requirement (`current` = most recent activity duration; `total` accumulates). It minimizes new logic.
- **Alternatives considered**:
  - Introduce a separate accumulator type → rejected as unnecessary given existing timer behavior.

### Decision 3 — Centralize formatting in `pkg/ui/notify`

- **Decision**: Implement the spec-required 3-line timing block formatting in `pkg/ui/notify` and remove ad-hoc `FormatTiming` string concatenation from call sites.
- **Rationale**: Reduces duplication and ensures consistent output across commands.
- **Alternatives considered**:
  - Keep formatting at call sites → rejected due to drift risk.

### Decision 4 — Backward compatibility strategy

- **Decision**: Default output contains **no timing**; when `--timing` is enabled, print the spec-defined timing block after each completion message.
- **Rationale**: Spec requires default-off and a new format.
- **Alternatives considered**:
  - Preserve legacy `"[stage: ...]"` output under `--timing` → rejected because the spec mandates the new block format.

## Implications / Follow-ups

- Snapshot tests that currently include `"[stage: ...]"` output must be updated to reflect default-off timing.
- Introduce/extend a single “completion” helper for printing the completion line plus optional timing block.
