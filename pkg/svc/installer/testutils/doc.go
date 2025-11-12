// Package testutils provides comprehensive test utilities for CNI installer tests.
//
// This package consolidates shared test helpers, mock expectations, and utility
// functions used across Calico, Cilium, and other CNI installer test suites to
// standardize test patterns and eliminate code duplication.
//
// # Organization
//
// The package is organized into focused files:
//   - helpers.go: Common test errors, assertions, kubeconfig helpers, K8s client factories
//   - installer_helpers.go: Generic installer test framework using Go generics
//   - helm_helpers.go: Helm mock expectation helpers
//   - http_helpers.go: HTTP test server and response helpers
//
// # Usage
//
// Import the package in your CNI installer tests:
//
//	import installertestutils "github.com/devantler-tech/ksail-go/pkg/svc/installer/testutils"
//
// Use the generic test framework for consistent testing:
//
//	installertestutils.RunInstallerScenarios(t, scenarios, newInstaller)
//	installertestutils.TestSetWaitForReadinessFunc(t, newInstaller)
//
// Set up Helm mock expectations:
//
//	installertestutils.ExpectAddRepository(t, client, expect, nil)
//	installertestutils.ExpectInstallChart(t, client, expect, nil)
package testutils
