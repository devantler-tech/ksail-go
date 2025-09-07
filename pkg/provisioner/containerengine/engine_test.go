package containerengine_test

import (
	"context"
	"errors"
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
	errDockerNotReady         = errors.New("docker not ready")
	errPodmanUserUnavailable  = errors.New("podman user unavailable")
	errPodmanUserNotReady     = errors.New("podman user not ready")
	errPodmanSystemUnavailable = errors.New("podman system unavailable")
	errPodmanSystemNotReady   = errors.New("podman system not ready")
	errServerVersionFailed    = errors.New("server version failed")
	errMockDockerCreation     = errors.New("mock docker creation error")
	errMockPodmanUserCreation = errors.New("mock podman user creation error")
	errMockPodmanSystemCreation = errors.New("mock podman system creation error")
)

// completePing returns a complete types.Ping struct to satisfy exhaustruct linter.
func completePing() types.Ping {
	return types.Ping{
		APIVersion:     "1.41",
		OSType:         "linux",
		Experimental:   false,
		BuilderVersion: "1",
		SwarmStatus:    nil,
	}
}

// dockerVersion returns a complete types.Version struct for Docker to satisfy exhaustruct linter.
func dockerVersion() types.Version {
	return types.Version{
		Platform: struct{ Name string }{Name: "Docker Engine - Community"},
		Components: nil,
		Version:  "24.0.0",
		APIVersion: "1.41", 
		MinAPIVersion: "1.12",
		GitCommit: "abcd123",
		GoVersion: "go1.19",
		Os: "linux",
		Arch: "amd64",
		KernelVersion: "5.4.0",
		Experimental: false,
		BuildTime: "2023-01-01T00:00:00.000000000Z",
	}
}

// podmanVersion returns a complete types.Version struct for Podman to satisfy exhaustruct linter.
func podmanVersion() types.Version {
	return types.Version{
		Platform: struct{ Name string }{Name: "Podman Engine"},
		Components: nil,
		Version:  "4.5.0",
		APIVersion: "1.41",
		MinAPIVersion: "1.12", 
		GitCommit: "efgh456",
		GoVersion: "go1.19",
		Os: "linux",
		Arch: "amd64",
		KernelVersion: "5.4.0",
		Experimental: false,
		BuildTime: "2023-01-01T00:00:00.000000000Z",
	}
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

func TestContainerEngine_CheckReady(t *testing.T) {
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
	// Common ping response structure to avoid duplication
	pingResponse := types.Ping{
		APIVersion:     "",
		OSType:         "",
		Experimental:   false,
		BuilderVersion: "",
		SwarmStatus:    nil,
	}

	return []struct {
		name        string
		setupMock   func(*provisioner.MockAPIClient)
		expectReady bool
		expectError bool
	}{
		{
			name: "container engine ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(pingResponse, nil)
			},
			expectReady: true,
			expectError: false,
		},
		{
			name: "container engine not ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(pingResponse, assert.AnError)
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
		{
			name: "Version string contains podman",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version:  "4.5.0-podman",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedName: "Podman",
		},
		{
			name: "Version string without podman defaults to Docker",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version:  "24.0.0",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedName: "Docker",
		},
		{
			name: "Empty platform and version returns Unknown",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version:  "",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedName: "Unknown",
		},
		{
			name: "ServerVersion error returns Unknown",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version: "",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: errServerVersionFailed,
			expectedName:     "Unknown",
		},
	}
}

// runNameTestCase executes a single name test case.
func runNameTestCase(t *testing.T, testCase nameTestCase) {
	t.Helper()
	
	mockClient := provisioner.NewMockAPIClient(t)
	if testCase.serverVersionErr != nil {
		mockClient.EXPECT().ServerVersion(context.Background()).Return(types.Version{
			Platform: struct{ Name string }{Name: ""},
			Components: nil,
			Version: "",
			APIVersion: "",
			MinAPIVersion: "",
			GitCommit: "",
			GoVersion: "",
			Os: "",
			Arch: "",
			KernelVersion: "",
			Experimental: false,
			BuildTime: "",
		}, testCase.serverVersionErr)
	} else {
		mockClient.EXPECT().ServerVersion(context.Background()).Return(testCase.serverVersion, nil)
	}
	
	engine, err := containerengine.NewContainerEngine(mockClient)
	require.NoError(t, err)

	assert.Equal(t, testCase.expectedName, engine.GetName())
}

func TestContainerEngine_Name(t *testing.T) {
	t.Parallel()

	tests := createNameTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runNameTestCase(t, testCase)
		})
	}
}

