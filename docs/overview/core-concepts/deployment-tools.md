# Deployment Tools

Deployment tools control how manifests move from your workstation into the cluster. Choose a tool during initialization with `ksail cluster init --deployment-tool` or set `spec.gitOps.engine` in `ksail.yaml`.

## Kubectl

[Kubectl](https://kubernetes.io/docs/reference/kubectl/) is the default for KSail-Go. The CLI renders your `k8s/kustomization.yaml` and uses `kubectl apply -k` with pruning. When you run `ksail workload reconcile`, KSail-Go re-applies the manifests and watches status through `kubectl rollout status -k`. Use Kubectl when you want minimal moving parts or run ad-hoc experiments.

## Flux

[Flux](https://fluxcd.io/) installs during cluster creation and operates in GitOps mode. KSail-Go scaffolds an `OCIRepository` pointing at the local registry created by `ksail cluster create` and reconciles the same `k8s/kustomization.yaml`. This option is ideal when you want to mirror production GitOps flows locally.

## Argo CD (Roadmap)

> **Note:** Argo CD support is planned but not yet available in KSail-Go. Follow the [project roadmap](https://github.com/devantler-tech/ksail-go/milestones) for updates.

Argo CD will offer an alternative GitOps controller with application-centric workflows once implemented.
