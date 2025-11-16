// Package testutils provides testing utilities to aid error handling in tests.
//
// This package contains generic file-related test utilities, testing helpers shared
// across packages, and various utilities for error handling and test setup/teardown.
//
// # CNI Installer Test Utilities
//
// CNI-specific test utilities are available directly in this package:
//
//	import "github.com/devantler-tech/ksail-go/pkg/testutils"
//
// The package provides:
//   - Generic installer test framework using Go generics (cni_installer_helpers.go)
//   - Helm mock expectation helpers (cni_helm_helpers.go)
//   - HTTP test server and response helpers (cni_http_helpers.go)
//   - Common test errors and assertions (cni_helpers.go)
package testutils
