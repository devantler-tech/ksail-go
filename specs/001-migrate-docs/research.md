# Research Findings — 001-migrate-docs

## Decision: Focus migration on core KSail documentation sets (overview, configuration, use-cases, supporting assets)

- **Rationale**: The source repository (`projects/ksail/docs/`) contains well-structured content under `overview/`, `configuration/`, `use-cases/`, and shared `images/`. These map directly to the existing but empty directories in `ksail-go/docs/`, creating a clear one-to-one migration path that preserves navigation intent. Keeping the scope to these folders satisfies FR-001/FR-002 while avoiding obsolete material (e.g., legacy FAQ/roadmap that references .NET internals).
- **Alternatives considered**:
  - *Migrate every markdown file wholesale*: Rejected because legacy FAQ/roadmap contain .NET-specific details that would need heavy rewriting and could blur the KSail-Go focus.
  - *Author entirely new docs*: Discarded as it would ignore valuable vetted material and extend delivery time.

## Decision: Update command references to KSail-Go hierarchy (`ksail cluster ...`, `ksail workload ...`)

- **Rationale**: The Go CLI splits responsibilities into grouped subcommands (see `src/cmd/cluster/*.go`, `src/cmd/workload/*.go`), replacing former top-level verbs like `ksail up` and `ksail down`. Documented examples must showcase the new usage (e.g., `ksail cluster init`, `ksail cluster create`, `ksail workload reconcile`) to deliver accurate guidance per FR-003 and SC-001.
- **Alternatives considered**:
  - *Keep legacy single-level commands and note future changes*: Rejected because it perpetuates confusion and contradicts the goal of reflecting the Go CLI.
  - *Provide dual command references (old + new)*: Dismissed to keep the docs simple (KISS) and avoid maintaining obsolete syntax.

## Decision: Strip Jekyll-only metadata, keep Markdown-compatible structure

- **Rationale**: Source docs include YAML front matter (`--- ... ---`) for the Just-the-Docs navigation system. Without adopting Jekyll (Option B), keeping the metadata offers no benefit and can confuse contributors. Removing or converting navigation cues to standard Markdown headings keeps files renderer-agnostic and aligns with the constraint of remaining Markdown-native.
- **Alternatives considered**:
  - *Retain front matter untouched*: Rejected because it adds noise during reviews and signals unsupported tooling.
  - *Port the Just-the-Docs stack*: Declined per clarified scope—would reintroduce tooling overhead we explicitly avoided.

## Decision: Validate docs with Markdownlint + Lychee plus spot GitHub preview

- **Rationale**: The repository already configures `.markdownlint.json` and `.lycheeignore`. Running `npx markdownlint-cli2 "docs/**/*.md"` (or the MegaLinter workflow) and `lychee --config .lycheeignore docs` before publishing ensures formatting and links remain healthy, fulfilling FR-006 and SC-002/SC-003 without introducing new dependencies.
- **Alternatives considered**:
  - *Rely solely on GitHub preview*: Rejected as it lacks automated failure signals and violates Test-First expectations.
  - *Introduce a new docs-specific CI pipeline*: Deemed unnecessary for this migration; existing lint/link tooling is sufficient.
