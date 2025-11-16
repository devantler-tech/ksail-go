# Tasks: Move All Go Source Code to src/

**Input**: Design documents from `/specs/001-move-all-source/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/validation-contracts.md, quickstart.md

**Tests**: Tests are NOT required for this feature as it is a structural reorganization with no new APIs. Validation is achieved through comprehensive validation contracts at each checkpoint.

**Organization**: Tasks are organized sequentially as this is an atomic operation (single PR). All file moves happen together, followed by configuration updates, then validation.

## Format: `- [ ] [ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Tasks without [P] must complete before subsequent tasks begin
- All file paths are relative to repository root `<repository-root>`

---

## Phase 1: Pre-Move Validation (Establish Baseline)

**Purpose**: Establish baseline metrics and verify starting state is valid

- [ ] T001 Verify git working directory is clean: `git status --porcelain` should return empty output
- [ ] T002 Run baseline build: `go build ./...` from repository root, verify exit code 0
- [ ] T003 Run baseline tests: `go test ./...` from repository root, verify exit code 0
- [ ] T004 Run baseline lint: `golangci-lint run --timeout 5m` from repository root, verify exit code 0
- [ ] T005 Capture baseline build time: `time go build -o bin/ksail .` from repository root, record duration for comparison
- [ ] T006 Verify mockery works in current state: `mockery` from repository root, verify exit code 0

**Checkpoint**: All baseline checks pass - ready to proceed with reorganization

---

## Phase 2: File Reorganization (Atomic Move)

**Purpose**: Move all Go source code and module files to src/ directory using git mv

**⚠️ CRITICAL**: Do NOT commit until all moves and configuration updates are complete

- [ ] T007 Create src directory: `mkdir src` in repository root
- [ ] T008 Move cmd directory: `git mv cmd src/` to preserve git history
- [ ] T009 Move pkg directory: `git mv pkg src/` to preserve git history
- [ ] T010 Move internal directory: `git mv internal src/` to preserve git history
- [ ] T011 Move main.go: `git mv main.go src/` to preserve git history
- [ ] T012 Move main_test.go: `git mv main_test.go src/` to preserve git history
- [ ] T013 Move go.mod: `git mv go.mod src/` to move module root
- [ ] T014 Move go.sum: `git mv go.sum src/` to move module checksums

**Checkpoint**: All source files moved to src/ with git tracking renames

---

## Phase 3: Configuration Updates (Path References)

**Purpose**: Update all configuration files to reference new src/ paths

### VS Code Configuration Updates

- [ ] T015 Update .vscode/tasks.json - go:build task: Change `cwd` from `${workspaceFolder}` to `${workspaceFolder}/src` in .vscode/tasks.json
- [ ] T016 Update .vscode/tasks.json - go:test task: Change `cwd` from `${workspaceFolder}` to `${workspaceFolder}/src` in .vscode/tasks.json
- [ ] T017 Update .vscode/tasks.json - go:fmt task: Change `cwd` from `${workspaceFolder}` to `${workspaceFolder}/src` in .vscode/tasks.json (if task references go fmt)
- [ ] T018 Update .vscode/tasks.json - go:lint task: Update golangci-lint `cwd` if needed (linter runs from root, may not need change)

### GitHub Workflows Configuration Updates

- [ ] T019 [P] Update .github/workflows/ci.yaml: Add `working-directory: src` to all Go build steps
- [ ] T020 [P] Update .github/workflows/ci.yaml: Add `working-directory: src` to all Go test steps
- [ ] T021 [P] Update .github/workflows/ci.yaml: Change `go-version-file` from `'go.mod'` to `'src/go.mod'` in actions/setup-go@v5 step
- [ ] T022 [P] Update .github/workflows/cd.yaml: Add `working-directory: src` to Go build/test steps if present
- [ ] T023 [P] Update .github/workflows/cd.yaml: Change `go-version-file` to `'src/go.mod'` in actions/setup-go@v5 step
- [ ] T024 [P] Update .github/workflows/release.yaml: Add `working-directory: src` to Go build steps if present
- [ ] T025 [P] Update .github/workflows/release.yaml: Change `go-version-file` to `'src/go.mod'` in actions/setup-go@v5 step

### GoReleaser Configuration Updates

- [ ] T026 Update .goreleaser.yaml: Change `main:` from `'.'` to `'./src'` in builds section for ksail binary
- [ ] T027 Update .goreleaser.yaml: Verify binary output path remains `bin/ksail` (or update if needed)
- [ ] T028 Update .goreleaser.yaml: Update any other path references to Go source files if present

### Scripts and Tools Configuration Updates

- [ ] T029 [P] Update .github/scripts/generate-schema.sh: Add `cd src` at the start or update go run paths
- [ ] T030 [P] Check .mockery.yml: Update paths if they reference absolute source locations (change to src/ prefix if needed)
- [ ] T031 [P] Update .github/scripts/run-golangci-lint.sh: Verify it runs from repository root (should not need changes)
- [ ] T032 [P] Update .github/scripts/run-mockery.sh: Verify mockery can find Go files in src/ (test after Phase 4)

**Checkpoint**: All configuration files updated to reference src/ paths

---

## Phase 4: Post-Move Validation (Verify Reorganization)

**Purpose**: Verify everything works with new structure before committing

- [ ] T033 Test build from src: `cd src && go build ./...` verify exit code 0
- [ ] T034 Test tests from src: `cd src && go test ./...` verify exit code 0
- [ ] T035 Test lint from root: `golangci-lint run --timeout 5m` from repository root, verify exit code 0
- [ ] T036 Test binary build to bin: `cd src && go build -o ../bin/ksail .` verify exit code 0
- [ ] T037 Verify binary exists: `ls -l bin/ksail` verify file exists at repository root
- [ ] T038 Test mockery: `mockery` from repository root, verify exit code 0 and mocks generated
- [ ] T039 Verify module path unchanged: `grep 'module github.com/devantler-tech/ksail-go' src/go.mod` verify exit code 0
- [ ] T040 Measure post-move build time: `cd src && time go build -o ../bin/ksail .` compare to T005 baseline (must be within 5% tolerance)
- [ ] T041 Test go commands with -C flag: `go -C src build ./...` verify exit code 0 (Go 1.20+ feature)

**Checkpoint**: All post-move validation checks pass - ready to commit

---

## Phase 5: Commit and Pre-Merge Validation

**Purpose**: Commit changes and prepare for merge

- [ ] T042 Stage all changes: `git add -A` to stage file moves and configuration updates
- [ ] T043 Commit with comprehensive message: Use commit message from quickstart.md including all changes and validation results
- [ ] T044 Run full test suite: `cd src && go test ./...` verify all tests pass
- [ ] T045 Run full lint: `golangci-lint run --timeout 5m` verify no errors
- [ ] T046 Verify git history preserved: `git log --follow src/main.go | head -n 20` verify history shows commits before move
- [ ] T047 Test VS Code build task: Open VS Code, reload workspace, run "go: build" task, verify success
- [ ] T048 Test VS Code test task: Run "go: test" task in VS Code, verify success
- [ ] T049 Test VS Code fmt task: Run "go: fmt" task in VS Code, verify success (if exists)
- [ ] T050 Test VS Code lint task: Run "go: lint" task in VS Code, verify success (if exists)

**Checkpoint**: All pre-merge checks pass, commit created, ready to push

---

## Phase 6: Push and CI Validation

**Purpose**: Push to feature branch and verify CI/CD pipelines

- [ ] T051 Push to feature branch: `git push origin 001-move-all-source` to trigger CI workflows
- [ ] T052 Monitor GitHub Actions CI workflow: Wait for .github/workflows/ci.yaml to complete, verify all steps pass
- [ ] T053 Verify CI build step: Check that build completes successfully with working-directory: src
- [ ] T054 Verify CI test step: Check that tests pass successfully with working-directory: src
- [ ] T055 Verify CI lint step: Check that linting passes
- [ ] T056 Test GoReleaser build: Run `cd src && goreleaser release --snapshot --clean` locally, verify binaries build
- [ ] T057 Test schema generation: Run `.github/scripts/generate-schema.sh` verify schemas generated correctly

**Checkpoint**: All CI checks pass on feature branch - ready for PR and merge

---

## Phase 7: External Consumer Validation (Optional Manual Test)

**Purpose**: Verify external packages can import ksail-go without changes

**⚠️ OPTIONAL**: This can be tested manually or skipped if time-constrained (import paths are unchanged by design)

- [ ] T058 Create test project: `mkdir /tmp/test-ksail-import && cd /tmp/test-ksail-import && go mod init test-import`
- [ ] T059 Import ksail package: `go get github.com/devantler-tech/ksail-go@001-move-all-source` verify resolution works
- [ ] T060 Create test file: Create test Go file importing `github.com/devantler-tech/ksail-go/pkg/...` verify autocomplete works
- [ ] T061 Build test project: `go build` in test project, verify it compiles successfully
- [ ] T062 Clean up test project: `rm -rf /tmp/test-ksail-import`

**Checkpoint**: External imports verified working (if performed) - ready for merge approval

---

## Phase 8: Post-Merge Validation (After Merge to Main)

**Purpose**: Verify production state after merge

> **⚠️ NOTE**: These tasks run AFTER the PR is merged to main branch

- [ ] T063 Checkout main branch: `git checkout main && git pull origin main` verify merge present
- [ ] T064 Build from main: `cd src && go build ./...` verify exit code 0
- [ ] T065 Test from main: `cd src && go test ./...` verify exit code 0
- [ ] T066 Monitor GitHub Actions: Verify ci.yaml workflow passes on main branch
- [ ] T067 Monitor release workflow: Verify release.yaml workflow (if triggered) passes
- [ ] T068 Verify schema generation in CI: Check that schema generation step passes in CI
- [ ] T069 Check code coverage reports: Verify codecov reports generated with correct src/ paths
- [ ] T070 Verify GoReleaser in CD: If cd.yaml triggers, verify GoReleaser builds successfully

**Checkpoint**: All post-merge checks pass - reorganization complete and verified in production

---

## Phase 9: Documentation and Communication

**Purpose**: Update documentation and notify team

- [ ] T071 [P] Update README.md: Add note about new src/ directory structure and build commands (if not already documented)
- [ ] T072 [P] Update CONTRIBUTING.md: Update build/test instructions to reference src/ directory
- [ ] T073 [P] Create migration guide: Document for team that workspace reload/IDE restart required after pulling changes
- [ ] T074 [P] Update project documentation: Update any other docs referencing repository structure
- [ ] T075 Notify team: Send announcement about reorganization with link to quickstart.md and migration notes

**Checkpoint**: Documentation updated, team notified - feature complete

---

## Rollback Plan (If Any Validation Fails)

**Purpose**: Define clear rollback procedure if failures occur

**Pre-Commit Rollback** (if failure during Phase 1-5):

- [ ] R001 Discard all changes: `git reset --hard HEAD` to undo all uncommitted changes
- [ ] R002 Clean untracked files: `git clean -fd` to remove src/ directory
- [ ] R003 Verify clean state: `git status` should show clean working directory
- [ ] R004 Document failure: Create issue documenting which task failed and why
- [ ] R005 Plan fix: Determine root cause and create new plan for corrected implementation

**Post-Commit Rollback** (if failure during Phase 6-7 before merge):

- [ ] R006 Reset to previous commit: `git reset --hard HEAD~1` to undo the reorganization commit
- [ ] R007 Force push if already pushed: `git push --force origin 001-move-all-source` (if needed)
- [ ] R008 Verify clean state: Build and test should work again
- [ ] R009 Document failure and plan fix as in R004-R005

**Post-Merge Rollback** (if failure during Phase 8 after merge):

- [ ] R010 Identify merge commit: `git log` to find the merge commit SHA
- [ ] R011 Revert merge commit: `git revert -m 1 <merge-commit-sha>` to undo the merge
- [ ] R012 Push revert: `git push origin main` to restore main branch
- [ ] R013 Verify main branch restored: Test build/tests work on main
- [ ] R014 Create fix issue: Document failure and create new PR with corrected implementation

---

## Dependencies & Execution Order

### Phase Dependencies (MUST execute sequentially)

1. **Phase 1: Pre-Move Validation** - No dependencies, start immediately
2. **Phase 2: File Reorganization** - Depends on Phase 1 completion
3. **Phase 3: Configuration Updates** - Depends on Phase 2 completion (all files must be moved first)
4. **Phase 4: Post-Move Validation** - Depends on Phase 3 completion (configs must be updated first)
5. **Phase 5: Commit and Pre-Merge** - Depends on Phase 4 completion (validation must pass first)
6. **Phase 6: Push and CI** - Depends on Phase 5 completion (commit must exist first)
7. **Phase 7: External Consumer** - Depends on Phase 6 completion (optional)
8. **Phase 8: Post-Merge** - Depends on PR being merged (triggers after merge)
9. **Phase 9: Documentation** - Can start after Phase 8 completion

### Within-Phase Parallelism

**Phase 3 (Configuration Updates)**: Tasks T019-T032 marked [P] can run in parallel as they edit different files

**Phase 9 (Documentation)**: Tasks T071-T075 marked [P] can run in parallel as they edit different files

### Critical Path

- Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 8 → Phase 9
- Phase 7 is optional and can be skipped without blocking merge
- Rollback tasks only execute if failures occur

---

## Implementation Strategy

### Single-Shot Execution (Atomic PR)

This feature MUST be implemented as a single atomic operation:

1. Complete Phase 1 (Pre-Move Validation) - Establish baseline ✅
2. Complete Phase 2 (File Reorganization) - Move all files at once ✅
3. Complete Phase 3 (Configuration Updates) - Update all configs ✅
4. Complete Phase 4 (Post-Move Validation) - Verify it works ✅
5. Complete Phase 5 (Commit) - Single commit with everything ✅
6. Complete Phase 6 (Push and CI) - Verify CI passes ✅
7. **STOP and REVIEW PR** - Get approval before merge
8. Merge PR to main
9. Complete Phase 8 (Post-Merge Validation) - Verify production ✅
10. Complete Phase 9 (Documentation) - Update docs and notify team ✅

**⚠️ DO NOT** split this across multiple PRs - the specification requires atomic migration

### Time Estimates

- Phase 1: 5 minutes (automated validation)
- Phase 2: 2 minutes (file moves)
- Phase 3: 15-20 minutes (configuration updates across multiple files)
- Phase 4: 5 minutes (automated validation)
- Phase 5: 10 minutes (commit + manual validation)
- Phase 6: 10-15 minutes (push + CI monitoring)
- Phase 7: 10 minutes (optional, if performed)
- Phase 8: 5 minutes (automated validation after merge)
- Phase 9: 15 minutes (documentation updates)

**Total Estimated Time**: 60-75 minutes for complete implementation

### Risk Mitigation

- **Pre-Move Validation**: Catches any existing issues before starting
- **Post-Move Validation**: Catches reorganization problems before commit
- **Pre-Merge Validation**: Catches issues before affecting main branch
- **Post-Merge Validation**: Catches issues in production environment
- **Rollback Plan**: Clear recovery path at every stage

---

## Success Criteria Mapping

| Success Criterion | Validated By |
|-------------------|--------------|
| SC-001: All source in src/ | T033-T041 (Phase 4 validation) |
| SC-002: Build/test success | T002, T003, T033, T034, T044, T064, T065 |
| SC-002a: Build time unchanged | T005, T040 (within 5% tolerance) |
| SC-003: Pipelines pass | T052-T055, T066-T070 |
| SC-004: IDE tasks work | T047-T050 |
| SC-005: Coverage reports | T069 |
| SC-006: Binary compilation | T036, T037, T056 |
| SC-007: Git history preserved | T046 |
| SC-008: Code quality tools | T004, T035, T045, T053, T055 |
| SC-009: Mock generation | T006, T038 |
| SC-010: Schema generation | T057, T068 |

---

## Notes

- This is an atomic operation - all tasks in Phases 1-6 must complete before merging
- Phase 7 is optional but recommended for confidence
- Phase 8 executes after merge to verify production state
- No parallel user stories - this is a single, atomic reorganization
- Rollback plan available at every stage for safety
- All validation contracts from contracts/validation-contracts.md are covered by these tasks
- Constitutional compliance verified: PATCH version, no new APIs, maintains simplicity