func TestContainerEngine_GetClient(t *testing.T) {
	t.Parallel()
	mockClient := provisioner.NewMockAPIClient(t)
	engine, err := containerengine.NewContainerEngine(mockClient)
	require.NoError(t, err)

	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngine_WithInjectedClient(t *testing.T) {
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

func TestNewContainerEngine_WithNilClient(t *testing.T) {
	t.Parallel()
	
	// Act
	engine, err := containerengine.NewContainerEngine(nil)
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "apiClient cannot be nil")
}

func TestNewContainerEngine_WithAvailableEngine(t *testing.T) {
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

func TestNewContainerEngine_APISignature(t *testing.T) {
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
func TestGetAutoDetectedClient_NoEngineAvailable(t *testing.T) {
	t.Parallel()
	
	// This test documents the expected behavior when no engine is available
	// Since we can't easily mock the real client creation in this environment,
	// we rely on integration testing with the actual environment state
	
	// The GetAutoDetectedClient function should either:
	// 1. Return a valid engine if Docker/Podman is available and working
	// 2. Return ErrNoContainerEngine if no engines are available or working
	
	engine, err := containerengine.GetAutoDetectedClient()
	
	// This assertion covers both success and failure cases
	if err != nil {
		assert.Equal(t, containerengine.ErrNoContainerEngine, err)
		assert.Nil(t, engine)
	} else {
		assert.NotNil(t, engine)
		assert.Contains(t, []string{"Docker", "Podman"}, engine.GetName())
		
		// If we got an engine, it should be ready
		ready, readyErr := engine.CheckReady(context.Background())
		if readyErr == nil {
			assert.True(t, ready)
		}
	}
}

func TestGetAutoDetectedClient_DockerSuccess(t *testing.T) {
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
	
	// Docker client succeeds and is ready
	mockClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	mockClient.EXPECT().ServerVersion(context.Background()).Return(dockerVersion(), nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

// createTestOverrides creates client creator overrides for testing.
type clientSetup struct {
	dockerClient        client.APIClient
	dockerErr           error
	podmanUserClient    client.APIClient
	podmanUserErr       error
	podmanSystemClient  client.APIClient
	podmanSystemErr     error
}

func createTestOverrides(setup clientSetup) map[string]containerengine.ClientCreator {
	return map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return setup.dockerClient, setup.dockerErr
		},
		"podman-user": func() (client.APIClient, error) {
			return setup.podmanUserClient, setup.podmanUserErr
		},
		"podman-system": func() (client.APIClient, error) {
			return setup.podmanSystemClient, setup.podmanSystemErr
		},
	}
}

func TestGetAutoDetectedClient_DockerNotReady_PodmanUserSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockDockerClient := provisioner.NewMockAPIClient(t)
	mockPodmanClient := provisioner.NewMockAPIClient(t)
	
	overrides := createTestOverrides(clientSetup{
		dockerClient:        mockDockerClient,
		dockerErr:           nil,
		podmanUserClient:    mockPodmanClient,
		podmanUserErr:       nil,
		podmanSystemClient:  nil,
		podmanSystemErr:     errPodmanSystemUnavailable,
	})
	
	// Docker client succeeds but is not ready
	mockDockerClient.EXPECT().Ping(context.Background()).Return(completePing(), errDockerNotReady)
	
	// Podman user client succeeds and is ready  
	mockPodmanClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	mockPodmanClient.EXPECT().ServerVersion(context.Background()).Return(podmanVersion(), nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, mockPodmanClient, engine.Client)
}

func TestGetAutoDetectedClient_DockerFails_PodmanUserNotReady_PodmanSystemSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockPodmanUserClient := provisioner.NewMockAPIClient(t)
	mockPodmanSystemClient := provisioner.NewMockAPIClient(t)
	
	overrides := createTestOverrides(clientSetup{
		dockerClient:        nil,
		dockerErr:           errDockerUnavailable,
		podmanUserClient:    mockPodmanUserClient,
		podmanUserErr:       nil,
		podmanSystemClient:  mockPodmanSystemClient,
		podmanSystemErr:     nil,
	})
	
	// Podman user client succeeds but is not ready
	mockPodmanUserClient.EXPECT().Ping(context.Background()).Return(completePing(), errPodmanUserNotReady)
	
	// Podman system client succeeds and is ready
	mockPodmanSystemClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	mockPodmanSystemClient.EXPECT().ServerVersion(context.Background()).Return(podmanVersion(), nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, mockPodmanSystemClient, engine.Client)
}

func TestGetAutoDetectedClient_AllClientsFail(t *testing.T) {
	t.Parallel()
	
	overrides := createTestOverrides(clientSetup{
		dockerClient:        nil,
		dockerErr:           errDockerUnavailable,
		podmanUserClient:    nil,
		podmanUserErr:       errPodmanUserUnavailable,
		podmanSystemClient:  nil,
		podmanSystemErr:     errPodmanSystemUnavailable,
	})
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestGetAutoDetectedClient_AllClientsCreateButNotReady(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockDockerClient := provisioner.NewMockAPIClient(t)
	mockPodmanUserClient := provisioner.NewMockAPIClient(t)
	mockPodmanSystemClient := provisioner.NewMockAPIClient(t)
	
	overrides := createTestOverrides(clientSetup{
		dockerClient:        mockDockerClient,
		dockerErr:           nil,
		podmanUserClient:    mockPodmanUserClient,
		podmanUserErr:       nil,
		podmanSystemClient:  mockPodmanSystemClient,
		podmanSystemErr:     nil,
	})
	
	// All clients create successfully but none are ready
	mockDockerClient.EXPECT().Ping(context.Background()).Return(completePing(), errDockerNotReady)
	mockPodmanUserClient.EXPECT().Ping(context.Background()).Return(completePing(), errPodmanUserNotReady)
	mockPodmanSystemClient.EXPECT().Ping(context.Background()).Return(completePing(), errPodmanSystemNotReady)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestGetAutoDetectedClient_PartialClientCreators(t *testing.T) {
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
	
	// Docker client succeeds and is ready
	mockClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	mockClient.EXPECT().ServerVersion(context.Background()).Return(dockerVersion(), nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

func TestDetectEngineType_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		serverVersion    types.Version
		serverVersionErr error
		expectedType     string
		expectError      bool
	}{
		{
			name:             "Platform name contains Docker",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: "Docker Engine - Community"},
				Components: nil,
				Version:  "24.0.0",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedType: "Docker",
			expectError:  false,
		},
		{
			name: "Platform name contains Podman",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: "Podman Engine"},
				Components: nil,
				Version:  "4.5.0",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedType: "Podman",
			expectError:  false,
		},
		{
			name: "Platform name empty, version contains podman",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version:  "4.5.0-podman",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedType: "Podman",
			expectError:  false,
		},
		{
			name: "Platform name empty, version without podman defaults to Docker",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version:  "24.0.0",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedType: "Docker",
			expectError:  false,
		},
		{
			name: "Both platform name and version empty",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version:  "",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: nil,
			expectedType:     "",
			expectError:      true,
		},
		{
			name: "ServerVersion API call fails",
			serverVersion: types.Version{
				Platform: struct{ Name string }{Name: ""},
				Components: nil,
				Version: "",
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			},
			serverVersionErr: errServerVersionFailed,
			expectedType:     "",
			expectError:      true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			
			mockClient := provisioner.NewMockAPIClient(t)
			if testCase.serverVersionErr != nil {
				mockClient.EXPECT().ServerVersion(context.Background()).Return(types.Version{
					Platform: struct{ Name string }{Name: ""},
					Components: nil,
					Version: "",
					APIVersion: "",
					MinAPIVersion: "",
					GitCommit: "",
					GoVersion: "",
					Os: "",
					Arch: "",
					KernelVersion: "",
					Experimental: false,
					BuildTime: "",
				}, testCase.serverVersionErr)
			} else {
				mockClient.EXPECT().ServerVersion(context.Background()).Return(testCase.serverVersion, nil)
			}
			
			engine, err := containerengine.NewContainerEngine(mockClient)
			require.NoError(t, err)

			// Use reflection to call the private detectEngineType method
			// Since we can't call it directly, test through GetName which uses it
			name := engine.GetName()
			
			if testCase.expectError {
				assert.Equal(t, "Unknown", name)
			} else {
				assert.Equal(t, testCase.expectedType, name)
			}
		})
	}
}

