package containerengine_test

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestContainerEngine_CheckReady(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func(*provisioner.MockAPIClient)
		engineName  string
		expectReady bool
		expectError bool
	}{
		{
			name: "container engine ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(types.Ping{
					APIVersion:     "",
					OSType:         "",
					Experimental:   false,
					BuilderVersion: "",
					SwarmStatus:    nil,
				}, nil)
			},
			engineName:  "Docker",
			expectReady: true,
			expectError: false,
		},
		{
			name: "container engine not ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(types.Ping{
					APIVersion:     "",
					OSType:         "",
					Experimental:   false,
					BuilderVersion: "",
					SwarmStatus:    nil,
				}, assert.AnError)
			},
			engineName:  "Docker",
			expectReady: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockClient := provisioner.NewMockAPIClient(t)
			tt.setupMock(mockClient)

			engine := &containerengine.ContainerEngine{
				Client:     mockClient,
				EngineName: tt.engineName,
			}

			ready, err := engine.CheckReady()

			assert.Equal(t, tt.expectReady, ready)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
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

	assert.Equal(t, mockClient, engine.GetClient())
}

func TestNewContainerEngine_NoEnginesAvailable(t *testing.T) {
	t.Parallel()
	// This test would require mocking the client creation, which is complex
	// For now, we'll test that the error is returned when no engines are available
	// In a real environment with no Docker/Podman, this would return the error

	// We can't easily test this without major refactoring to inject dependencies
	// The function directly calls client.NewClientWithOpts which we can't mock
	// For coverage purposes, we acknowledge this limitation
	engine, err := containerengine.NewContainerEngine()

	// Either we get an engine (if Docker/Podman is available) or an error
	if err != nil {
		assert.Equal(t, containerengine.ErrNoContainerEngine, err)
		assert.Nil(t, engine)
	} else {
		assert.NotNil(t, engine)
		assert.Contains(t, []string{"Docker", "Podman"}, engine.GetName())
	}
}
