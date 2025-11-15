# Data Model: Move All Go Source Code to src/

**Feature**: Move All Go Source Code to src/
**Phase**: 1 - Design
**Date**: 2025-11-16

## Overview

This feature involves a structural reorganization with no new data entities. The "data" in this context refers to the repository structure and file system organization.

## Entities

### FileSystemEntry

Represents a file or directory in the repository that needs to be handled during reorganization.

**Attributes**:

- `path`: string - Current path relative to repository root
- `type`: enum(file, directory) - Whether this is a file or directory
- `category`: enum(source, config, artifact, documentation) - Classification of the entry
- `action`: enum(move, stay, update_reference) - What should happen to this entry
- `targetPath`: string (optional) - New path if action is "move"

**Validation Rules**:

- `path` must be a valid relative path
- If `action` is "move", `targetPath` must be specified
- Source files must have `category` = "source"
- Config files must have `category` = "config"

**State Transitions**:

1. Initial → Analyzed (after categorization)
2. Analyzed → Moved (after git mv execution)
3. Analyzed → Updated (after reference update for configs)
4. Analyzed → Unchanged (if action is "stay")

### ConfigurationReference

Represents a configuration file that contains references to source code paths.

**Attributes**:

- `filePath`: string - Path to the configuration file
- `type`: enum(vscode_task, github_workflow, go_config, script) - Type of configuration
- `references`: array of PathReference - List of path references that need updating

**PathReference Sub-entity**:

- `originalPath`: string - Current path reference
- `newPath`: string - Updated path reference
- `context`: string - The line or section where the reference appears

**Validation Rules**:

- All `references` must have valid `originalPath` and `newPath`
- `filePath` must exist in repository
- After update, file must remain valid (syntax check)

### ValidationCheckpoint

Represents a validation checkpoint in the migration process.

**Attributes**:

- `stage`: enum(pre_move, post_move, pre_merge, post_merge) - When validation occurs
- `checks`: array of ValidationCheck - Individual validation checks
- `status`: enum(pending, running, passed, failed) - Current status
- `timestamp`: datetime - When validation was performed

**ValidationCheck Sub-entity**:

- `name`: string - Name of the check (e.g., "go build", "go test", "golangci-lint")
- `command`: string - Command executed
- `workingDirectory`: string - Where command runs
- `expectedExitCode`: integer - Expected exit code (usually 0)
- `actualExitCode`: integer (optional) - Actual exit code after execution
- `output`: string (optional) - Command output
- `passed`: boolean - Whether check passed

**Validation Rules**:

- All checks in a checkpoint must pass for status to be "passed"
- If any check fails, checkpoint status is "failed"
- Validation must occur at all defined stages (pre_move, post_move, pre_merge, post_merge)

## Relationships

```text
FileSystemEntry 1--* ConfigurationReference
  (a file system entry may be referenced by multiple configs)

ValidationCheckpoint *--* ValidationCheck
  (a checkpoint contains multiple validation checks)

ConfigurationReference 1--* PathReference
  (a config file may contain multiple path references)
```

## File Categories

### Source Code (action: move to src/)

- All .go files in cmd/, pkg/, internal/
- main.go, main_test.go
- go.mod, go.sum

### Configuration (action: stay at root, may need reference updates)

- `.github/` (workflows, scripts) - needs reference updates
- `.vscode/` (tasks, settings) - needs reference updates
- `.golangci.yml` - may need path updates
- `.mockery.yml` - may need path updates
- `.goreleaser.yaml` - needs path updates
- All other dotfiles - stay unchanged

### Artifacts (action: stay at root or reconfigure output)

- `bin/` - stays at root, output path configured
- `coverage.txt` - generated file, path may need update

### Documentation (action: stay at root)

- `README.md`, `CONTRIBUTING.md`, `LICENSE`
- `docs/` directory
- `specs/` directory

### Samples/Examples (action: stay at root)

- `k8s/` - Kubernetes manifests
- `kind.yaml`, `k3d.yaml`, `ksail.yaml` - cluster configs
- `schemas/` - JSON schemas

## Domain Rules

1. **Atomicity**: All file moves must happen in a single atomic operation (one PR)
2. **History Preservation**: All moves must use `git mv` to preserve history
3. **Import Path Immutability**: External import paths must remain unchanged
4. **Build Reproducibility**: Build times and outputs must remain identical
5. **Validation Gate**: All validation checkpoints must pass before proceeding
6. **Rollback Safety**: Any checkpoint failure triggers immediate rollback via git revert

## Notes

This is a structural feature with no business logic or runtime data. All "entities" represent the repository state and validation artifacts, not application domain models.
