# Phase 0 Research – Consolidate Cluster Commands

## Decision: Introduce `cluster` parent command that owns all lifecycle subcommands

- **Rationale**: Current `cmd/root.go` registers lifecycle actions (`NewUpCmd`, `NewDownCmd`, etc.) directly on the root command. Grouping them under a dedicated parent keeps the root surface area lean while matching the feature spec’s UX goal. Cobra already supports nested commands with shared help; creating a `NewClusterCmd` wrapper lets us reuse the existing handlers (`HandleUpRunE`, etc.) without changing their execution signatures.
- **Alternatives Considered**:
  - *Alias-only approach*: Keep legacy commands and add aliases pointing to new subcommands. Rejected per clarification #1 requiring the legacy commands to be removed entirely.
  - *Flag-based grouping*: Use a `--cluster` flag on existing commands. Rejected because it does not reduce command clutter and complicates help output.

## Decision: Move lifecycle command constructors into a dedicated `cmd/cluster` module

- **Rationale**: Each lifecycle command lives in its own file at repository root (`cmd/up.go`, `down.go`, etc.). Relocating or renaming them under `cmd/cluster` (for example, `cmd/cluster/up.go`) keeps the package organized around the new hierarchy and prevents accidental re-registration on the root command. Unit tests in `cmd/*.go` already isolate constructors, easing the move.
- **Alternatives Considered**:
  - *Keep constructors in root package and register via helper slice*: Minimally invasive but risks future regressions because nothing stops developers from re-adding them directly to root.
  - *Inline anonymous subcommands inside `NewClusterCmd`*: Reduces file count but loses existing handler reuse and test structure.

## Decision: Update command tests to reflect nested invocation while maintaining handler coverage

- **Rationale**: Tests such as `cmd/up_test.go` leverage `testutils.TestSimpleCommandCreation` to validate command metadata. After relocation, these tests must assert the new parent command path (e.g., `cluster up`) and ensure root command help emphasises the new grouping. We will keep handler tests intact to satisfy the constitution’s TDD requirement.
- **Alternatives Considered**:
  - *Delete existing tests and rely on manual QA*: Violates constitutional principle II (TDD-first).
  - *Only test the parent `cluster` command*: Would miss regressions in individual subcommand wiring and flag bindings.

## Decision: Refresh help text and documentation hooks without runtime banners

- **Rationale**: Clarification #3 mandates documentation-only notification. Updating `root --help` and `cluster --help` descriptions plus README snippets satisfies the requirement without adding runtime warnings. Cobra automatically surfaces descriptions in help output once we adjust `Short` and `Long` strings.
- **Alternatives Considered**:
  - *One-time warning banner*: Rejected per clarification #3.
  - *Dedicated migration command*: Overkill for a CLI surface reshuffle.
