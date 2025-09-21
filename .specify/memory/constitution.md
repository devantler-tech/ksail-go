<!-- Sync Impact Report:
Version change: N/A → 1.0.0 (initial constitution)
Added sections:
- I. Library-First Architecture
- II. CLI-Driven Interface
- III. Test-First Development (NON-NEGOTIABLE)
- IV. Comprehensive Testing Strategy
- V. Clean Architecture & Interfaces
- Development Standards
- Quality Assurance

Templates requiring updates:
✅ .specify/templates/plan-template.md (Constitution Check section aligns)
✅ .specify/templates/spec-template.md (functional requirements align)
✅ .specify/templates/tasks-template.md (task categorization aligns)
✅ .specify/templates/agent-file-template.md (compatible)

Follow-up TODOs: None
-->

# KSail Go Constitution

## Core Principles

### I. Library-First Architecture

Every feature MUST start as a standalone library package in `pkg/`. Libraries MUST be self-contained, independently testable, well-documented with comprehensive README files, and have a clear single purpose. No organizational-only libraries are permitted - each package must provide concrete functionality that can be used independently.

**Rationale**: Promotes modularity, reusability, and maintainability. Enables easier testing and allows components to be extracted or reused in other projects if needed.

### II. CLI-Driven Interface

Every library MUST expose its core functionality via CLI commands using the Cobra framework. All commands MUST follow a consistent text in/out protocol: configuration via flags/arguments/stdin → results to stdout, errors to stderr. Commands MUST support both human-readable and machine-readable output formats where applicable.

**Rationale**: Ensures consistent user experience, enables automation and scripting, and provides clear boundaries between user interface and business logic.

### III. Test-First Development (NON-NEGOTIABLE)

Test-Driven Development is MANDATORY for all new functionality. The cycle MUST be: write failing tests → implement minimal code to pass tests → refactor for quality. All tests MUST pass before any code is merged. No exceptions.

**Rationale**: Ensures code quality, prevents regressions, forces clear thinking about requirements, and provides living documentation of expected behavior.

### IV. Comprehensive Testing Strategy

Testing MUST cover multiple levels: unit tests for all business logic, snapshot testing for CLI command outputs using go-snaps, integration tests for external dependencies using mocks, and comprehensive test coverage with proper naming conventions. Test functions MUST follow Go community standards: `TestMethodName` pattern with `t.Run` subtests for scenarios.

**Rationale**: Provides confidence in changes, catches regressions early, and ensures consistent behavior across different environments and use cases.

### V. Clean Architecture & Interfaces

All packages MUST follow clean architecture principles with clear separation of concerns. Heavy use of interfaces is REQUIRED for testability and extensibility. Components MUST accept dependencies as interfaces, support context for cancellation/timeouts, and maintain clear boundaries between domains.

**Rationale**: Enables maintainable and testable code, supports dependency injection, and allows for easy mocking and testing of complex interactions.

## Development Standards

All code MUST pass mega-linter-runner with Go flavor configuration before merge. This includes formatting, linting, security scanning, and quality checks. GoLangCI-lint MAY be used for faster feedback during development but mega-linter is the authoritative quality gate.

Repository structure MUST follow established Go conventions: `cmd/` for CLI implementations, `pkg/` for reusable libraries, `internal/` for private utilities. Each package MUST include comprehensive README documentation.

Mock generation using mockery is REQUIRED when interface definitions change. All mocks MUST be generated and committed alongside interface changes.

**Rationale**: Ensures consistent code quality, maintainability, and follows Go community best practices while providing clear project structure.

## Quality Assurance

Snapshot testing using go-snaps is MANDATORY for all CLI command outputs. Any changes to command output, help text, or error messages REQUIRE snapshot updates. Snapshots provide regression protection and document expected behavior.

Build verification MUST include: tests pass, code builds successfully, CLI functionality works, linting passes. The complete validation sequence MUST be run before any changes are merged.

Performance considerations: Build times (~0.5s first time, ~0.1s cached) and test execution (~37s) are monitored. Timeouts MUST be set appropriately to prevent premature cancellation of long-running operations.

**Rationale**: Maintains consistent user experience, prevents unintended changes to public interfaces, and ensures reliability across different environments.

## Governance

This constitution supersedes all other development practices and guidelines. All pull requests and code reviews MUST verify compliance with these principles. Any complexity or deviation MUST be explicitly justified and documented.

Amendment of this constitution requires: documented rationale for the change, community review and approval, migration plan for existing code if applicable, and update of all dependent templates and documentation.

Version management follows semantic versioning: MAJOR for backward incompatible governance changes, MINOR for new principles or expanded guidance, PATCH for clarifications and non-semantic refinements.

Use `.github/copilot-instructions.md` for detailed runtime development guidance and tool-specific instructions.

**Version**: 1.0.0 | **Ratified**: 2025-09-21 | **Last Amended**: 2025-09-21
