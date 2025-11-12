// Package testutils provides testing utilities to aid error handling in tests.
//
// This package contains generic file-related test utilities, testing helpers shared
// across packages, and various utilities for error handling and test setup/teardown.
//
// # CNI Installer Test Utilities
//
// For CNI-specific test utilities, see the cnihelpers subpackage:
//
//	import "github.com/devantler-tech/ksail-go/pkg/testutils/cnihelpers"
//
// The cnihelpers package provides:
//   - Generic installer test framework using Go generics
//   - Helm mock expectation helpers
//   - HTTP test server and response helpers
//   - Common test errors and assertions
package testutils
