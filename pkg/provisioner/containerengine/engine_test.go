package containerengine_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test error variables to avoid dynamic error creation.
var (
	errDockerUnavailable       = errors.New("docker unavailable")
	errDockerNotReady          = errors.New("docker not ready")
	errPodmanUserUnavailable   = errors.New("podman user unavailable")
	errPodmanUserNotReady      = errors.New("podman user not ready")
	errPodmanSystemUnavailable = errors.New("podman system unavailable")
	errPodmanSystemNotReady    = errors.New("podman system not ready")
	errServerVersionFailed     = errors.New("server version failed")
)

// completePing returns a types.Ping struct for testing.
func completePing() types.Ping {
	return types.Ping{
		APIVersion: "1.41",
		OSType:     "linux",
	}
}

// createVersion creates a types.Version struct with specified platform name and version.
func createVersion(platformName, version string) types.Version {
	return types.Version{
		Platform:   struct{ Name string }{Name: platformName},
		Version:    version,
		APIVersion: "1.41",
	}
}

// dockerVersion returns a complete types.Version struct for Docker.
func dockerVersion() types.Version {
	return createVersion("Docker Engine - Community", "24.0.0")
}

// podmanVersion returns a complete types.Version struct for Podman.
func podmanVersion() types.Version {
	return createVersion("Podman Engine", "4.5.0")
}

// emptyVersion returns an empty types.Version for error testing.
func emptyVersion() types.Version {
	return createVersion("", "")
}

// assertAutoDetectionResult is a helper function to avoid code duplication
// when testing auto-detection behavior of NewContainerEngine.
func assertAutoDetectionResult(t *testing.T, engine *containerengine.ContainerEngine, err error) {
	t.Helper()

	if err != nil {
		assert.Equal(t, containerengine.ErrNoContainerEngine, err)
		assert.Nil(t, engine)
	} else {
		assert.NotNil(t, engine)
		assert.Contains(t, []string{"Docker", "Podman"}, engine.GetName())
	}
}

// assertSuccessfulEngineCreation consolidates the common pattern of asserting successful engine creation.
func assertSuccessfulEngineCreation(
	t *testing.T,
	engine *containerengine.ContainerEngine,
	err error,
	expectedName string,
	expectedClient client.APIClient,
) {
	t.Helper()

	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, expectedName, engine.GetName())
	assert.Equal(t, expectedClient, engine.Client)
}

// setupMockClientForEngineTest sets up a mock client for engine testing with server version expectations.
func setupMockClientForEngineTest(t *testing.T, testCase nameTestCase) *provisioner.MockAPIClient {
	t.Helper()

	mockClient := provisioner.NewMockAPIClient(t)
	if testCase.serverVersionErr != nil {
		mockClient.EXPECT().
			ServerVersion(context.Background()).
			Return(emptyVersion(), testCase.serverVersionErr)
	} else {
		mockClient.EXPECT().ServerVersion(context.Background()).Return(testCase.serverVersion, nil)
	}

	return mockClient
}

// assertDockerEngineSuccess sets up Docker client expectations and asserts successful creation.
func assertDockerEngineSuccess(
	t *testing.T,
	mockClient *provisioner.MockAPIClient,
	overrides map[string]containerengine.ClientCreator,
) {
	t.Helper()

	// Docker client succeeds and is ready
	mockClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	mockClient.EXPECT().ServerVersion(context.Background()).Return(dockerVersion(), nil)

	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)

	// Assert
	assertSuccessfulEngineCreation(t, engine, err, "Docker", mockClient)
}

func TestContainerEngineCheckReady(t *testing.T) {
	t.Parallel()

	tests := createContainerEngineTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockClient := provisioner.NewMockAPIClient(t)
			testCase.setupMock(mockClient)

			engine, err := containerengine.NewContainerEngine(mockClient)
			require.NoError(t, err)

			ready, err := engine.CheckReady(context.Background())

			assert.Equal(t, testCase.expectReady, ready)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createContainerEngineTestCases() []struct {
	name        string
	setupMock   func(*provisioner.MockAPIClient)
	expectReady bool
	expectError bool
} {
	return []struct {
		name        string
		setupMock   func(*provisioner.MockAPIClient)
		expectReady bool
		expectError bool
	}{
		{
			name: "container engine ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(completePing(), nil)
			},
			expectReady: true,
			expectError: false,
		},
		{
			name: "container engine not ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(completePing(), assert.AnError)
			},
			expectReady: false,
			expectError: true,
		},
	}
}

