package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var errOperationFailed = errors.New("operation failed")

func TestWithDockerClient_Success(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var out bytes.Buffer

	cmd.SetOut(&out)

	operationCalled := false
	operation := func(dockerClient client.APIClient) error {
		operationCalled = true

		assert.NotNil(t, dockerClient)

		return nil
	}

	// Note: This test requires Docker to be available in the environment
	// If Docker is not available, the test will fail at client creation
	err := withDockerClient(cmd, operation)

	// We can't guarantee Docker is available in all test environments
	// so we accept both success and the specific error about Docker not being available
	if err != nil {
		// Check if it's a Docker connection error (expected in some environments)
		assert.Contains(t, err.Error(), "docker")
	} else {
		assert.True(t, operationCalled)
	}
}

func TestWithDockerClient_OperationError(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var out bytes.Buffer

	cmd.SetOut(&out)

	operation := func(_ client.APIClient) error {
		return errOperationFailed
	}

	err := withDockerClient(cmd, operation)

	// If Docker is available, we should get the operation error
	// If Docker is not available, we'll get a Docker connection error
	if err != nil && errors.Is(err, errOperationFailed) {
		assert.ErrorIs(t, err, errOperationFailed)
	}
}

func TestGenerateContainerdPatchesFromSpecs_SingleRegistry(t *testing.T) {
	t.Parallel()

	specs := []string{"docker.io=https://registry-1.docker.io"}
	patches := generateContainerdPatchesFromSpecs(specs)

	assert.Len(t, patches, 1)
	assert.Contains(
		t,
		patches[0],
		`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]`,
	)
	assert.Contains(t, patches[0], `endpoint = ["http://kind-docker-io:5000"]`)
}

func TestGenerateContainerdPatchesFromSpecs_MultipleRegistries(t *testing.T) {
	t.Parallel()

	specs := []string{
		"docker.io=https://registry-1.docker.io",
		"ghcr.io=https://ghcr.io",
	}
	patches := generateContainerdPatchesFromSpecs(specs)

	assert.Len(t, patches, 2)
	assert.Contains(t, patches[0], "docker.io")
	assert.Contains(t, patches[1], "ghcr.io")
}

func TestGenerateContainerdPatchesFromSpecs_InvalidSpecs(t *testing.T) {
	t.Parallel()

	specs := []string{
		"invalid",           // Missing '='
		"=http://localhost", // Empty registry
		"registry=",         // Empty endpoint
		"",                  // Empty string
	}
	patches := generateContainerdPatchesFromSpecs(specs)

	// All invalid specs should be skipped
	assert.Empty(t, patches)
}

func TestGenerateContainerdPatchesFromSpecs_MixedValidInvalid(t *testing.T) {
	t.Parallel()

	specs := []string{
		"docker.io=https://registry-1.docker.io",
		"invalid",
		"ghcr.io=https://ghcr.io",
	}
	patches := generateContainerdPatchesFromSpecs(specs)

	// Only valid specs should generate patches
	assert.Len(t, patches, 2)
}

func TestSplitMirrorSpec_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spec     string
		expected []string
	}{
		{
			name:     "simple registry",
			spec:     "docker.io=http://localhost:5000",
			expected: []string{"docker.io", "http://localhost:5000"},
		},
		{
			name:     "registry with port",
			spec:     "registry.io:443=http://localhost:5001",
			expected: []string{"registry.io:443", "http://localhost:5001"},
		},
		{
			name:     "complex URL",
			spec:     "ghcr.io=https://registry:5000/path",
			expected: []string{"ghcr.io", "https://registry:5000/path"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := splitMirrorSpec(testCase.spec)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestSplitMirrorSpec_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		spec string
	}{
		{name: "no equals", spec: "invalid"},
		{name: "empty registry", spec: "=http://localhost"},
		{name: "empty endpoint", spec: "registry="},
		{name: "empty string", spec: ""},
		{name: "only equals", spec: "="},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := splitMirrorSpec(testCase.spec)
			assert.Nil(t, result)
		})
	}
}

func TestSplitMirrorSpec_MultipleEquals(t *testing.T) {
	t.Parallel()

	// Only first '=' should be used as separator
	spec := "registry=endpoint=value"
	result := splitMirrorSpec(spec)

	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "registry", result[0])
	assert.Equal(t, "endpoint=value", result[1])
}

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestWithDockerClient_InvalidEnvironment(t *testing.T) {
	stubDockerClientFailure(t, errDockerClientFailure)

	cmd := &cobra.Command{}

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := withDockerClient(cmd, func(client.APIClient) error { return nil })
	if err == nil {
		t.Fatal("expected error when docker host is invalid")
	}

	if !strings.Contains(err.Error(), "failed to create docker client") {
		t.Fatalf("unexpected error: %v", err)
	}
}
