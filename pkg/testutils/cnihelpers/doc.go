// Package cnihelpers provides comprehensive test utilities for CNI installer tests.
//
// This package consolidates shared test helpers, mock expectations, and utility
// functions used across Calico, Cilium, and other CNI installer test suites to
// standardize test patterns and eliminate code duplication.
//
// # Organization
//
// The package is organized into focused files:
//   - cni_helpers.go: Common test errors, assertions, kubeconfig helpers, K8s client factories
//   - cni_installer_helpers.go: Generic installer test framework using Go generics
//   - cni_helm_helpers.go: Helm mock expectation helpers
//   - cni_http_helpers.go: HTTP test server and response helpers
//
// # Usage
//
// Import the package in your CNI installer tests:
//
//	import "github.com/devantler-tech/ksail-go/pkg/testutils/cnihelpers"
//
// Use the generic test framework for consistent testing:
//
//	cnihelpers.RunInstallerScenarios(t, scenarios, newInstaller)
//	cnihelpers.TestSetWaitForReadinessFunc(t, newInstaller)
//
// Set up Helm mock expectations:
//
//	cnihelpers.ExpectAddRepository(t, client, expect, nil)
//	cnihelpers.ExpectInstallChart(t, client, expect, nil)
package cnihelpers
