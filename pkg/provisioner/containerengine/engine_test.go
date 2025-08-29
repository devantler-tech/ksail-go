package containerengine

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestUnifiedContainerEngine_CheckReady(t *testing.T) {
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
				m.EXPECT().Ping(context.Background()).Return(types.Ping{}, nil)
			},
			engineName:  "Docker",
			expectReady: true,
			expectError: false,
		},
		{
			name: "container engine not ready",
			setupMock: func(m *provisioner.MockAPIClient) {
				m.EXPECT().Ping(context.Background()).Return(types.Ping{}, assert.AnError)
			},
			engineName:  "Docker",
			expectReady: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := provisioner.NewMockAPIClient(t)
			tt.setupMock(mockClient)

			engine := &UnifiedContainerEngine{
				client: mockClient,
				name:   tt.engineName,
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

func TestUnifiedContainerEngine_Name(t *testing.T) {
	engine := &UnifiedContainerEngine{
		name: "Docker",
	}

	assert.Equal(t, "Docker", engine.Name())
}