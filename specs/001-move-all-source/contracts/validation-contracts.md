# Validation Contracts

**Feature**: Move All Go Source Code to src/
**Phase**: 1 - Design
**Date**: 2025-11-16

## Overview

This document defines the validation contracts that must be satisfied at each checkpoint during the reorganization. These are not API contracts but rather validation contracts ensuring the reorganization maintains system integrity.

## Pre-Move Validation Contract

**Checkpoint**: Before any files are moved
**Purpose**: Establish baseline and ensure starting state is valid

### Required Checks

```yaml
- name: "Baseline Build"
  command: "go build ./..."
  working_directory: "."
  expected_exit_code: 0
  description: "Verify project builds successfully in current state"

- name: "Baseline Tests"
  command: "go test ./..."
  working_directory: "."
  expected_exit_code: 0
  description: "Verify all tests pass in current state"

- name: "Baseline Lint"
  command: "golangci-lint run --timeout 5m"
  working_directory: "."
  expected_exit_code: 0
  description: "Verify code quality standards met in current state"

- name: "Baseline Build Time"
  command: "time go build -o bin/ksail ."
  working_directory: "."
  expected_exit_code: 0
  capture: "build_time"
  description: "Capture baseline build time for comparison"

- name: "Clean Git State"
  command: "git status --porcelain"
  working_directory: "."
  expected_output: ""
  description: "Verify no uncommitted changes before reorganization"
```

**Success Criteria**: All checks must pass with exit code 0

## Post-Move Validation Contract

**Checkpoint**: After files are moved but before commit
**Purpose**: Verify reorganized structure works correctly

### Post-Move Required Checks

```yaml
- name: "Post-Move Build"
  command: "go build ./..."
  working_directory: "src"
  expected_exit_code: 0
  description: "Verify project builds after reorganization"

- name: "Post-Move Tests"
  command: "go test ./..."
  working_directory: "src"
  expected_exit_code: 0
  description: "Verify all tests pass after reorganization"

- name: "Post-Move Lint"
  command: "golangci-lint run --timeout 5m"
  working_directory: "."
  expected_exit_code: 0
  description: "Verify code quality maintained after reorganization"

- name: "Post-Move Build Time"
  command: "time go build -o ../bin/ksail ."
  working_directory: "src"
  expected_exit_code: 0
  compare_to: "build_time"
  tolerance: "5%"
  description: "Verify build time unchanged (within 5% tolerance)"

- name: "Mock Generation"
  command: "mockery"
  working_directory: "."
  expected_exit_code: 0
  description: "Verify mock generation works with new structure"

- name: "Binary Output Location"
  command: "test -f bin/ksail"
  working_directory: "."
  expected_exit_code: 0
  description: "Verify binary output in correct location"

- name: "Module Path Validation"
  command: "grep 'module github.com/devantler-tech/ksail-go' src/go.mod"
  working_directory: "."
  expected_exit_code: 0
  description: "Verify module path unchanged in go.mod"
```

**Success Criteria**: All checks must pass, build time within tolerance

## Pre-Merge Validation Contract

**Checkpoint**: After all changes committed, before merging to main
**Purpose**: Comprehensive validation including CI/CD simulation

### Pre-Merge Required Checks

```yaml
- name: "Full Build"
  command: "go build ./..."
  working_directory: "src"
  expected_exit_code: 0
  description: "Full project build"

- name: "Full Test Suite"
  command: "go test ./..."
  working_directory: "src"
  expected_exit_code: 0
  description: "Complete test suite execution"

- name: "Full Lint"
  command: "golangci-lint run --timeout 5m"
  working_directory: "."
  expected_exit_code: 0
  description: "Complete linting validation"

- name: "VS Code Build Task"
  command: "# Simulate VS Code task"
  description: "Verify VS Code tasks work with updated paths"
  manual: true

- name: "VS Code Test Task"
  command: "# Simulate VS Code task"
  description: "Verify VS Code test task works"
  manual: true

- name: "Git History Check"
  command: "git log --follow src/main.go | head -n 20"
  working_directory: "."
  expected_exit_code: 0
  description: "Verify git history preserved for moved files"

- name: "External Import Test"
  command: "# Create test project importing ksail-go packages"
  description: "Verify external consumers can import packages"
  manual: true
  external: true
```

**Success Criteria**: All automated checks pass, manual checks documented as successful

## Post-Merge Validation Contract

**Checkpoint**: After merging to main branch
**Purpose**: Verify production state and CI/CD pipeline

### Post-Merge Required Checks

```yaml
- name: "CI Build Status"
  description: "Verify GitHub Actions ci.yaml workflow passed"
  source: "GitHub Actions API"
  expected: "success"

- name: "CI Test Status"
  description: "Verify all CI tests passed"
  source: "GitHub Actions API"
  expected: "success"

- name: "CI Lint Status"
  description: "Verify CI linting passed"
  source: "GitHub Actions API"
  expected: "success"

- name: "Release Build Test"
  description: "Verify GoReleaser can build from new structure"
  source: "GitHub Actions workflow"
  expected: "success"
  note: "Can be simulated with goreleaser release --snapshot --clean"

- name: "Schema Generation"
  description: "Verify schema generation scripts work"
  command: ".github/scripts/generate-schema.sh"
  working_directory: "."
  expected_exit_code: 0

- name: "Coverage Report"
  description: "Verify code coverage reporting works with new paths"
  source: "codecov.io"
  expected: "report generated"
```

**Success Criteria**: All CI/CD pipelines green, no production issues

## Rollback Contract

**Checkpoint**: If any validation fails
**Purpose**: Define clear rollback procedure

### Rollback Steps

```yaml
- name: "Identify Failure"
  description: "Document which validation check failed and why"
  required: true

- name: "Revert Merge Commit"
  command: "git revert -m 1 <merge-commit-sha>"
  description: "Revert the merge commit to restore main branch"
  condition: "post-merge failure"

- name: "Discard Local Changes"
  command: "git reset --hard HEAD"
  description: "Discard uncommitted changes"
  condition: "pre-commit failure"

- name: "Create Fix Issue"
  description: "Document the failure and create issue for corrected implementation"
  required: true

- name: "New Fix PR"
  description: "Create new pull request with fixes addressing the failure"
  required: true
```

**Success Criteria**: System restored to working state, failure documented, fix planned

## Contract Compliance

All validation contracts are mandatory. No exceptions or partial passes allowed. Any failure triggers the Rollback Contract.

## Notes

- Validation contracts enforce the "both pre-merge and post-merge validation" requirement from the specification
- Build time tolerance allows for minor environment variations while catching significant regressions
- Manual checks acknowledge some validations require human verification (IDE, external consumer testing)
