# Contracts — 001-migrate-docs

This documentation-only feature introduces no new service APIs, CLI flags, or external contracts. All work items operate on Markdown sources inside the repository.

## Implications

- No OpenAPI/GraphQL schemas to maintain for this change.
- CLI behavior is referenced for documentation accuracy only; existing commands remain unchanged in code.
- Validation responsibilities reside in markdownlint and lychee checks documented in the plan.
