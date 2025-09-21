<!-- Sync Impact Report:
Version change: 1.0.0 → 1.1.0 (minor: added new principles and expanded guidance)
Modified principles:
- Enhanced IV. Comprehensive Testing Strategy with function length limits
- Expanded Quality Assurance section with performance requirements

Added sections:
- VI. User Experience Consistency
- VII. Performance and Reliability Standards
- Expanded Development Standards with specific linting requirements

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

Testing MUST cover multiple levels: unit tests for all business logic, snapshot testing for CLI command outputs using go-snaps, integration tests for external dependencies using mocks, and comprehensive test coverage with proper naming conventions. Test functions MUST follow Go community standards: `TestMethodName` pattern with `t.Run` subtests for scenarios. All functions MUST stay under 60 lines to comply with funlen linting standards - break large test functions into smaller, focused test functions when needed.

**Rationale**: Provides confidence in changes, catches regressions early, ensures consistent behavior across different environments and use cases, and maintains code readability through manageable function sizes.

### V. Clean Architecture & Interfaces

All packages MUST follow clean architecture principles with clear separation of concerns. Heavy use of interfaces is REQUIRED for testability and extensibility. Components MUST accept dependencies as interfaces, support context for cancellation/timeouts, and maintain clear boundaries between domains.

**Rationale**: Enables maintainable and testable code, supports dependency injection, and allows for easy mocking and testing of complex interactions.

### VI. User Experience Consistency

All CLI commands MUST provide consistent user interfaces with standardized flag naming conventions, uniform output formatting, and predictable error messaging. Help text MUST be comprehensive and follow established patterns. CLI responses MUST be deterministic and suitable for both human consumption and automation. All user-facing text MUST be validated through snapshot testing to prevent unintended changes.

**Rationale**: Creates a predictable and learnable interface that reduces cognitive load for users and enables reliable automation and scripting.

### VII. Performance and Reliability Standards

All operations MUST complete within expected timeframes: builds (~0.5s first time, ~0.1s cached), full test suite (~37s), and linting operations MUST have appropriate timeouts set. Long-running operations MUST NEVER be prematurely cancelled and MUST provide progress indicators where applicable. Resource usage MUST be reasonable for typical development environments. Error handling MUST be comprehensive with clear, actionable error messages.

**Rationale**: Ensures reliable development workflows, prevents frustration from timeouts or resource exhaustion, and provides clear feedback when issues occur.

## Development Standards

All code MUST pass mega-linter-runner with Go flavor configuration before merge. This includes formatting, linting, security scanning, and quality checks. GoLangCI-lint with funlen enforcement (60-line function limit) MUST also pass - functions exceeding this limit MUST be refactored into smaller, focused functions. Both mega-linter and golangci-lint are required quality gates with mega-linter being the authoritative comprehensive check.

Repository structure MUST follow established Go conventions: `cmd/` for CLI implementations, `pkg/` for reusable libraries, `internal/` for private utilities. Each package MUST include comprehensive README documentation with usage examples and API documentation.

Mock generation using mockery is REQUIRED when interface definitions change. All mocks MUST be generated and committed alongside interface changes. Configuration files (`.mockery.yml`, `.golangci.yml`, `.mega-linter.yml`) MUST be maintained to ensure consistent quality enforcement.

**Rationale**: Ensures consistent code quality, maintainability, follows Go community best practices, and enforces function size limits for better readability and maintainability.

## Quality Assurance

Snapshot testing using go-snaps is MANDATORY for all CLI command outputs. Any changes to command output, help text, or error messages REQUIRE snapshot updates. Snapshots provide regression protection and document expected behavior for user experience consistency.

Build verification MUST include: tests pass, code builds successfully, CLI functionality works, both mega-linter and golangci-lint pass. The complete validation sequence MUST be run before any changes are merged. Performance benchmarks MUST be monitored for regressions.

Performance requirements: Build times (~0.5s first time, ~0.1s cached) and test execution (~37s) are monitored and MUST NOT regress significantly. Timeouts MUST be set appropriately (60s+ for builds, 120s+ for tests) to prevent premature cancellation of long-running operations. Memory usage MUST remain reasonable for development environments.

**Rationale**: Maintains consistent user experience, prevents unintended changes to public interfaces, ensures reliability across different environments, and provides predictable performance characteristics.

## Governance

This constitution supersedes all other development practices and guidelines. All pull requests and code reviews MUST verify compliance with these principles. Any complexity or deviation MUST be explicitly justified and documented with clear rationale for why the deviation serves the project's goals.

Amendment of this constitution requires: documented rationale for the change, community review and approval, migration plan for existing code if applicable, and update of all dependent templates and documentation. Changes affecting user experience or performance standards require additional validation.

Version management follows semantic versioning: MAJOR for backward incompatible governance changes, MINOR for new principles or expanded guidance, PATCH for clarifications and non-semantic refinements. Technical decisions MUST be guided by these principles with explicit justification when trade-offs are necessary.

Use `.github/copilot-instructions.md` for detailed runtime development guidance and tool-specific instructions. This governance model ensures that technical decisions consistently support code quality, testing standards, user experience consistency, and performance requirements.

**Version**: 1.1.0 | **Ratified**: 2025-09-21 | **Last Amended**: 2025-09-21
