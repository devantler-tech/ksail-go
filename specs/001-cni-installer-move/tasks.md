# Tasks: CNI Installer Consolidation

**Input**: Design documents from `/specs/001-cni-installer-move/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅

**Tests**: Constitution Principle II requires maintaining existing test coverage through the relocation. All existing tests will be moved with source files.

**Organization**: Tasks are grouped by user story to enable independent validation of each story's success criteria.

## Format: `[ID] [P] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Go project structure:
- `pkg/` - Core business logic packages
- `cmd/` - CLI command implementations
- Tests co-located with source (`*_test.go`)

---

## Phase 1: Setup (Preparation & Baseline)

**Purpose**: Establish baseline and prepare for safe relocation

- [ ] T001 Capture baseline test results: `go test ./pkg/svc/installer/... > /tmp/pre-move-tests.txt`
- [ ] T002 [P] Capture baseline lint output: `golangci-lint run > /tmp/pre-move-lint.txt`
- [ ] T003 [P] Document current import paths: `grep -r "pkg/svc/installer" . --include="*.go" > /tmp/pre-move-imports.txt`
- [ ] T004 [P] Verify mockery configuration in `.mockery.yml` includes installer interfaces
- [ ] T005 Create target directory structure: `mkdir -p pkg/svc/installer/cni/{cilium,calico}`

---

## Phase 2: Foundational (No blocking prerequisites for this refactor)

**Purpose**: This is a pure refactor—no foundational infrastructure changes needed

**⚠️ Note**: This phase is empty because the refactor preserves all existing infrastructure unchanged. Proceed directly to user story implementation.

**Checkpoint**: No blocking work—user story implementation can begin immediately

---

## Phase 3: User Story 1 - Maintain installer functionality after package move (Priority: P1) 🎯 MVP

**Goal**: Relocate CNI packages while preserving all existing functionality, test coverage, and CLI output

**Independent Test**: Run `go test ./pkg/svc/installer/cni/...` and verify all tests pass; run `ksail cluster up` and verify Cilium installs successfully with unchanged output

### File Relocation for User Story 1

- [x] T006 [P] [US1] Move shared helpers: `git mv pkg/svc/installer/cni_helpers.go pkg/svc/installer/cni/base.go`
- [x] T007 [P] [US1] Move shared helpers tests: `git mv pkg/svc/installer/cni_helpers_test.go pkg/svc/installer/cni/base_test.go`
- [x] T008 [P] [US1] Move Cilium package: `git mv pkg/svc/installer/cilium pkg/svc/installer/cni/cilium`
- [x] T009 [P] [US1] Move Calico package: `git mv pkg/svc/installer/calico pkg/svc/installer/cni/calico`

### Package Declaration Updates for User Story 1

- [x] T010 [US1] Update package name in `pkg/svc/installer/cni/base.go` from `installer` to `cni`
- [x] T011 [US1] Update package name in `pkg/svc/installer/cni/base_test.go` from `installer` to `cni`
- [x] T012 [US1] Verify package name in `pkg/svc/installer/cni/cilium/installer.go` is correct; update if needed, and verify imports
- [x] T013 [US1] Verify package name and imports in `pkg/svc/installer/cni/calico/installer.go` (should be `calicoinstaller`)

### Import Path Updates for User Story 1

- [x] T014 [US1] Update imports in `pkg/svc/installer/cni/cilium/installer.go`: Changed `"github.com/devantler-tech/ksail-go/pkg/svc/installer"` to `"github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"`
- [x] T015 [US1] Update imports in `pkg/svc/installer/cni/calico/installer.go`: Changed `"github.com/devantler-tech/ksail-go/pkg/svc/installer"` to `"github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"`
- [ ] T016 [US1] Search and update all imports in `cmd/` directory (cross-platform):  
      ```sh
      grep -rl "pkg/svc/installer/cilium\|pkg/svc/installer/calico" cmd/ | xargs sed -i.bak 's|pkg/svc/installer/cilium|pkg/svc/installer/cni/cilium|g; s|pkg/svc/installer/calico|pkg/svc/installer/cni/calico|g'
      find cmd/ -name "*.bak" -delete
      ```
- [ ] T017 [US1] Search and update all imports in `pkg/` directory (excluding installer itself, cross-platform):  
      ```sh
      find pkg/ -name "*.go" -not -path "pkg/svc/installer/*" -exec grep -l "pkg/svc/installer/cilium\|pkg/svc/installer/calico" {} \; | xargs sed -i.bak 's|pkg/svc/installer/cilium|pkg/svc/installer/cni/cilium|g; s|pkg/svc/installer/calico|pkg/svc/installer/cni/calico|g'
      find pkg/ -name "*.bak" -delete
      ```

