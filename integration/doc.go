// Package integration provides comprehensive integration tests for KSail CLI commands.
// These tests validate end-to-end command execution flows using stub implementations
// instead of mocks, ensuring proper integration between command layer and business logic.
//
// Each test file covers one command with all its possible execution paths:
//   - init_test.go: Tests ksail init command flows
//   - cluster_*_test.go: Tests all ksail cluster subcommand flows
//   - workload_*_test.go: Tests all ksail workload subcommand flows
//
// The tests use stub implementations from pkg/stubs to avoid external dependencies
// while maintaining realistic execution paths.
package integration