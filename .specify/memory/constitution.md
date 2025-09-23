<!--
Sync Impact Report:
- Version change: 1.0.0 → 1.1.0
- Modified principles: None (validation demonstrates successful adherence)
- Added sections: None
- Removed sections: None
- Implementation milestone: Configuration validation system successfully implemented following TDD principles
- Templates requiring updates:
  ✅ constitution.md updated
  ✅ plan-template.md aligned (version reference updated)
  ⚠ spec-template.md verified (no updates needed)
  ⚠ tasks-template.md verified (no updates needed)
- Follow-up TODOs: none
-->

# KSail Constitution

## Core Principles

### I. Code Quality First

All code MUST pass comprehensive quality gates before merge. Every package follows Go best practices with mandatory linting via golangci-lint. Code MUST be self-documenting through clear naming, structured organization in cmd/ and pkg/ directories, and comprehensive godoc comments. Dependencies are strictly controlled via depguard rules, and all auto-generated files (mocks) are excluded from quality checks.

Rationale: KSail manages critical Kubernetes infrastructure. Poor code quality leads to unreliable cluster operations, difficult debugging, and user frustration.

### II. Test-Driven Development (NON-NEGOTIABLE)

TDD is mandatory: Tests written → User approved → Tests fail → Then implement. Red-Green-Refactor cycle strictly enforced. Every package MUST have corresponding *_test.go files. System tests validate complete workflows (init → up → status → list → start → reconcile → down) across all supported distributions (Kind, K3d, EKS). Snapshot testing via go-snaps ensures CLI output consistency.

Successfully demonstrated in the configuration validation system implementation (September 2025), where contract tests were written first, failed as expected, then implementation followed to make tests pass.

Rationale: Kubernetes tooling failure has severe consequences. TDD ensures robust functionality and prevents regressions in complex orchestration scenarios.

### III. User Experience Consistency

CLI interface MUST provide predictable, intuitive workflows across all Kubernetes distributions. Commands follow consistent patterns: ksail [action] with standardized help text, error messages, and output formatting. Support both JSON and human-readable formats. UI feedback uses standardized notify package for consistent error/success reporting. All operations provide clear progress indication and meaningful error messages.

Rationale: Users switch between different Kubernetes distributions frequently. Inconsistent interfaces create cognitive overhead and reduce productivity.

### IV. Performance Excellence

Cluster operations MUST complete within defined SLA boundaries: cluster creation <3 minutes, status checks <10 seconds, reconciliation <5 minutes. Memory usage MUST remain under reasonable limits during concurrent operations. Build times MUST stay under 90 seconds, test execution under 60 seconds. All external tool dependencies (kubectl, helm, flux) MUST be efficiently managed to avoid unnecessary overhead.

Rationale: Developer productivity depends on fast feedback loops. Slow operations interrupt development workflow and reduce adoption.

## Quality Standards

**Go Version**: Minimum Go 1.24.0 as specified in go.mod
**Testing Coverage**: Minimum 80% code coverage with comprehensive integration tests
**Linting**: Zero tolerance for golangci-lint violations using project .golangci.yml configuration
**Documentation**: All public functions, types, and packages MUST have godoc comments
**Dependencies**: External dependencies limited to approved list in depguard configuration
**Security**: All dependencies scanned via Trivy, SOPS integration for secrets management
**Build Validation**: All builds MUST pass mega-linter comprehensive validation before release

## Development Workflow

**Branch Strategy**: Feature branches from main with descriptive names (test/feature-name)
**Code Review**: All PRs require review and MUST pass CI checks including system tests
**Testing Gates**: Unit tests, integration tests, and system tests MUST pass before merge
**Release Process**: Semantic versioning with automated releases via GoReleaser
**Documentation**: Changes affecting user workflow MUST update corresponding documentation
**Compliance Verification**: Every PR MUST verify constitution compliance via checklist review

## Governance

This constitution supersedes all other development practices and guidelines. All code reviews MUST verify compliance with these principles. Technical complexity MUST be justified against these standards - if a change violates principles without clear necessity, it MUST be simplified or rejected.

Amendments require:

1. Documentation of rationale and impact assessment
2. Review and approval from project maintainers
3. Migration plan for existing code that conflicts with new requirements
4. Update of all dependent templates and guidance files

Use `.github/copilot-instructions.md` for runtime development guidance and tool-specific implementation details.

**Version**: 1.1.0 | **Ratified**: 2025-09-22 | **Last Amended**: 2025-09-23