### Mock Regeneration for User Story 1

- [ ] T018 [US1] Regenerate all mocks with updated import paths: `mockery`
- [ ] T019 [US1] Verify mock imports reference new paths: `grep -r "pkg/svc/installer/cni" . --include="mock*.go"`

### Validation for User Story 1

- [ ] T020 [US1] Verify build succeeds: `go build ./...` (must succeed with zero errors)
- [ ] T021 [US1] Run CNI package tests: `go test -v ./pkg/svc/installer/cni/... | tee /tmp/post-move-cni-tests.txt` (must pass all tests in ≤90s per QC-003)
- [ ] T022 [US1] Run full test suite: `go test ./...` (must pass all tests)
- [ ] T023 [US1] Verify no old import paths remain: `grep -r "pkg/svc/installer/cilium\|pkg/svc/installer/calico" . --include="*.go" | grep -v "/cni/"` (must return zero results)
- [ ] T024 [US1] Run linter: `golangci-lint run` (must show zero new warnings)
- [ ] T025 [US1] Compare test timing: Verify CNI tests complete within 90 seconds (compare `/tmp/post-move-cni-tests.txt` timing)
- [ ] T026 [US1] Verify readiness callback wiring: `grep -r "waitForReadiness" pkg/svc/installer/cni/` (ensure callbacks still passed to base constructor)
- [ ] T027 [US1] Early dependency tidy: `go mod tidy` (catches dangling imports before integration tests)
- [ ] T028 [US1] Count test files pre vs post move (requires saved count from T001) and log equality: `echo PRE:$PRE_CNI_TEST_CT POST:$(find pkg/svc/installer/cni -name '*_test.go' | wc -l)`
- [ ] T029 [US1] Verify k8sutil unaffected: `grep -r "pkg/svc/installer/cni" pkg/svc/installer/k8sutil || echo 'OK: no unintended dependency'`
- [ ] T030 [US1] Grep for accidental new direct stdout logging: `grep -r "fmt.Println\|os.Stdout" pkg/svc/installer/cni/` (should be empty)

### Integration Smoke Test for User Story 1

- [ ] T031 [US1] Test Kind cluster with Cilium: `cd /tmp && mkdir -p ksail-test && cd ksail-test && ksail cluster init --distribution Kind && ksail up` (Cilium must install successfully)
- [ ] T032 [US1] Verify CLI output unchanged: Compare `ksail up` output with baseline (excluding timestamps)—notify/timer patterns must be identical
- [ ] T033 [US1] Clean up test cluster: `ksail down && cd .. && rm -rf ksail-test`

### Remediation Additions (US1 Observability & Calico)

- [ ] T034 [US1] Calico smoke test (if supported): repeat Kind cluster up using Calico configuration; verify identical notify/timer patterns
- [ ] T035 [US1] Capture Helm values before/after move for Cilium & Calico (serialize to /tmp/helm-values-{cilium,calico}.json) and diff to confirm unchanged
- [ ] T036 [US1] Security regression check: confirm no new privileged flags by grepping for `securityContext` changes in moved files
- [ ] T037 [US1] Lint diff against baseline: `golangci-lint run > /tmp/post-move-lint.txt && diff -u /tmp/pre-move-lint.txt /tmp/post-move-lint.txt || echo 'No new lint issues'`

**Checkpoint US1**: At this point, all CNI installer code relocated, builds succeed, tests pass, and cluster creation works identically to before the move

---

## Phase 4: User Story 2 - Simplify adding new CNIs (Priority: P2)

**Goal**: Add package documentation and contributor guidance making it clear where new CNIs should be placed

**Independent Test**: A contributor can scaffold a new CNI installer following `quickstart.md` and have it compile with proper imports

### Documentation for User Story 2

- [ ] T101 [P] [US2] Create package documentation: Create `pkg/svc/installer/cni/doc.go` with package overview, structure explanation, and "Adding a new CNI" guidance
- [ ] T102 [P] [US2] Update CONTRIBUTING.md: Add section referencing `pkg/svc/installer/cni/` as canonical location for CNI installers with link to `quickstart.md`
- [ ] T103 [P] [US2] Add inline godoc comments to `pkg/svc/installer/cni/base.go`: Document CNIInstallerBase struct and all exported functions with usage examples

### Validation for User Story 2

- [ ] T104 [US2] Verify package docs render correctly: `go doc github.com/devantler-tech/ksail-go/pkg/svc/installer/cni` (must show package overview and exported types)
- [ ] T105 [US2] Test contributor workflow: Create stub CNI following `quickstart.md` in `/tmp/test-cni-scaffold/` and verify it compiles with `go build ./...`
- [ ] T106 [US2] Verify documentation links: Check that CONTRIBUTING.md correctly links to `quickstart.md` and package paths

