# KSail Legacy → KSail-Go Command Translation

| Legacy Command | KSail-Go Command | Notes |
|----------------|------------------|-------|
| `ksail init` | `ksail cluster init` | Initialization now lives under the `cluster` namespace with additional flags for output directory and mirror registries. |
| `ksail up` | `ksail cluster create` (+ `ksail cluster start` when resuming) | Cluster creation and startup are separate subcommands; creating a cluster is idempotent and start handles previously created clusters. |
| `ksail down` | `ksail cluster delete` | Deletes the current cluster context and related resources. |
| `ksail start` | `ksail cluster start` | Explicitly starts a stopped cluster. |
| `ksail stop` | `ksail cluster stop` | Stops a running cluster without deleting artifacts. |
| `ksail status` | `ksail cluster info` | Provides cluster status and key information. |
| `ksail list` | `ksail cluster list` | Lists discovered clusters across providers. |
| `ksail connect` | `ksail cluster connect` | Opens an interactive session (e.g., k9s) into the cluster. |
| `ksail update` | `ksail workload reconcile` | Replaces the generic update verb with workload reconciliation against the cluster. |
| `ksail validate` | `ksail workload wait` (pending success criteria) | Waiting on workloads provides validation feedback for applied manifests. |
| `ksail gen ...` | `ksail workload gen ...` | Generator commands are grouped beneath `workload gen` with resource-specific subcommands. |
| `ksail secrets encrypt` | `ksail cipher encrypt` | Secrets management uses the `cipher` namespace for SOPS operations. |
| `ksail secrets decrypt` | `ksail cipher decrypt` | |
| `ksail secrets edit` | `ksail cipher edit` | |
| `ksail secrets add` / `rm` / `list` / `import` / `export` | `ksail cipher ...` (future roadmap) | Cipher currently focuses on encrypt/edit/decrypt; other subcommands remain roadmap items if needed. |
| `ksail gen flux ...` | `ksail workload gen helm-release` / `helm-repository` / etc. | Helm-related generators surface through the workload generator commands. |
| `ksail gen config ksail` | `ksail cluster init --config` (scaffold) | Cluster scaffolding now handled during `cluster init` with config flags. |
| `ksail gen native <resource>` | `ksail workload gen <resource>` | Mirrors kubectl dry-run generation for native resources. |
