# Documentation Validation Workflow

## Tooling Summary

| Check | Command | Purpose | Notes |
|-------|---------|---------|-------|
| Markdown style | `npx markdownlint-cli2 "docs/**/*.md"` | Enforces formatting and heading consistency across migrated docs. | Uses repo `.markdownlint.json`; requires Node.js/npm. |
| Link integrity | `lychee --config .lycheeignore docs` | Validates internal and external links referenced in documentation. | Requires Lychee CLI installed via Homebrew (`brew install lychee`) or Cargo. |
| Manual preview | GitHub Markdown preview or VS Code built-in preview | Confirms layout, tables, and images render as expected. | Spot-check representative pages after automated checks. |

## Execution Cadence

1. **During Migration (per section)**
   - Run `npx markdownlint-cli2 "docs/**/*.md"` after migrating a batch of pages.
   - Address lint errors immediately to avoid compounding fixes.
2. **After Link Updates**
   - Run `lychee --config .lycheeignore docs` to confirm local and cross-repository links resolve.
   - For transient network failures, re-run the command to confirm outcome.
3. **Pre-PR Validation**
   - Execute both commands sequentially and record results (timestamp and outcome) in `specs/001-migrate-docs/validation-log.md`.
   - Use VS Code or GitHub web preview to review key pages: `docs/overview/index.md`, `docs/configuration/index.md`, and `docs/use-cases/local-development.md`.
4. **Usability Verification**
   - After automated checks pass, conduct the operator walkthrough (Task T018) and log findings in the validation log.

## Failure Handling

- Capture failing command output in the validation log for traceability.
- For broken links, update URLs or add allowed patterns to `.lycheeignore` with justification.
- For lint failures caused by intentional formatting, annotate the Markdown with HTML comments (`<!-- markdownlint-disable -->`) sparingly and document the rationale.
