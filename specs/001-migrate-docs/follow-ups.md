# Follow-up Tasks — KSail Documentation Migration

The following items were intentionally deferred from the 001-migrate-docs feature. Track them in future iterations to close documentation gaps and improve the publishing experience.

## Documentation Content

- Regenerate CLI screenshots (`docs/images/ksail-cli-dark.png`, `docs/images/ksail-cli-light.png`) once the Go-based UI output settles.
- Evaluate whether any remaining KSail FAQ or roadmap content should be rewritten for KSail-Go or replaced with new guidance.
- Expand the use-case catalog with additional scenarios (e.g., GitOps bootstrap, Talos integration) after the core workflows gain validation.

## Tooling & Publishing

- Decide on a publishing strategy (GitHub Pages vs. raw Markdown) and capture automation steps if a static site is adopted.
- Integrate documentation validation (`npx markdownlint-cli2`, `lychee docs`) into CI pipelines if not already enforced by existing workflows.
- Investigate search or navigation helpers (table of contents generation, link checking on PRs) to keep the doc set maintainable.

## Validation & Maintenance

- Schedule periodic link checks against external dependencies to catch upstream URL changes (for example kustomize.io).
- Establish a cadence for reviewing `.lycheeignore` to ensure ignored endpoints still warrant exclusion.
- Add a quarterly doc audit to ensure new KSail-Go features are documented and cross-linked appropriately.