func TestGetAutoDetectedClient_WithEmptyOverrides(t *testing.T) {
	t.Parallel()
	
	// Test with empty map - should use default client creators
	emptyOverrides := map[string]containerengine.ClientCreator{}
	
	// Act - this will attempt to use real client creators
	engine, err := containerengine.GetAutoDetectedClient(emptyOverrides)
	
	// Assert - either success or expected error
	if err != nil {
		assert.Equal(t, containerengine.ErrNoContainerEngine, err)
		assert.Nil(t, engine)
	} else {
		assert.NotNil(t, engine)
		assert.Contains(t, []string{"Docker", "Podman"}, engine.GetName())
	}
}

func TestGetAutoDetectedClient_WithNilOverrides(t *testing.T) {
	t.Parallel()
	
	// Test with nil overrides - should use default client creators
	var nilOverrides map[string]containerengine.ClientCreator
	
	// Act - this will attempt to use real client creators
	engine, err := containerengine.GetAutoDetectedClient(nilOverrides)
	
	// Assert - either success or expected error
	if err != nil {
		assert.Equal(t, containerengine.ErrNoContainerEngine, err)
		assert.Nil(t, engine)
	} else {
		assert.NotNil(t, engine)
		assert.Contains(t, []string{"Docker", "Podman"}, engine.GetName())
	}
}

func TestContainsHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		platformName  string
		version       string
		expectedName  string
	}{
		{
			name:         "Docker exact match in platform",
			platformName: "Docker",
			version:      "1.0.0",
			expectedName: "Docker",
		},
		{
			name:         "Docker case insensitive in platform",
			platformName: "DOCKER ENGINE",
			version:      "1.0.0",
			expectedName: "Docker",
		},
		{
			name:         "Docker substring in platform",
			platformName: "Docker Engine - Community",
			version:      "1.0.0",
			expectedName: "Docker",
		},
		{
			name:         "Podman in platform",
			platformName: "Podman Engine",
			version:      "4.5.0",
			expectedName: "Podman",
		},
		{
			name:         "Empty platform, podman in version",
			platformName: "",
			version:      "4.5.0-podman",
			expectedName: "Podman",
		},
		{
			name:         "Empty platform, no podman in version defaults to Docker",
			platformName: "",
			version:      "24.0.0",
			expectedName: "Docker",
		},
		{
			name:         "No match anywhere",
			platformName: "Something else",
			version:      "1.0.0",
			expectedName: "Docker", // Defaults to Docker
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			
			mockClient := provisioner.NewMockAPIClient(t)
			
			version := types.Version{
				Platform: struct{ Name string }{Name: testCase.platformName},
				Components: nil,
				Version:  testCase.version,
				APIVersion: "",
				MinAPIVersion: "",
				GitCommit: "",
				GoVersion: "",
				Os: "",
				Arch: "",
				KernelVersion: "",
				Experimental: false,
				BuildTime: "",
			}
			
			mockClient.EXPECT().ServerVersion(context.Background()).Return(version, nil)
			
			engine, err := containerengine.NewContainerEngine(mockClient)
			require.NoError(t, err)
			
			assert.Equal(t, testCase.expectedName, engine.GetName())
		})
	}
}

func TestTryCreateEngine_NewContainerEngineFailure(t *testing.T) {
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

func TestClientCreation_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("GetDockerClient handles creation properly", func(t *testing.T) {
		t.Parallel()
		
		// This tests that the function doesn't panic and returns either a client or error
		client, err := containerengine.GetDockerClient()
		
		// Both success and failure are valid outcomes depending on environment
		if err != nil {
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), "failed to create Docker client")
		} else {
			assert.NotNil(t, client)
		}
	})

	t.Run("GetPodmanUserClient handles creation properly", func(t *testing.T) {
		t.Parallel()
		
		// This tests that the function doesn't panic and returns either a client or error
		client, err := containerengine.GetPodmanUserClient()
		
		// Both success and failure are valid outcomes depending on environment
		if err != nil {
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), "failed to create Podman user client")
		} else {
			assert.NotNil(t, client)
		}
	})

	t.Run("GetPodmanSystemClient handles creation properly", func(t *testing.T) {
		t.Parallel()
		
		// This tests that the function doesn't panic and returns either a client or error
		client, err := containerengine.GetPodmanSystemClient()
		
		// Both success and failure are valid outcomes depending on environment
		if err != nil {
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), "failed to create Podman system client")
		} else {
			assert.NotNil(t, client)
		}
	})
}