// nameTestCase represents a test case for engine name detection.
type nameTestCase struct {
	name             string
	serverVersion    types.Version
	serverVersionErr error
	expectedName     string
}

// createNameTestCases returns test cases for engine name detection.
func createNameTestCases() []nameTestCase {
	return []nameTestCase{
		// Basic detection cases
		{
			name:             "Docker engine detected",
			serverVersion:    dockerVersion(),
			serverVersionErr: nil,
			expectedName:     "Docker",
		},
		{
			name:             "Podman engine detected",
			serverVersion:    podmanVersion(),
			serverVersionErr: nil,
			expectedName:     "Podman",
		},
		// Edge cases
		{
			name:             "Version string contains podman",
			serverVersion:    createVersion("", "4.5.0-podman"),
			serverVersionErr: nil,
			expectedName:     "Podman",
		},
		{
			name:             "Version string without podman defaults to Docker",
			serverVersion:    createVersion("", "24.0.0"),
			serverVersionErr: nil,
			expectedName:     "Docker",
		},
		// Error cases
		{
			name:             "Empty platform and version returns Unknown",
			serverVersion:    emptyVersion(),
			serverVersionErr: nil,
			expectedName:     "Unknown",
		},
		{
			name:             "ServerVersion error returns Unknown",
			serverVersion:    emptyVersion(),
			serverVersionErr: errServerVersionFailed,
			expectedName:     "Unknown",
		},
	}
}

// runNameTestCase executes a single name test case.
func runNameTestCase(t *testing.T, testCase nameTestCase) {
	t.Helper()

	mockClient := setupMockClientForEngineTest(t, testCase)

	engine, err := containerengine.NewContainerEngine(mockClient)
	require.NoError(t, err)

	assert.Equal(t, testCase.expectedName, engine.GetName())
}

func TestContainerEngineName(t *testing.T) {
	t.Parallel()

	tests := createNameTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runNameTestCase(t, testCase)
		})
	}
}

