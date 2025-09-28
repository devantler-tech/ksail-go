# Phase 1 Data Model

## Command Entities

### WorkloadCommandGroup

- **Purpose**: Parent Cobra command mounted at `ksail workload` that scopes all workload-related operations.
- **Fields**:
  - `Use`: `"workload"`
  - `Short`: High-level summary ("Manage workload operations")
  - `Long`: Extended help describing reconciliation, apply, and install workflows (forward-looking).
  - `Aliases`: _None_ (explicit namespace per clarification).
  - `PersistentFlags`: _None for now_ (future iteration may add shared flags).
- **Relationships**: Owns three immediate subcommands (`WorkloadReconcile`, `WorkloadApply`, `WorkloadInstall`).

### WorkloadReconcileCommand

- **Purpose**: Placeholder command describing upcoming reconciliation behavior and emitting a "coming soon" notice.
- **Fields**:
  - `Use`: `"reconcile"`
  - `Short`: "Trigger workload reconciliation".
  - `Long`: Mentions future integration with GitOps/Flux reconciliation flows.
  - `RunE`: Prints the placeholder message via `notify.Infoln` and returns `nil`.
- **State & Validation**:
  - Exits with code 0.
  - When no cluster context exists, behavior remains the same (per clarification).

### WorkloadApplyCommand

- **Purpose**: Placeholder for applying local manifests, ensuring messaging consistency until implementation lands.
- **Fields**:
  - `Use`: `"apply"`
  - `Short`: "Apply workload manifests (coming soon)".
  - `Long`: Explains target behavior (wrapping `kubectl apply`) and notes current limitations.
  - `RunE`: Prints placeholder message via `notify.Infoln` and returns `nil`.
- **State & Validation**:
  - Exits with code 0 regardless of flags (none yet).

### WorkloadInstallCommand

- **Purpose**: Placeholder for Helm-based installs.
- **Fields**:
  - `Use`: `"install"`
  - `Short`: "Install workload Helm charts (coming soon)".
  - `Long`: Describes envisioned Helm integration path.
  - `RunE`: Prints placeholder message via `notify.Infoln` and returns `nil`.
- **State & Validation**:
  - Exits with code 0 regardless of inputs (none yet).

## Error Handling Entity

### ReconcileCommandMigration

- **Purpose**: Guidance output triggered when a user runs the removed top-level `ksail reconcile` command.
- **Behavior**:
  - Detect err string matching `unknown command "reconcile" for "ksail"` when executing the root command.
  - Append actionable message: `"Command 'reconcile' is now 'ksail workload reconcile'."`
  - Use `notify.Errorln` to preserve styled error output.
- **Validation**:
  - Unit tests ensure the message is appended and exit code is 1 (same as unknown command) to maintain CLI semantics.
