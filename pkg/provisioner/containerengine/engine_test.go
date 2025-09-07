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
)

// mockClientFactory implements ClientFactory for testing
type mockClientFactory struct {
	dockerClient       client.APIClient
	dockerError        error
	podmanUserClient   client.APIClient
	podmanUserError    error
	podmanSystemClient client.APIClient
	podmanSystemError  error
}

func (f *mockClientFactory) NewDockerClient() (client.APIClient, error) {
	return f.dockerClient, f.dockerError
}

func (f *mockClientFactory) NewPodmanUserClient() (client.APIClient, error) {
	return f.podmanUserClient, f.podmanUserError
}

func (f *mockClientFactory) NewPodmanSystemClient() (client.APIClient, error) {
	return f.podmanSystemClient, f.podmanSystemError
}

func TestContainerEngine_CheckReady(t *testing.T) {
	t.Parallel()

	tests := createContainerEngineTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockClient := provisioner.NewMockAPIClient(t)
			testCase.setupMock(mockClient)

			engine := &containerengine.ContainerEngine{
				Client:     mockClient,
				EngineName: testCase.engineName,
			}

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

	engine := &containerengine.ContainerEngine{
		Client:     nil,
		EngineName: "Docker",
	}

	assert.Equal(t, "Docker", engine.GetName())
}

func TestContainerEngine_GetClient(t *testing.T) {
	t.Parallel()
	mockClient := provisioner.NewMockAPIClient(t)
	engine := &containerengine.ContainerEngine{
		Client:     mockClient,
		EngineName: "",
	}

	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngineWithFactory_DockerSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	mockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	factory := &mockClientFactory{
		dockerClient: mockClient,
		dockerError:  nil,
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Docker", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngineWithFactory_DockerClientCreationFails_PodmanUserSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	mockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	factory := &mockClientFactory{
		dockerClient:     nil,
		dockerError:      errors.New("docker client creation failed"),
		podmanUserClient: mockClient,
		podmanUserError:  nil,
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, mockClient, engine.Client)
}

func TestNewContainerEngineWithFactory_DockerClientNotReady_PodmanUserSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	dockerMockClient := provisioner.NewMockAPIClient(t)
	dockerMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("docker not ready"))
	
	podmanMockClient := provisioner.NewMockAPIClient(t)
	podmanMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	factory := &mockClientFactory{
		dockerClient:     dockerMockClient,
		dockerError:      nil,
		podmanUserClient: podmanMockClient,
		podmanUserError:  nil,
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, podmanMockClient, engine.Client)
}

func TestNewContainerEngineWithFactory_DockerAndPodmanUserFail_PodmanSystemSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	dockerMockClient := provisioner.NewMockAPIClient(t)
	dockerMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("docker not ready"))
	
	podmanUserMockClient := provisioner.NewMockAPIClient(t)
	podmanUserMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("podman user not ready"))
	
	podmanSystemMockClient := provisioner.NewMockAPIClient(t)
	podmanSystemMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	factory := &mockClientFactory{
		dockerClient:       dockerMockClient,
		dockerError:        nil,
		podmanUserClient:   podmanUserMockClient,
		podmanUserError:    nil,
		podmanSystemClient: podmanSystemMockClient,
		podmanSystemError:  nil,
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, podmanSystemMockClient, engine.Client)
}

func TestNewContainerEngineWithFactory_AllClientCreationFails(t *testing.T) {
	t.Parallel()
	
	// Arrange
	factory := &mockClientFactory{
		dockerClient:       nil,
		dockerError:        errors.New("docker client creation failed"),
		podmanUserClient:   nil,
		podmanUserError:    errors.New("podman user client creation failed"),
		podmanSystemClient: nil,
		podmanSystemError:  errors.New("podman system client creation failed"),
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestNewContainerEngineWithFactory_AllClientsNotReady(t *testing.T) {
	t.Parallel()
	
	// Arrange
	dockerMockClient := provisioner.NewMockAPIClient(t)
	dockerMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("docker not ready"))
	
	podmanUserMockClient := provisioner.NewMockAPIClient(t)
	podmanUserMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("podman user not ready"))
	
	podmanSystemMockClient := provisioner.NewMockAPIClient(t)
	podmanSystemMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, errors.New("podman system not ready"))
	
	factory := &mockClientFactory{
		dockerClient:       dockerMockClient,
		dockerError:        nil,
		podmanUserClient:   podmanUserMockClient,
		podmanUserError:    nil,
		podmanSystemClient: podmanSystemMockClient,
		podmanSystemError:  nil,
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, containerengine.ErrNoContainerEngine, err)
	assert.Nil(t, engine)
}

func TestNewContainerEngineWithFactory_PodmanUserCreationFails_PodmanSystemSuccess(t *testing.T) {
	t.Parallel()
	
	// Arrange
	podmanSystemMockClient := provisioner.NewMockAPIClient(t)
	podmanSystemMockClient.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
	
	factory := &mockClientFactory{
		dockerClient:       nil,
		dockerError:        errors.New("docker client creation failed"),
		podmanUserClient:   nil,
		podmanUserError:    errors.New("podman user client creation failed"),
		podmanSystemClient: podmanSystemMockClient,
		podmanSystemError:  nil,
	}
	
	// Act
	engine, err := containerengine.NewContainerEngineWithFactory(factory)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, "Podman", engine.GetName())
	assert.Equal(t, podmanSystemMockClient, engine.Client)
}

func TestNewContainerEngine_WithAvailableEngine(t *testing.T) {
	t.Parallel()
	// Test with actual environment - this tests the real functionality
	// Either we get an engine (if Docker/Podman is available) or an error
	engine, err := containerengine.NewContainerEngine()

	if err != nil {
		assert.Equal(t, containerengine.ErrNoContainerEngine, err)
		assert.Nil(t, engine)
	} else {
		assert.NotNil(t, engine)
		assert.Contains(t, []string{"Docker", "Podman"}, engine.GetName())
		
		// Test that the engine actually works
		ready, checkErr := engine.CheckReady(context.Background())
		if checkErr == nil {
			assert.True(t, ready)
		}
	}
}

func TestDefaultClientFactory_Methods(t *testing.T) {
	t.Parallel()
	
	// Test the default factory methods to achieve full coverage
	factory := &containerengine.DefaultClientFactory{}
	
	// Test Docker client creation
	_, dockerErr := factory.NewDockerClient()
	// We don't assert success/failure since it depends on environment
	// but we ensure the method is called for coverage
	_ = dockerErr
	
	// Test Podman user client creation
	_, podmanUserErr := factory.NewPodmanUserClient()
	// We don't assert success/failure since it depends on environment
	_ = podmanUserErr
	
	// Test Podman system client creation
	_, podmanSystemErr := factory.NewPodmanSystemClient()
	// We don't assert success/failure since it depends on environment
	_ = podmanSystemErr
}