func TestContainerEngineGetClient(t *testing.T) {
	t.Parallel()
	mockClient := provisioner.NewMockAPIClient(t)
	engine, err := containerengine.NewContainerEngine(mockClient)
	require.NoError(t, err)

	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngineWithInjectedClient(t *testing.T) {
	t.Parallel()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	mockClient.EXPECT().ServerVersion(context.Background()).Return(dockerVersion(), nil)

	// Act
	engine, err := containerengine.NewContainerEngine(mockClient)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngineWithNilClient(t *testing.T) {
	t.Parallel()

	// Act
	engine, err := containerengine.NewContainerEngine(nil)

	// Assert
	require.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "apiClient cannot be nil")
}

func TestNewContainerEngineWithAvailableEngine(t *testing.T) {
	t.Parallel()
	// Test with actual environment - this tests the real functionality
	// Use GetAutoDetectedClient for auto-detection since NewContainerEngine no longer does auto-detection
	engine, err := containerengine.GetAutoDetectedClient()
	assertAutoDetectionResult(t, engine, err)

	// Additional test: if we got a valid engine, test that it actually works
	if err == nil && engine != nil {
		ready, checkErr := engine.CheckReady(context.Background())
		if checkErr == nil {
			assert.True(t, ready)
		}
	}
}

func TestNewContainerEngineAPISignature(t *testing.T) {
	t.Parallel()

	t.Run("dependency injection mode", func(t *testing.T) {
		t.Parallel()
		mockClient := provisioner.NewMockAPIClient(t)
		mockClient.EXPECT().ServerVersion(context.Background()).Return(dockerVersion(), nil)

		// Test that we can inject a client and detect engine type
		engine, err := containerengine.NewContainerEngine(mockClient)

		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.Equal(t, "Docker", engine.GetName())
		assert.Equal(t, mockClient, engine.Client)
	})

	t.Run("auto-detection mode", func(t *testing.T) {
		t.Parallel()
		// Test auto-detection using GetAutoDetectedClient
		engine, err := containerengine.GetAutoDetectedClient()

		// Either we get an engine or an error, both are valid
		assertAutoDetectionResult(t, engine, err)
	})
}

func TestGetDockerClient(t *testing.T) {
	t.Parallel()

	// This test just verifies the function exists and returns a client or error
	// The actual Docker client creation depends on environment
	client, err := containerengine.GetDockerClient()

	// Either we get a client or an error, both are valid
	if err != nil {
		assert.Nil(t, client)
	} else {
		assert.NotNil(t, client)
	}
}

func TestGetPodmanUserClient(t *testing.T) {
	t.Parallel()

	// This test verifies the function exists and attempts to create a Podman user client
	client, err := containerengine.GetPodmanUserClient()

	// Either we get a client or an error, both are valid depending on environment
	if err != nil {
		assert.Nil(t, client)
	} else {
		assert.NotNil(t, client)
	}
}

func TestGetPodmanSystemClient(t *testing.T) {
	t.Parallel()

	// This test verifies the function exists and attempts to create a Podman system client
	client, err := containerengine.GetPodmanSystemClient()

	// Either we get a client or an error, both are valid depending on environment
	if err != nil {
		assert.Nil(t, client)
	} else {
		assert.NotNil(t, client)
	}
}

func TestGetAutoDetectedClient(t *testing.T) {
	t.Parallel()

	// Test the auto-detection function directly
	engine, err := containerengine.GetAutoDetectedClient()

	// Use the same assertion helper as other auto-detection tests
	assertAutoDetectionResult(t, engine, err)
}

// Test scenarios that might not be easily testable with real clients.
func TestGetAutoDetectedClientNoEngineAvailable(t *testing.T) {
	t.Parallel()

	// This test documents the expected behavior when no engine is available
	// Since we can't easily mock the real client creation in this environment,
	// we rely on system testing with the actual environment state

	// The GetAutoDetectedClient function should either:
	// 1. Return a valid engine if Docker/Podman is available and working
	// 2. Return ErrNoContainerEngine if no engines are available or working

	engine, err := containerengine.GetAutoDetectedClient()

	// This assertion covers both success and failure cases
	assertAutoDetectionResult(t, engine, err)
}

func TestGetAutoDetectedClientDockerSuccess(t *testing.T) {
	t.Parallel()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)

	// Create client creators using simple map
	overrides := map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return mockClient, nil
		},
		"podman-user": func() (client.APIClient, error) {
			return nil, errPodmanUserUnavailable
		},
		"podman-system": func() (client.APIClient, error) {
			return nil, errPodmanSystemUnavailable
		},
	}

	assertDockerEngineSuccess(t, mockClient, overrides)
}

// createTestOverrides creates client creator overrides for testing.
func createTestOverrides(dockerClient client.APIClient, dockerErr error,
	podmanUserClient client.APIClient, podmanUserErr error,
	podmanSystemClient client.APIClient, podmanSystemErr error,
) map[string]containerengine.ClientCreator {
	return map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return dockerClient, dockerErr
		},
		"podman-user": func() (client.APIClient, error) {
			return podmanUserClient, podmanUserErr
		},
		"podman-system": func() (client.APIClient, error) {
			return podmanSystemClient, podmanSystemErr
		},
	}
}