### Remediation Additions (US2 Documentation)

- [ ] T107 [US2] Inline comment audit: search for old paths in code comments `grep -r "pkg/svc/installer/calico\|pkg/svc/installer/cilium" . --include="*.go" | grep -v "/cni/"`
- [ ] T108 [US2] Ensure release notes (T056) include explicit link to `specs/001-cni-installer-move/quickstart.md`

**Checkpoint US2**: At this point, contributors have clear guidance on adding new CNIs, and package documentation is complete

---

## Phase 5: User Story 3 - Maintain upgrade path for existing imports (Priority: P3)

**Goal**: Ensure all internal references updated and no stale paths remain that could confuse maintainers

**Independent Test**: Run `go build ./...` and `golangci-lint run` with zero errors; search codebase for old paths returns no results

### Cleanup for User Story 3

- [ ] T201 [US3] Verify old directories removed: Check that `pkg/svc/installer/cilium/` and `pkg/svc/installer/calico/` no longer exist (should have been deleted by `git mv`)
- [ ] T202 [US3] Search for stale references in test files: `grep -r "pkg/svc/installer/cilium\|pkg/svc/installer/calico" . --include="*_test.go" | grep -v "/cni/"` (must return zero results)
- [ ] T203 [US3] Search for stale references in documentation: `grep -r "pkg/svc/installer/cilium\|pkg/svc/installer/calico" . --include="*.md" | grep -v "/cni/"` (must return zero results or update docs)
- [ ] T204 [US3] Search for stale references in YAML/config files: `grep -r "pkg/svc/installer/cilium\|pkg/svc/installer/calico" . --include="*.yaml" --include="*.yml"` (must return zero results)

### Final Validation for User Story 3

- [ ] T205 [US3] Run static analysis: `golangci-lint run` (must pass with zero errors/warnings, no "undefined: installer" or missing package errors)
- [ ] T206 [US3] Verify import consistency: `go mod tidy` (must complete without changes)
- [ ] T207 [US3] Run full build pipeline: `go build ./...` (must succeed in ≤5 minutes per performance budget)
- [ ] T208 [US3] Compare with baseline: Verify no new lint warnings introduced (compare `/tmp/post-move-lint.txt` with `/tmp/pre-move-lint.txt`)

### Remediation Additions (US3 Cleanup)

- [ ] T209 [US3] Stale mock path check: `grep -r "pkg/svc/installer/calico\|pkg/svc/installer/cilium" . --include="mock*.go" | grep -v "/cni/"`
- [ ] T210 [US3] JSON/YAML doc path search: extend T204 with `--include="*.json"` for completeness
- [ ] T211 [US3] Helm version check: `helm version --short | grep -E 'v3\.[8-9]|v3\.1[0-9]'`
- [ ] T212 [US3] Aggregate timing summary: write combined report `/tmp/cni-move-timings.txt` (include test & build durations)
- [ ] T213 [US3] Optional rollback simulation: introduce temporary bad import, confirm failure, revert commit (document in report)

**Checkpoint US3**: All old import paths removed, static analysis passes, build pipeline succeeds within time budget

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and evidence collection for success criteria

- [ ] T053 [P] Run CI validation: Push to feature branch `001-cni-installer-move` and verify GitHub Actions CI passes all checks
- [ ] T054 [P] Capture timing evidence: Document CNI test timing from CI output (must be ≤90s per SC-001)
- [ ] T055 [P] Capture build timing evidence: Document full repository build time from CI output (must be ≤5min per established baseline)
- [ ] T056 Update release notes: Add entry describing internal package restructuring for maintainers
- [ ] T057 Create comparison report: Summarize before/after package structure and confirm all success criteria met (SC-001 through SC-004)

### Remediation Additions (Polish)

- [ ] T058 [P] Add section to comparison report with Helm values diff & security regression evidence
- [ ] T059 [P] Append timing summary `/tmp/cni-move-timings.txt` into final comparison report
- [ ] T060 [P] Verify no direct stdout prints introduced (repeat T030 after all changes)
- [ ] T061 [P] Final grep that old paths absent across repo including mocks/docs/json: combined command
- [ ] T062 [P] Tag commit for rollback reference: `git tag cni-move-baseline`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies—can start immediately
- **Foundational (Phase 2)**: Empty—no blocking work
- **User Story 1 (Phase 3)**: Depends on Setup completion—MUST complete before US2/US3 (other stories depend on relocated files)
- **User Story 2 (Phase 4)**: Depends on US1 completion—documents the new structure
- **User Story 3 (Phase 5)**: Depends on US1 completion—validates the migration
- **Polish (Phase 6)**: Depends on US1, US2, US3 completion

