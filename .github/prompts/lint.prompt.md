
---
description: Run linting tools, fix reported problems, and confirm a clean build.
---

Always read any user-supplied arguments before acting:

```
$ARGUMENTS
```

## Quick prep

- Work from the repository root (directory containing `go.mod`).
- Skim the relevant configuration files (`.golangci.yml`, `.jscpd.json`, `.cspell.json`, `CONTRIBUTING.md`) only when you need specific settings or exclusions—don’t summarize them.

## Linting workflow

1. **golangci-lint**
   - Run `golangci-lint run --timeout=5m`.
   - Apply automatic fixes with `golangci-lint run --fix` (or targeted `golangci-lint fmt`) when available.
   - Tackle remaining issues directly in the source—shorten functions, reduce complexity, add helpers, rename variables, etc.—until the tool passes.

2. **jscpd**
   - Run `jscpd` to detect duplicated code.
   - Consolidate repeated blocks into shared helpers or utilities so the report is clean.

3. **cspell**
   - Run `cspell` using the project configuration.
   - Fix typos or, if a term is correct, add it to the dictionary file mentioned in `.cspell.json`.

## Fixing expectations

- Bias toward hands-on fixes instead of documentation.
- Group similar lint issues and resolve them together.
- Keep Go idioms, readability, and testability intact; adjust tests when lint requires it (e.g., add `t.Helper()`).
- Avoid touching `testutils` or generated code unless the lint points there explicitly.

## Validation

After each major batch of fixes:

- Re-run the relevant linter to confirm the issue is gone.
- Run `go test ./...` to guard against regressions.

You’re done when all linting tools exit cleanly and the tests pass. Provide a short note describing the concrete fixes you made (no progress checklists or extended reports).

