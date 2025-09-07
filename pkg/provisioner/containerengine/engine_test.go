package containerengine_test

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertAutoDetectionResult is a helper function to avoid code duplication
// when testing auto-detection behavior of NewContainerEngine
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