func TestGetAutoDetectedClientFallbackScenarios(t *testing.T) {
	t.Parallel()

	t.Run("DockerNotReady_PodmanUserSuccess", func(t *testing.T) {
		t.Parallel()

		// Test: Docker client creates but fails ping, Podman user works
		mockDockerClient := provisioner.NewMockAPIClient(t)
		mockPodmanClient := provisioner.NewMockAPIClient(t)

		overrides := createTestOverrides(
			mockDockerClient, nil,
			mockPodmanClient, nil,
			nil, errPodmanSystemUnavailable,
		)

		mockDockerClient.EXPECT().
			Ping(context.Background()).
			Return(completePing(), errDockerNotReady)
		mockPodmanClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
		mockPodmanClient.EXPECT().ServerVersion(context.Background()).Return(podmanVersion(), nil)

		engine, err := containerengine.GetAutoDetectedClient(overrides)

		assertSuccessfulEngineCreation(t, engine, err, "Podman", mockPodmanClient)
	})

	t.Run("DockerFails_PodmanUserNotReady_PodmanSystemSuccess", func(t *testing.T) {
		t.Parallel()

		// Test: Docker creation fails entirely, then user Podman succeeds creation but fails ping,
		// finally system Podman succeeds both creation and ping
		mockPodmanUserClient := provisioner.NewMockAPIClient(t)
		mockPodmanSystemClient := provisioner.NewMockAPIClient(t)

		// Different client setup pattern than above test
		overrides := map[string]containerengine.ClientCreator{
			"docker": func() (client.APIClient, error) {
				return nil, errDockerUnavailable
			},
			"podman-user": func() (client.APIClient, error) {
				return mockPodmanUserClient, nil
			},
			"podman-system": func() (client.APIClient, error) {
				return mockPodmanSystemClient, nil
			},
		}

		mockPodmanUserClient.EXPECT().
			Ping(context.Background()).
			Return(completePing(), errPodmanUserNotReady)
		mockPodmanSystemClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
		mockPodmanSystemClient.EXPECT().
			ServerVersion(context.Background()).
			Return(podmanVersion(), nil)

		engine, err := containerengine.GetAutoDetectedClient(overrides)

		assertSuccessfulEngineCreation(t, engine, err, "Podman", mockPodmanSystemClient)
	})
}

func TestGetAutoDetectedClientAllClientsFail(t *testing.T) {
	t.Parallel()

	overrides := createTestOverrides(
		nil, errDockerUnavailable,
		nil, errPodmanUserUnavailable,
		nil, errPodmanSystemUnavailable,
	)

	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)

	// Assert
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestGetAutoDetectedClientAllClientsCreateButNotReady(t *testing.T) {
	t.Parallel()

	// Arrange
	mockDockerClient := provisioner.NewMockAPIClient(t)
	mockPodmanUserClient := provisioner.NewMockAPIClient(t)
	mockPodmanSystemClient := provisioner.NewMockAPIClient(t)

	overrides := createTestOverrides(
		mockDockerClient, nil,
		mockPodmanUserClient, nil,
		mockPodmanSystemClient, nil,
	)

	// All clients create successfully but none are ready
	mockDockerClient.EXPECT().Ping(context.Background()).Return(completePing(), errDockerNotReady)
	mockPodmanUserClient.EXPECT().
		Ping(context.Background()).
		Return(completePing(), errPodmanUserNotReady)
	mockPodmanSystemClient.EXPECT().
		Ping(context.Background()).
		Return(completePing(), errPodmanSystemNotReady)

	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)

	// Assert
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestGetAutoDetectedClientPartialClientCreators(t *testing.T) {
	t.Parallel()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)

	// Test with only Docker creator - other clients will use defaults
	overrides := map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return mockClient, nil
		},
		// podman-user and podman-system will use default functions
	}

	assertDockerEngineSuccess(t, mockClient, overrides)
}

type edgeCaseTest struct {
	name             string
	serverVersion    types.Version
	serverVersionErr error
	expectedType     string
	expectError      bool
}

func getEdgeCasesTestData() []edgeCaseTest {
	return []edgeCaseTest{
		{
			"Platform name contains Docker",
			createVersion("Docker Engine - Community", "24.0.0"),
			nil,
			"Docker",
			false,
		},
		{
			"Platform name contains Podman",
			createVersion("Podman Engine", "4.5.0"),
			nil,
			"Podman",
			false,
		},
		{
			"Platform name empty, version contains podman",
			createVersion("", "4.5.0-podman"),
			nil,
			"Podman",
			false,
		},
		{
			"Platform name empty, version without podman defaults to Docker",
			createVersion("", "24.0.0"),
			nil,
			"Docker",
			false,
		},
		{"Both platform name and version empty", emptyVersion(), nil, "", true},
		{"ServerVersion API call fails", emptyVersion(), errServerVersionFailed, "", true},
	}
}