func TestClientCreation_SuccessPaths(t *testing.T) {
	t.Parallel()

	t.Run("GetDockerClient success path", func(t *testing.T) {
		t.Parallel()
		
		// Create a mock client for testing the success path
		mockClient := provisioner.NewMockAPIClient(t)
		
		// Temporarily replace the client creator for this test
		originalCreator := containerengine.DefaultDockerClientCreator
		containerengine.DefaultDockerClientCreator = func() (client.APIClient, error) {
			return mockClient, nil
		}

		defer func() {
			containerengine.DefaultDockerClientCreator = originalCreator
		}()
		
		// Test the success path
		client, err := containerengine.GetDockerClient()
		
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, mockClient, client)
	})

	t.Run("GetPodmanUserClient success path", func(t *testing.T) {
		t.Parallel()
		
		// Create a mock client for testing the success path
		mockClient := provisioner.NewMockAPIClient(t)
		
		// Temporarily replace the client creator for this test
		originalCreator := containerengine.DefaultPodmanUserClientCreator
		containerengine.DefaultPodmanUserClientCreator = func() (client.APIClient, error) {
			return mockClient, nil
		}

		defer func() {
			containerengine.DefaultPodmanUserClientCreator = originalCreator
		}()
		
		// Test the success path
		client, err := containerengine.GetPodmanUserClient()
		
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, mockClient, client)
	})

	t.Run("GetPodmanSystemClient success path", func(t *testing.T) {
		t.Parallel()
		
		// Create a mock client for testing the success path
		mockClient := provisioner.NewMockAPIClient(t)
		
		// Temporarily replace the client creator for this test
		originalCreator := containerengine.DefaultPodmanSystemClientCreator
		containerengine.DefaultPodmanSystemClientCreator = func() (client.APIClient, error) {
			return mockClient, nil
		}

		defer func() {
			containerengine.DefaultPodmanSystemClientCreator = originalCreator
		}()
		
		// Test the success path
		client, err := containerengine.GetPodmanSystemClient()
		
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, mockClient, client)
	})

	t.Run("GetDockerClient error path", func(t *testing.T) {
		t.Parallel()
		
		// Temporarily replace the client creator to return an error
		originalCreator := containerengine.DefaultDockerClientCreator
		containerengine.DefaultDockerClientCreator = func() (client.APIClient, error) {
			return nil, errMockDockerCreation
		}

		defer func() {
			containerengine.DefaultDockerClientCreator = originalCreator
		}()
		
		// Test the error path
		client, err := containerengine.GetDockerClient()
		
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to create Docker client")
		assert.Contains(t, err.Error(), "mock docker creation error")
	})

	t.Run("GetPodmanUserClient error path", func(t *testing.T) {
		t.Parallel()
		
		// Temporarily replace the client creator to return an error
		originalCreator := containerengine.DefaultPodmanUserClientCreator
		containerengine.DefaultPodmanUserClientCreator = func() (client.APIClient, error) {
			return nil, errMockPodmanUserCreation
		}

		defer func() {
			containerengine.DefaultPodmanUserClientCreator = originalCreator
		}()
		
		// Test the error path
		client, err := containerengine.GetPodmanUserClient()
		
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to create Podman user client")
		assert.Contains(t, err.Error(), "mock podman user creation error")
	})

	t.Run("GetPodmanSystemClient error path", func(t *testing.T) {
		t.Parallel()
		
		// Temporarily replace the client creator to return an error
		originalCreator := containerengine.DefaultPodmanSystemClientCreator
		containerengine.DefaultPodmanSystemClientCreator = func() (client.APIClient, error) {
			return nil, errMockPodmanSystemCreation
		}

		defer func() {
			containerengine.DefaultPodmanSystemClientCreator = originalCreator
		}()
		
		// Test the error path
		client, err := containerengine.GetPodmanSystemClient()
		
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to create Podman system client")
		assert.Contains(t, err.Error(), "mock podman system creation error")
	})
}
