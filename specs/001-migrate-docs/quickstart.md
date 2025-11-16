# Quickstart — Documentation Migration to KSail-Go

## Prerequisites

- Access to both repositories within the monorepo (`projects/ksail` and `projects/ksail-go`).
- Node.js or another runtime capable of running `npx markdownlint-cli2`.
- Lychee CLI (`cargo install lychee` or Homebrew `brew install lychee`).
- Familiarity with KSail-Go CLI structure (`ksail cluster`, `ksail workload`).

## Workflow

1. **Create workspace snapshot**
   - Checkout branch `001-migrate-docs` in `projects/ksail-go`.
   - Optionally open `projects/ksail/docs/` in a split view for reference.
2. **Inventory current docs**
   - Record relevant files from `projects/ksail/docs/` (overview, configuration, use-cases, images).
   - Capture legacy commands that require translation to the Go CLI.
3. **Copy and normalize content**
   - Copy Markdown files into the corresponding `ksail-go/docs/` directory.
   - Remove Jekyll front matter (`---` blocks) and adjust headings to standard Markdown.
   - Update table-of-contents or navigation references to relative links that work in GitHub Markdown.
4. **Rewrite for KSail-Go semantics**
   - Replace legacy commands (`ksail up`, `ksail down`, etc.) with their Go equivalents (`ksail cluster create`, `ksail cluster delete`, etc.).
   - Update configuration examples to match Go-based file names/paths where they differ.
   - Refresh screenshots or diagrams if CLI output has changed; otherwise ensure asset references resolve to `docs/images/`.
5. **Validate documentation**
   - Run `npx markdownlint-cli2 "docs/**/*.md"` from the repository root.
   - Run `lychee --config .lycheeignore docs` to validate links.
   - Spot-check key pages in a Markdown preview (VS Code or GitHub web preview).
6. **Document findings**
   - Update `specs/001-migrate-docs/spec.md` migration summary with sections migrated/deferred.
   - Note follow-up tasks for publishing or remaining docs in `specs/001-migrate-docs/tasks.md` (Phase 2).
7. **Open PR checklist**
   - Ensure lint checks pass and `git status` shows updated docs only.
   - Provide before/after notes for major command or structural changes in the PR description.

## Validation Commands

```bash
npx markdownlint-cli2 "docs/**/*.md"
lychee --config .lycheeignore docs
```