func runEdgeCaseTest(t *testing.T, testCase edgeCaseTest) {
	t.Helper()

	mockClient := setupMockClientForEngineTest(t, nameTestCase{
		name: testCase.name, serverVersion: testCase.serverVersion,
		serverVersionErr: testCase.serverVersionErr, expectedName: testCase.expectedType,
	})

	engine, err := containerengine.NewContainerEngine(mockClient)
	require.NoError(t, err)

	name := engine.GetName()
	if testCase.expectError {
		assert.Equal(t, "Unknown", name)
	} else {
		assert.Equal(t, testCase.expectedType, name)
	}
}

func TestDetectEngineTypeEdgeCases(t *testing.T) {
	t.Parallel()

	testCases := getEdgeCasesTestData()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runEdgeCaseTest(t, testCase)
		})
	}
}

func TestGetAutoDetectedClientWithEmptyOverrides(t *testing.T) {
	t.Parallel()

	// Test with empty map - should use default client creators
	emptyOverrides := map[string]containerengine.ClientCreator{}

	// Act - this will attempt to use real client creators
	engine, err := containerengine.GetAutoDetectedClient(emptyOverrides)

	// Assert - either success or expected error
	assertAutoDetectionResult(t, engine, err)
}

func TestGetAutoDetectedClientWithNilOverrides(t *testing.T) {
	t.Parallel()

	// Test with nil overrides - should use default client creators
	var nilOverrides map[string]containerengine.ClientCreator

	// Act - this will attempt to use real client creators
	engine, err := containerengine.GetAutoDetectedClient(nilOverrides)

	// Assert - either success or expected error
	assertAutoDetectionResult(t, engine, err)
}

func TestContainsHelper(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name, platformName, version, expectedName string
	}{
		{"Docker exact match in platform", "Docker", "1.0.0", "Docker"},
		{"Docker case insensitive in platform", "DOCKER ENGINE", "1.0.0", "Docker"},
		{"Docker substring in platform", "Docker Engine - Community", "1.0.0", "Docker"},
		{"Podman in platform", "Podman Engine", "4.5.0", "Podman"},
		{"Empty platform, podman in version", "", "4.5.0-podman", "Podman"},
		{"Empty platform, no podman in version defaults to Docker", "", "24.0.0", "Docker"},
		{"No match anywhere", "Something else", "1.0.0", "Docker"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockClient := provisioner.NewMockAPIClient(t)
			version := createVersion(testCase.platformName, testCase.version)
			mockClient.EXPECT().ServerVersion(context.Background()).Return(version, nil)

			engine, err := containerengine.NewContainerEngine(mockClient)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedName, engine.GetName())
		})
	}
}

func TestTryCreateEngineNewContainerEngineFailure(t *testing.T) {
	t.Parallel()

	// This test covers the edge case where a client creator returns a nil client
	// which should cause NewContainerEngine to fail with ErrAPIClientNil
	creator := func() (client.APIClient, error) {
		// Return nil client to trigger ErrAPIClientNil in NewContainerEngine
		var nilClient client.APIClient

		return nilClient, nil
	}

	engine, err := containerengine.GetAutoDetectedClient(map[string]containerengine.ClientCreator{
		"docker":        creator,
		"podman-user":   creator,
		"podman-system": creator,
	})

	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestClientCreationAllScenarios(t *testing.T) {
	t.Parallel()

	// Test all client creation functions
	clientFunctions := map[string]func() (client.APIClient, error){
		"Docker":        containerengine.GetDockerClient,
		"Podman user":   containerengine.GetPodmanUserClient,
		"Podman system": containerengine.GetPodmanSystemClient,
	}

	for clientName, clientFunc := range clientFunctions {
		t.Run(fmt.Sprintf("Get%sClient handles creation properly", clientName), func(t *testing.T) {
			t.Parallel()

			// This tests that the function doesn't panic and returns either a client or error
			client, err := clientFunc()

			// Both success and failure are valid outcomes depending on environment
			if err != nil {
				assert.Nil(t, client)
				assert.Contains(
					t,
					err.Error(),
					fmt.Sprintf("failed to create %s client", clientName),
				)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}
