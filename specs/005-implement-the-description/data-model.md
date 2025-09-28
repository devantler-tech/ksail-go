# Phase 1 Data Model – KSail Cluster Up Command

## Entity: KSail Cluster Spec (`v1alpha1.Cluster`)

| Field | Type | Description | Notes |
|-------|------|-------------|-------|
| `spec.distribution` | `Distribution` enum (`Kind`, `K3d`, `EKS`) | Selects the provisioner implementation. | Must be validated before executing the command. |
| `spec.distributionConfig` | `string` | Path to distribution-specific config file (kind.yaml, k3d.yaml, eks.yaml). | Resolved relative to workspace using config-manager helpers. |
| `spec.connection.kubeconfig` | `string` | Kubeconfig path to merge/update after provisioning. | Defaults to `~/.kube/config`; expanded via `pathutils`. |
| `spec.connection.context` | `string` | Expected kube context name after provisioning. | Used to switch active context and to validate readiness. |
| `spec.connection.timeout` | `metav1.Duration` | Upper bound for provisioning + readiness waits. | Defaults to 5 minutes; flag-overridable. |
| `spec.options.eks.awsProfile` | `string` | Optional AWS profile for eksctl. | When empty rely on ambient credentials. |

## Entity: Distribution Configs

| Config | Loader | Key Fields consumed during `cluster up` |
|--------|--------|-----------------------------------------|
| Kind (`v1alpha4.Cluster`) | `kind.NewConfigManager` | `Name`, networking/node settings needed for provider. |
| K3d (`v1alpha5.SimpleConfig`) | `k3d.NewConfigManager` | `Name`, server/agent counts, kubeconfig options. |
| EKS (`eksctl.ClusterConfig`) | `eks.NewConfigManager` | `Metadata.Name`, `Metadata.Region`, node group definitions for scaling readiness. |

## Entity: Dependency Check Result

| Field | Type | Description |
|-------|------|-------------|
| `engineReady` | `bool` | True when Docker/Podman reachable (Kind/K3d only). |
| `engineName` | `string` | Friendly engine label from `containerengine.ContainerEngine`. |
| `awsCredentialsReady` | `bool` | True when AWS credential chain resolves (EKS only). |
| `messages` | `[]string` | Actionable error or success messages for CLI output. |

## Entity: Provisioning Outcome

| Field | Type | Description |
|-------|------|-------------|
| `clusterName` | `string` | Effective cluster name used by provider. |
| `distribution` | `string` | Distribution key for reporting (Kind/K3d/EKS). |
| `kubeconfigPath` | `string` | Resolved kubeconfig location after merge. |
| `context` | `string` | Active context set by the command. |
| `ready` | `bool` | Indicates readiness wait succeeded. |
| `duration` | `time.Duration` | Total time from provisioning start to readiness success. |

## Entity: Readiness Probe Configuration

| Field | Type | Description |
|-------|------|-------------|
| `timeout` | `time.Duration` | Upper bound for wait loop (default 5m). |
| `pollInterval` | `time.Duration` | Interval between Kubernetes readiness checks (default 5s). |
| `successCriteria` | `struct` | At minimum: all schedule-able nodes `Ready` and default namespace reachable. |