### User Story Dependencies

- **User Story 1 (P1)**: BLOCKING—all other stories depend on this
  - Relocates actual files
  - Updates imports
  - Ensures builds work
- **User Story 2 (P2)**: Can run after US1 completes
  - Documents new structure (can run in parallel with US3)
- **User Story 3 (P3)**: Can run after US1 completes
  - Validates migration completeness (can run in parallel with US2)

### Within Each User Story

**User Story 1 (sequential within groups)**:

1. File relocation (T006-T009) → All [P] can run in parallel
2. Package declaration updates (T010-T013) → Sequential (one file at a time)
3. Import path updates (T014-T017) → Sequential (dependencies on previous steps)
4. Mock regeneration (T018-T019) → After imports updated
5. Validation (T020-T025) → After all changes applied
6. Integration test (T026-T028) → Final validation

**User Story 2**:

- All documentation tasks (T029-T031) are [P] and can run in parallel
- Validation tasks (T032-T034) run after documentation complete

**User Story 3**:

- All cleanup search tasks (T035-T038) can run in any order
- Final validation (T039-T042) runs after cleanup verified

### Parallel Opportunities

**Phase 1 Setup** (all can run in parallel):

- T002, T003, T004 (baseline capture)

**Phase 3 User Story 1**:

- T006-T009 (file moves - all different files)
- T020-T024 (validation checks - different commands)

**Phase 4 User Story 2**:

- T029-T031 (documentation creation - different files)

**Phase 5 User Story 3**:

- T035-T038 (cleanup verification - different search patterns)

**Phase 6 Polish**:

- T043-T045 (evidence collection - different sources)

**Between User Stories** (after US1 completes):

- User Story 2 and User Story 3 can proceed in parallel

---

## Parallel Example: User Story 1 File Relocation

```bash
# Launch all file moves in parallel (git mv operations):
Task T006: "git mv pkg/svc/installer/cni_helpers.go pkg/svc/installer/cni/base.go"
Task T007: "git mv pkg/svc/installer/cni_helpers_test.go pkg/svc/installer/cni/base_test.go"
Task T008: "git mv pkg/svc/installer/cilium pkg/svc/installer/cni/cilium"
Task T009: "git mv pkg/svc/installer/calico pkg/svc/installer/cni/calico"
```

## Parallel Example: User Story 1 Validation

```bash
# Launch validation checks in parallel:
Task T020: "go build ./..."
Task T021: "go test -v ./pkg/svc/installer/cni/..."
Task T023: "grep -r old-paths search"
Task T024: "golangci-lint run"
```

---

## Implementation Strategy

### Atomic Refactor (Recommended)

1. Complete Phase 1: Setup (capture baselines)
2. Complete Phase 3: User Story 1 in single commit
   - All file moves (T006-T009)
   - All package updates (T010-T013)
   - All import updates (T014-T017)
   - Mock regeneration (T018-T019)
   - Validation (T020-T025)
3. **CRITICAL**: Commit as atomic unit—do NOT commit partial state
4. Run integration smoke test (T026-T028)
5. Complete Phase 4: User Story 2 (documentation)
6. Complete Phase 5: User Story 3 (final cleanup validation)
7. Complete Phase 6: Polish (CI validation and evidence)

### Why Atomic?

- Import updates and file moves must happen together—partial state breaks builds
- Git revert of single commit restores working state if issues discovered
- Simplifies code review—one cohesive change vs. multiple dependent commits

### Rollback Strategy

If any validation task fails in Phase 3:

```bash
# Revert the entire relocation commit
git reset --hard HEAD~1

# Investigate failure
# Fix issue
# Retry from T006
```

---

## Success Criteria Mapping

Tasks mapped to specification success criteria:

- **SC-001** (Tests pass in ≤90s): T021, T025, T044
- **SC-002** (Build succeeds, no import errors): T020, T023, T039, T041, T043, T051
- **SC-003** (Docs reference new location): T030, T034, T036, T048
- **SC-004** (CI passes within baselines): T043, T045, T046, T049

---

## Notes

- All tasks include exact file paths or shell commands for clarity
- [P] tasks can run in parallel (different files, no dependencies)
- [US1]/[US2]/[US3] labels map tasks to user stories from spec.md
- User Story 1 is BLOCKING—must complete before US2/US3
- User Story 2 and 3 can run in parallel after US1
- Atomic commit strategy critical for this refactor—don't commit partial state
- Constitution Principle II satisfied: All existing tests preserved and relocated with source
- Constitution Principle IV satisfied: Timing evidence captured (T025, T044, T045)
