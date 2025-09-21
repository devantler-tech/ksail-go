<!--
SYNC IMPACT REPORT
Version change: Initial → 1.0.0
Added sections: All core principles, Code Quality Standards, Performance Requirements, Governance
Modified principles: N/A (initial creation)
Removed sections: N/A (initial creation)
Templates requiring updates:
✅ constitution.md (updated)
⚠ plan-template.md (requires Constitution Check section alignment)
⚠ spec-template.md (verify requirements alignment)
⚠ tasks-template.md (ensure task categorization reflects principles)
Follow-up TODOs: None
-->

# KSail Go Constitution

## Core Principles

### I. Code Quality Excellence (NON-NEGOTIABLE)

All code MUST pass comprehensive linting before merge. Code quality is enforced through automated tooling and MUST NOT be compromised. mega-linter with go flavor is the primary quality gate, supplemented by golangci-lint for Go-specific validation. All pull requests MUST achieve zero linting violations before approval.

**Rationale**: Consistent code quality prevents technical debt accumulation and ensures maintainability across a distributed team. Automated enforcement reduces subjective code review friction.

### II. Test-First Development (NON-NEGOTIABLE)

Test-Driven Development (TDD) is mandatory: Tests written → Tests fail → Implementation → Tests pass → Refactor. Unit tests MUST achieve high coverage and integration tests MUST validate all external dependencies and CLI workflows. Snapshot testing is required for all CLI command outputs to ensure consistent user experience.

**Rationale**: TDD ensures robust, regression-free code while snapshot testing guarantees CLI consistency. This approach prevents breaking changes and builds user confidence in the tool.

### III. User Experience Consistency

CLI interfaces MUST provide consistent patterns: help text formatting, error messaging, output formatting, and command structure. All commands MUST follow Cobra framework conventions with standardized flag patterns and human-readable output. UI components (symbols, colors, notifications) MUST use established patterns from `cmd/ui/` package.

**Rationale**: Kubernetes tooling users expect consistent, professional CLI experiences. Inconsistent interfaces create friction and reduce adoption.

### IV. Performance Requirements

Build times MUST remain under 60 seconds, test execution under 120 seconds, and CLI startup under 2 seconds. Performance regressions MUST be caught in CI and addressed before merge. Cluster operations MUST complete within reasonable timeouts (up to 10 minutes for full cluster bootstrap).

**Rationale**: Developer productivity depends on fast feedback loops. Users expect responsive tooling for local development workflows.

### V. Observable and Debuggable

All operations MUST provide clear, actionable output with appropriate log levels. Error messages MUST include context and next steps. CLI commands MUST support both human-readable and structured output formats. Failures MUST provide sufficient information for troubleshooting.

**Rationale**: Kubernetes complexity requires excellent observability. Users need to understand what's happening and how to fix problems when they occur.

## Code Quality Standards

**Linting Requirements**:

- Primary: mega-linter-runner -f go (comprehensive, CI-consistent, auto-fixes enabled)
- Supplementary: golangci-lint (Go-specific fast feedback during development)
- All violations MUST be resolved before merge
- No exceptions without documented technical justification

**Testing Requirements**:

- Unit test coverage MUST be maintained at current levels or higher
- All CLI commands MUST have snapshot tests using go-snaps
- Mock generation via mockery MUST be automated and kept current
- Integration tests MUST cover real CLI workflows

**Documentation Requirements**:

- All packages MUST have comprehensive README files
- Public APIs MUST have Go doc comments
- CLI help text MUST be accurate and helpful
- GitHub Copilot instructions MUST be maintained for agent effectiveness

## Performance Requirements

**Build Performance**:

- `go build` MUST complete under 60 seconds (typical: ~0.5s cached, ~30s first run)
- `go test` MUST complete under 120 seconds (typical: ~37s)
- Dependency resolution MUST complete under 60 seconds (typical: ~0.03s cached)

**Runtime Performance**:

- CLI startup (help/version) MUST complete under 2 seconds
- Cluster operations MAY take up to 10 minutes (with appropriate progress indicators)
- Build processes MUST provide timeout protection to prevent CI hangs

**Resource Constraints**:

- Memory usage MUST remain reasonable for local development environments
- CPU usage during normal operations MUST not interfere with other development tools
- Storage footprint MUST be minimized through efficient dependency management

## Governance

This constitution supersedes all other development practices and policies. All implementation decisions MUST align with these principles. When technical constraints conflict with principles, solutions MUST be architected to minimize violations while documenting necessary compromises.

**Amendment Process**:

- Constitution changes require explicit version bumps following semantic versioning
- MAJOR: Principle removal or fundamental redefinition
- MINOR: New principles or significant guidance expansion
- PATCH: Clarifications, typo fixes, non-semantic refinements

**Compliance Requirements**:

- All pull requests MUST verify principle compliance during review
- Feature implementations MUST pass Constitution Check in implementation plans
- Technical debt that violates principles MUST be tracked and prioritized for resolution
- Use `.github/copilot-instructions.md` for runtime development guidance and tool-specific instructions

**Version**: 1.0.0 | **Ratified**: 2025-09-21 | **Last Amended**: 2025-09-21
