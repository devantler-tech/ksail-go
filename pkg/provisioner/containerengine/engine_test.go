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

			engine, err := containerengine.NewContainerEngine(mockClient, testCase.engineName)
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
	engineName  string
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
		engineName  string
		expectReady bool
		expectError bool
	}{
		{
			name: "container engine ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(pingResponse, nil)
			},
			engineName:  "Docker",
			expectReady: true,
			expectError: false,
		},
		{
			name: "container engine not ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(pingResponse, assert.AnError)
			},
			engineName:  "Docker",
			expectReady: false,
			expectError: true,
		},
	}
}

func TestContainerEngine_Name(t *testing.T) {
	t.Parallel()

	mockClient := provisioner.NewMockAPIClient(t)
	engine, err := containerengine.NewContainerEngine(mockClient, "Docker")
	require.NoError(t, err)

	assert.Equal(t, "Docker", engine.GetName())
}

func TestContainerEngine_GetClient(t *testing.T) {
	t.Parallel()
	mockClient := provisioner.NewMockAPIClient(t)
	engine, err := containerengine.NewContainerEngine(mockClient, "Docker")
	require.NoError(t, err)

	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngine_WithInjectedClient(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	engineName := "TestEngine"
	
	// Act
	engine, err := containerengine.NewContainerEngine(mockClient, engineName)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, engineName, engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngine_WithNilClient(t *testing.T) {
	t.Parallel()
	
	// Act
	engine, err := containerengine.NewContainerEngine(nil, "TestEngine")
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "apiClient cannot be nil")
}

func TestNewContainerEngine_WithEmptyEngineName(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	
	// Act
	engine, err := containerengine.NewContainerEngine(mockClient, "")
	
	// Assert
	require.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "engineName cannot be empty")
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
		
		// Test that we can inject a client and engine name
		engine, err := containerengine.NewContainerEngine(mockClient, "TestEngine")
		
		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.Equal(t, "TestEngine", engine.GetName())
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
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

func TestGetAutoDetectedClient_DockerNotReady_PodmanUserSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockDockerClient := provisioner.NewMockAPIClient(t)
	mockPodmanClient := provisioner.NewMockAPIClient(t)
	
	// Create client creators using simple map
	overrides := map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return mockDockerClient, nil
		},
		"podman-user": func() (client.APIClient, error) {
			return mockPodmanClient, nil
		},
		"podman-system": func() (client.APIClient, error) {
			return nil, errPodmanSystemUnavailable
		},
	}
	
	// Docker client succeeds but is not ready
	mockDockerClient.EXPECT().Ping(context.Background()).Return(completePing(), errDockerNotReady)
	
	// Podman user client succeeds and is ready  
	mockPodmanClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	
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
	
	// Create client creators using simple map
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
	
	// Podman user client succeeds but is not ready
	mockPodmanUserClient.EXPECT().Ping(context.Background()).Return(completePing(), errPodmanUserNotReady)
	
	// Podman system client succeeds and is ready
	mockPodmanSystemClient.EXPECT().Ping(context.Background()).Return(completePing(), nil)
	
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
	
	// Create client creators that all fail using simple map
	overrides := map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return nil, errDockerUnavailable
		},
		"podman-user": func() (client.APIClient, error) {
			return nil, errPodmanUserUnavailable
		},
		"podman-system": func() (client.APIClient, error) {
			return nil, errPodmanSystemUnavailable
		},
	}
	
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
	
	// Create client creators using simple map
	overrides := map[string]containerengine.ClientCreator{
		"docker": func() (client.APIClient, error) {
			return mockDockerClient, nil
		},
		"podman-user": func() (client.APIClient, error) {
			return mockPodmanUserClient, nil
		},
		"podman-system": func() (client.APIClient, error) {
			return mockPodmanSystemClient, nil
		},
	}
	
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
	
	// Act
	engine, err := containerengine.GetAutoDetectedClient(overrides)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

