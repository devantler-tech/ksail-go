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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClientFactory is a mock implementation of ClientFactory for testing
type MockClientFactory struct {
	mock.Mock
}

func (m *MockClientFactory) GetDockerClient() (client.APIClient, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.APIClient), args.Error(1)
}

func (m *MockClientFactory) GetPodmanUserClient() (client.APIClient, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.APIClient), args.Error(1)
}

func (m *MockClientFactory) GetPodmanSystemClient() (client.APIClient, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.APIClient), args.Error(1)
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
				assert.Error(t, err)
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

func TestNewContainerEngine_WithAvailableEngine(t *testing.T) {
	t.Parallel()
	// Test with actual environment - this tests the real functionality
	// Either we get an engine (if Docker/Podman is available) or an error
	engine, err := containerengine.NewContainerEngine(nil, "")
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
		// Test that nil client triggers auto-detection
		engine, err := containerengine.NewContainerEngine(nil, "")
		
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

// Test scenarios that might not be easily testable with real clients
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

func TestGetAutoDetectedClientWithFactory_DockerSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockFactory := &MockClientFactory{}
	mockClient := provisioner.NewMockAPIClient(t)
	
	// Docker client succeeds and is ready
	mockFactory.On("GetDockerClient").Return(mockClient, nil)
	mockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClientWithFactory(mockFactory)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
	
	mockFactory.AssertExpectations(t)
}

func TestGetAutoDetectedClientWithFactory_DockerNotReady_PodmanUserSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockFactory := &MockClientFactory{}
	mockDockerClient := provisioner.NewMockAPIClient(t)
	mockPodmanClient := provisioner.NewMockAPIClient(t)
	
	// Docker client succeeds but is not ready
	mockFactory.On("GetDockerClient").Return(mockDockerClient, nil)
	mockDockerClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("docker not ready"))
	
	// Podman user client succeeds and is ready  
	mockFactory.On("GetPodmanUserClient").Return(mockPodmanClient, nil)
	mockPodmanClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClientWithFactory(mockFactory)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, mockPodmanClient, engine.Client)
	
	mockFactory.AssertExpectations(t)
}

func TestGetAutoDetectedClientWithFactory_DockerFails_PodmanUserNotReady_PodmanSystemSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockFactory := &MockClientFactory{}
	mockPodmanUserClient := provisioner.NewMockAPIClient(t)
	mockPodmanSystemClient := provisioner.NewMockAPIClient(t)
	
	// Docker client creation fails
	mockFactory.On("GetDockerClient").Return(nil, errors.New("docker unavailable"))
	
	// Podman user client succeeds but is not ready
	mockFactory.On("GetPodmanUserClient").Return(mockPodmanUserClient, nil)
	mockPodmanUserClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("podman user not ready"))
	
	// Podman system client succeeds and is ready
	mockFactory.On("GetPodmanSystemClient").Return(mockPodmanSystemClient, nil)
	mockPodmanSystemClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	// Act
	engine, err := containerengine.GetAutoDetectedClientWithFactory(mockFactory)
	
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, mockPodmanSystemClient, engine.Client)
	
	mockFactory.AssertExpectations(t)
}

func TestGetAutoDetectedClientWithFactory_AllClientsFail(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockFactory := &MockClientFactory{}
	
	// All client creations fail
	mockFactory.On("GetDockerClient").Return(nil, errors.New("docker unavailable"))
	mockFactory.On("GetPodmanUserClient").Return(nil, errors.New("podman user unavailable"))
	mockFactory.On("GetPodmanSystemClient").Return(nil, errors.New("podman system unavailable"))
	
	// Act
	engine, err := containerengine.GetAutoDetectedClientWithFactory(mockFactory)
	
	// Assert
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
	
	mockFactory.AssertExpectations(t)
}

func TestGetAutoDetectedClientWithFactory_AllClientsCreateButNotReady(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockFactory := &MockClientFactory{}
	mockDockerClient := provisioner.NewMockAPIClient(t)
	mockPodmanUserClient := provisioner.NewMockAPIClient(t)
	mockPodmanSystemClient := provisioner.NewMockAPIClient(t)
	
	// All clients create successfully but none are ready
	mockFactory.On("GetDockerClient").Return(mockDockerClient, nil)
	mockDockerClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("docker not ready"))
	
	mockFactory.On("GetPodmanUserClient").Return(mockPodmanUserClient, nil)
	mockPodmanUserClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("podman user not ready"))
	
	mockFactory.On("GetPodmanSystemClient").Return(mockPodmanSystemClient, nil)
	mockPodmanSystemClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("podman system not ready"))
	
	// Act
	engine, err := containerengine.GetAutoDetectedClientWithFactory(mockFactory)
	
	// Assert
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
	
	mockFactory.AssertExpectations(t)
}

func TestDefaultClientFactory(t *testing.T) {
	t.Parallel()
	
	factory := &containerengine.DefaultClientFactory{}
	
	// Test GetDockerClient
	dockerClient, err := factory.GetDockerClient()
	if err != nil {
		assert.Nil(t, dockerClient)
	} else {
		assert.NotNil(t, dockerClient)
	}
	
	// Test GetPodmanUserClient  
	podmanUserClient, err := factory.GetPodmanUserClient()
	if err != nil {
		assert.Nil(t, podmanUserClient)
	} else {
		assert.NotNil(t, podmanUserClient)
	}
	
	// Test GetPodmanSystemClient
	podmanSystemClient, err := factory.GetPodmanSystemClient()
	if err != nil {
		assert.Nil(t, podmanSystemClient)
	} else {
		assert.NotNil(t, podmanSystemClient)
	}
}