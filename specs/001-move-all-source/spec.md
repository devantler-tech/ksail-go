# Feature Specification: Move All Go Source Code to src/

**Feature Branch**: `001-move-all-source`
**Created**: 2025-11-15
**Status**: Draft
**Input**: User description: "move all go source code to src/"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Building the Project (Priority: P1)

A developer clones the repository and builds the project using standard Go build commands. The project builds successfully with all source code organized under the `src/` directory.

**Why this priority**: This is the most critical user journey as it affects every developer interaction with the codebase. Without successful builds, no development work can proceed.

**Independent Test**: Can be fully tested by running `go build ./...` and `go test ./...` from the repository root and verifying all packages compile and tests pass.

**Acceptance Scenarios**:

1. **Given** a fresh clone of the repository, **When** developer runs `go build ./...`, **Then** all packages build successfully
2. **Given** the reorganized structure, **When** developer runs `go test ./...`, **Then** all tests execute and pass
3. **Given** the new structure, **When** developer runs `go run .`, **Then** the main application starts correctly

---

### User Story 2 - Developer Working with IDE (Priority: P2)

A developer opens the project in their IDE (VS Code, GoLand, etc.) and all Go tooling (autocomplete, go to definition, refactoring tools) continues to work seamlessly with the new directory structure.

**Why this priority**: IDE functionality is essential for productive development, but the project can still be built and run from the command line if IDE support temporarily breaks.

**Independent Test**: Can be tested by opening the project in VS Code and verifying that Go extension features (autocomplete, diagnostics, navigation) work correctly.

**Acceptance Scenarios**:

1. **Given** the project open in VS Code, **When** developer types code in any Go file, **Then** autocomplete suggestions appear correctly
2. **Given** the project open in an IDE, **When** developer uses "Go to Definition" on an import, **Then** navigation works correctly
3. **Given** the new structure, **When** developer runs the build task in VS Code, **Then** the build completes successfully

---

### User Story 3 - CI/CD Pipeline Execution (Priority: P1)

The CI/CD pipeline runs automated builds, tests, linting, and releases. All workflows continue to function correctly with the new directory structure.

**Why this priority**: CI/CD pipeline failure blocks all code merges and releases. This is critical for maintaining development velocity.

**Independent Test**: Can be tested by pushing changes to a feature branch and verifying all GitHub Actions workflows pass.

**Acceptance Scenarios**:

1. **Given** a pull request with the new structure, **When** CI workflow runs, **Then** all build, test, and lint steps pass
2. **Given** the reorganized codebase, **When** GoReleaser workflow executes, **Then** binaries are built and published successfully
3. **Given** the new import paths, **When** schema generation script runs, **Then** schemas are generated correctly

---

### User Story 4 - External Package Consumers (Priority: P3)

Developers using KSail as a Go library in their own projects can continue to import and use packages without breaking changes to public API import paths.

**Why this priority**: While important for library consumers, this affects external users rather than internal development. The project can function as a CLI tool without this.

**Independent Test**: Can be tested by creating a sample Go project that imports KSail packages and verifying imports resolve correctly.

**Acceptance Scenarios**:

1. **Given** an external project importing `github.com/devantler-tech/ksail-go/pkg/...`, **When** they build their project, **Then** imports resolve correctly
2. **Given** the new structure, **When** external developers use go doc, **Then** package documentation remains accessible

---

### Edge Cases

- What happens when build tools reference absolute paths to source files?
- How do relative imports within test files handle the new structure?
- What happens to code coverage report paths after the move?
- How do debugging tools handle source file paths in the new structure?
- What happens to git history and blame information for moved files?

## Requirements *(mandatory)*

### Assumptions

- The project follows standard language ecosystem conventions for module and package management
- Existing tooling supports configurable source directory paths
- The development team has access to appropriate version control migration commands
- Current import paths are module-relative rather than absolute file paths
- IDE and editor tools support custom source directory configurations
- Automated pipelines use configurable path settings rather than hardcoded locations
- The binary output location is configurable in build tooling

### Functional Requirements

- **FR-001**: All source code files MUST be moved from the repository root to a `src/` directory while maintaining the existing package structure
- **FR-002**: The main application entry point and its tests MUST be moved to src/ directory
- **FR-003**: All module references and import paths MUST remain unchanged to maintain backward compatibility
- **FR-004**: Package manager configuration files MUST remain at the repository root to comply with ecosystem conventions
- **FR-005**: All build and test commands MUST continue to work without requiring changes to command syntax
- **FR-006**: All existing test files MUST continue to execute successfully after the reorganization
- **FR-007**: IDE tooling and editor configurations MUST be updated to reference the new source locations
- **FR-008**: Automated build and deployment pipelines MUST be updated to build and test from the new directory structure
- **FR-009**: Binary compilation configuration MUST be updated to reference the new application entry point location
- **FR-010**: Documentation files MUST remain at the repository root for visibility and accessibility
- **FR-011**: Project configuration files MUST remain at the repository root or be updated to reference new paths
- **FR-012**: Version control history MUST be preserved to maintain file lineage and change tracking information
- **FR-013**: The binary output directory MUST remain at the repository root or be explicitly configured
- **FR-014**: All automation scripts that reference source files MUST be updated to use the new paths

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All source code files are located under the `src/` directory with zero source files remaining in the repository root (only package manager and configuration files remain at root)
- **SC-002**: All build and test commands complete successfully with zero errors
- **SC-003**: All automated deployment pipelines pass without modification to test logic, only path configuration updates
- **SC-004**: All IDE automation tasks execute successfully from the workspace
- **SC-005**: Code coverage reports generate successfully with correct file path references
- **SC-006**: Binary compilation succeeds for all target platforms from the new structure
- **SC-007**: Version control change tracking remains intact for all moved files, allowing developers to trace file history across the reorganization
- **SC-008**: All code quality validation tools execute successfully against the new structure
- **SC-009**: Test helper code generation completes successfully with outputs created in the correct locations
- **SC-010**: Documentation generation and validation scripts execute without errors
