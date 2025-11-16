# Validation Log — Documentation Migration

| Date | Check | Command | Result | Notes |
|------|-------|---------|--------|-------|
| 2025-11-16 | Markdownlint | `npx markdownlint-cli2 "docs/**/*.md"` | Failed | Legacy KSail docs trigger MD013/MD025/MD033 violations; remediation required post-migration |
| 2025-11-16 | Lychee | `lychee --config .lycheeignore docs` | Failed | CLI expects TOML config; adjust to use `.lycheeignore` automatically or `--exclude-path` before re-running |
| 2025-11-16 | Lychee (core concepts) | `lychee docs/overview/core-concepts` | Passed | `.lycheeignore` is honored automatically; zero errors at 30 OK links |
| 2025-11-16 | Lychee (full docs) | `lychee docs` | Failed | Missing `docs/images/architecture.drawio.png` asset and `https://kubernetes-sigs.github.io/kustomize/` returns 404 — address in asset migration (T008) and link update |
| 2025-11-16 | Markdownlint (core concepts) | `npx markdownlint-cli2 "docs/overview/core-concepts/*.md"` | Passed | New core concept docs lint clean after replacing tables with sectioned content |
| 2025-11-16 | Markdownlint (full docs) | `npx markdownlint-cli2 "docs/**/*.md"` | Passed | Post-migration tables reformatted; entire doc set lint clean |
| 2025-11-16 | Lychee (full docs) | `lychee docs` | Passed | Assets copied and Kustomize link updated (T008, T009); 0 errors, 2 redirects |
| 2025-11-16 | Usability walkthrough | `docs/overview/index.md` → `docs/use-cases/local-development.md` → `docs/configuration/index.md` | Passed | Manual review following local development scenario using KSail-Go commands; no gaps detected |
