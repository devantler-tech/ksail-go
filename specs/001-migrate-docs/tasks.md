# Tasks: Migrate KSail Documentation to KSail-Go

**Input**: Design documents from `/specs/001-migrate-docs/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md, contracts/
**Tests**: Markdownlint (`npx markdownlint-cli2 "docs/**/*.md"`), Lychee (`lychee docs`), GitHub Markdown preview spot-checks

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish tracking artifacts required across all stories

- [x] T001 Create migration checklist skeleton in `specs/001-migrate-docs/migration-checklist.md` listing source paths, target paths, owners, and status columns

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Capture shared mappings and validation steps before migrating content

- [x] T002 Build command translation table in `specs/001-migrate-docs/command-map.md` covering legacy KSail syntax to KSail-Go equivalents
- [x] T003 Document validation workflow in `specs/001-migrate-docs/validation-plan.md` describing markdownlint, lychee, and preview steps
- [x] T003a Execute `npx markdownlint-cli2 "docs/**/*.md"` and note pass/fail in `specs/001-migrate-docs/validation-log.md`
- [x] T003b Execute `lychee --config .lycheeignore docs` and note pass/fail in `specs/001-migrate-docs/validation-log.md`

---

## Phase 3: User Story 1 - Read Updated KSail-Go Docs (Priority: P1) 🎯 MVP

**Goal**: Deliver migrated documentation that reflects KSail-Go commands, structure, and assets so operators can rely on the repository docs alone

**Independent Test**: Reviewer opens the migrated configuration, overview, and use-case docs under `docs/` and confirms every command and link points to KSail-Go content without referencing the legacy ksail project

### Implementation for User Story 1

- [x] T004 [P] [US1] Migrate overview landing pages from the legacy KSail repository ([source](https://github.com/devantler-tech/ksail/tree/main/docs/overview/)) into `docs/overview/index.md`, `docs/overview/project-structure.md`, and `docs/overview/support-matrix.md` with KSail-Go terminology
- [x] T005 [P] [US1] Migrate core concepts set from the legacy KSail documentation (see [KSail repository](https://github.com/devantler-tech/ksail/tree/main/docs/overview/core-concepts/)) into `docs/overview/core-concepts/` while removing Jekyll front matter and updating command references
- [x] T006 [P] [US1] Migrate configuration guides (`cli-options.md`, `declarative-config.md`, `index.md`) into `docs/configuration/` and align file paths with KSail-Go configs
- [x] T007 [P] [US1] Migrate use-case guides (`local-development.md`, `learning-kubernetes.md`, `e2e-testing-in-cicd.md`, `index.md`) into `docs/use-cases/` with updated workflows
- [x] T008 [P] [US1] Copy required assets from the legacy KSail repository ([source](https://github.com/devantler-tech/ksail/tree/main/docs/images/)) into `docs/images/` and fix image references across migrated pages
- [x] T009 [US1] Update intra-doc navigation and cross-links across `docs/overview/`, `docs/configuration/`, and `docs/use-cases/` to ensure relative links resolve within ksail-go

---

## Phase 4: User Story 2 - Preview Documentation Locally (Priority: P2)

**Goal**: Ensure contributors can preview and validate the Markdown documentation without additional tooling setup

**Independent Test**: Follow the documented preview workflow on a clean checkout and confirm markdownlint and lychee complete successfully while the key docs render correctly

### Implementation for User Story 2

- [x] T010 [P] [US2] Create `docs/README.md` describing documentation layout and local preview commands (markdownlint, lychee, editor preview)
- [x] T011 [P] [US2] Update `README.md` documentation section to reference the migrated `docs/` content and outline the preview workflow
- [x] T012 [US2] Extend `CONTRIBUTING.md` with contributor instructions for running markdownlint and lychee before submitting doc changes

---

## Phase 5: User Story 3 - Understand Migration Coverage (Priority: P3)

**Goal**: Provide stakeholders with clear visibility into migrated sections, deferred items, and future publishing work

**Independent Test**: Review the migration summary and confirm it identifies each imported KSail document, notes any deferrals, and lists explicit follow-up items

### Implementation for User Story 3

- [x] T013 [P] [US3] Update `specs/001-migrate-docs/migration-checklist.md` with final status, command updates, and notes for each section
- [x] T014 [US3] Add a migration summary section to `specs/001-migrate-docs/spec.md` detailing migrated vs deferred content and rationale
- [x] T015 [US3] Capture publishing follow-ups and outstanding documentation tasks in `specs/001-migrate-docs/follow-ups.md`

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consolidate validation evidence and align supporting guides with the migrated docs

- [x] T016 Record results of markdownlint and lychee runs in `specs/001-migrate-docs/validation-log.md`, including timestamps and outcomes
- [x] T017 Refresh `specs/001-migrate-docs/quickstart.md` to reference the migrated docs and finalized validation workflow
- [x] T018 Conduct usability walkthrough: have a KSail-Go operator follow migrated quick start/configuration docs end-to-end and log reviewer, date, and outcome in `specs/001-migrate-docs/validation-log.md`

---

## Dependencies & Execution Order

- **Phase 1 → Phase 2**: Complete T001 before documenting command translations and validation workflows
- **Phase 2 → Phase 3**: Finish T002 and T003 to inform story-specific rewrites
- **User Story Order**: Execute User Story 1 (P1) before starting Stories 2 and 3; Story 2 depends on migrated docs existing, and Story 3 summarizes outcomes from prior work
- **Polish Phase**: T016 and T017 occur after all user stories finish to capture final validation evidence

## Parallel Execution Examples

- **User Story 1**: T004, T005, T006, T007, and T008 can run in parallel across different doc folders before T009 consolidates navigation updates
- **User Story 2**: T010 and T011 can proceed together while T012 finalizes contributor guidance
- **User Story 3**: T013 can begin once T009 completes, while T014 and T015 follow after T013 captures final statuses

## Implementation Strategy

- **MVP (Story 1)**: Complete Phases 1–3 to deliver migrated docs as the first releasable increment
- **Incremental Delivery**: Layer Story 2 updates for contributor workflows, then Story 3 summaries and follow-ups
- **Validation Cadence**: Run markdownlint and lychee after each story phase, logging final runs during Phase 6
