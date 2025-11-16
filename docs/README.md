# KSail-Go Documentation

This repository hosts Markdown-only documentation rendered directly in GitHub. The `docs/` tree mirrors the structure of the original KSail project and is organized by topic so you can browse or link to specific guidance.

## Directory layout

- `overview/` – Product overview, architecture, project structure, and support matrix.
- `configuration/` – CLI flag reference, declarative configuration guides, and precedence rules.
- `use-cases/` – Scenario playbooks covering learning, local development, and CI/CD pipelines.
- `images/` – Shared diagrams used across the guides.

Each folder contains additional `README`-style index pages to help you navigate deeper.

## Local preview workflow

The documentation is validated with markdownlint and Lychee before every pull request. Run the same commands locally to catch issues early:

```bash
# From the repository root
npx markdownlint-cli2 "docs/**/*.md"
lychee docs
```

Markdownlint enforces formatting rules (headings, tables, code fences), while Lychee scans for broken links and missing assets. The ignore list for Lychee lives in `.lycheeignore`.

For a rendered preview, open the Markdown files directly in VS Code or use `Markdown: Open Preview` to check anchor links and admonitions. GitHub renders the same Markdown dialect, so local preview is typically sufficient.

## Writing guidelines

- Prefer relative links (for example `../configuration/index.md`) so the docs work both on GitHub and in local editors.
- Keep command snippets runnable; avoid shell prompts (`$`) and use fenced blocks with language hints.
- Update cross-links when moving or renaming files so navigation remains intact.
- If you add new assets, store them under `docs/images/` and update Lychee ignores only when absolutely necessary.

## Need help?

Open an issue or start a discussion in the repository if you spot gaps or want to propose additional guides. Contributions are welcome—see the root `README.md` and `CONTRIBUTING.md` for details.
