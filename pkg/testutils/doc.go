// Package testutils provides testing utilities and helpers for KSail tests.
//
// This package contains test doubles, assertion helpers, file utilities,
// and other common testing infrastructure used across the KSail test suite.
//
// # Test Doubles
//
// The package provides test doubles for cluster provisioners:
//   - StubFactory: Test double for clusterprovisioner.Factory
//   - StubProvisioner: Test double for clusterprovisioner.ClusterProvisioner
//   - RecordingTimer: Test double for timer.Timer interface
//
// # Assertion Helpers
//
// Various assertion helpers for common test patterns:
//   - Error assertions: AssertErrWrappedContains, AssertErrContains, ExpectErrorContains
//   - Value assertions: ExpectEqual, ExpectNoError, ExpectNotNil, ExpectTrue
//   - String assertions: AssertStringContains, AssertStringContainsOneOf
//   - File assertions: AssertFileEquals
//
// # Kubernetes Test Helpers
//
// Utilities for creating fake Kubernetes clients and mock resources:
//   - Deployment clients: CreateReadyDeploymentClient, CreateUnreadyDeploymentClient
//   - DaemonSet clients: CreateReadyDaemonSetClient, CreateUnreadyDaemonSetClient
//   - Kubeconfig writers: WriteKubeconfig, WriteServerBackedKubeconfig
//
// # HTTP Test Helpers
//
// Utilities for creating test HTTP servers and mock responses:
//   - NewTestAPIServer: Creates test HTTP server with custom handler
//   - ServeDeployment, ServeDaemonSet: Mock Kubernetes API responses
//   - EncodeJSON: Encodes and writes JSON responses
//
// # Configuration and File Helpers
//
// Utilities for test configuration and file operations:
//   - WriteValidKsailConfig: Creates valid KSail configuration for tests
//   - CreateConfigManager: Creates config manager with test configuration
//   - NewCommand: Creates test Cobra command with output buffers
//
// # Marshalling Helpers
//
// Generic helpers for marshalling and unmarshalling in tests:
//   - MustMarshal, MustUnmarshal, MustUnmarshalString
package testutils
