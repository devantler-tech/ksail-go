// Package k8sutil provides shared Kubernetes utilities for installers.
//
// This package is organized into focused modules:
//   - rest_config.go: Kubernetes REST configuration building
//   - polling.go: Generic polling utilities for readiness checks
//   - daemonset.go: DaemonSet-specific readiness checks
//   - deployment.go: Deployment-specific readiness checks
//   - multi_resource.go: Multi-resource readiness coordination
package k8sutil
