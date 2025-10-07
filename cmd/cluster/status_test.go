package cluster //nolint:testpackage // Requires internal access to helper functions.

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/spf13/cobra"
)

// TestStatusCommandMetadata verifies the status command has correct metadata.
func TestStatusCommandMetadata(t *testing.T) {
	t.Parallel()

	cmd := NewStatusCmd(newTestRuntime())

	if cmd.Use != "status" {
		t.Errorf("expected Use to be 'status', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

var (
	errClientCreationFailed = errors.New("client creation failed")
	errGetStatusesFailed    = errors.New("get statuses failed")
)

// TestHandleStatusRunE tests the status command handler.
//
//nolint:paralleltest,funlen
func TestHandleStatusRunE(t *testing.T) {
	tests := []struct {
		name                    string
		setupMocks              func(*mockClientProvider, *mockComponentStatusProvider)
		expectedOutputSubstring string
		expectError             bool
		expectedErrorSubstring  string
	}{
		{
			name: "success with component statuses",
			setupMocks: func(mcp *mockClientProvider, mcsp *mockComponentStatusProvider) {
				mcp.On("CreateClient", mock.Anything, mock.Anything).
					Return(&kubernetes.Clientset{}, nil)

				statuses := []corev1.ComponentStatus{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "scheduler"},
						Conditions: []corev1.ComponentCondition{
							{
								Type:    corev1.ComponentHealthy,
								Status:  corev1.ConditionTrue,
								Message: "ok",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "controller-manager"},
						Conditions: []corev1.ComponentCondition{
							{
								Type:    corev1.ComponentHealthy,
								Status:  corev1.ConditionTrue,
								Message: "ok",
							},
						},
					},
				}

				mcsp.On("GetComponentStatuses", mock.Anything, mock.Anything).
					Return(statuses, nil)
			},
			expectedOutputSubstring: "scheduler",
			expectError:             false,
		},
		{
			name: "success with no component statuses",
			setupMocks: func(mcp *mockClientProvider, mcsp *mockComponentStatusProvider) {
				mcp.On("CreateClient", mock.Anything, mock.Anything).
					Return(&kubernetes.Clientset{}, nil)

				mcsp.On("GetComponentStatuses", mock.Anything, mock.Anything).
					Return([]corev1.ComponentStatus{}, nil)
			},
			expectedOutputSubstring: "no component statuses found",
			expectError:             false,
		},
		{
			name: "client creation failure",
			setupMocks: func(mcp *mockClientProvider, _ *mockComponentStatusProvider) {
				mcp.On("CreateClient", mock.Anything, mock.Anything).
					Return(nil, errClientCreationFailed)
			},
			expectError:            true,
			expectedErrorSubstring: "failed to create kubernetes client",
		},
		{
			name: "get component statuses failure",
			setupMocks: func(mcp *mockClientProvider, mcsp *mockComponentStatusProvider) {
				mcp.On("CreateClient", mock.Anything, mock.Anything).
					Return(&kubernetes.Clientset{}, nil)

				mcsp.On("GetComponentStatuses", mock.Anything, mock.Anything).
					Return(nil, errGetStatusesFailed)
			},
			expectError:            true,
			expectedErrorSubstring: "failed to get component statuses",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cleanup := testutils.SetupValidWorkingDir(t)
			t.Cleanup(cleanup)

			// Create mocks
			mockClient := &mockClientProvider{}
			mockStatusProvider := &mockComponentStatusProvider{}
			mockTimer := &mockTimer{}

			// Setup mocks
			testCase.setupMocks(mockClient, mockStatusProvider)
			mockTimer.On("Start").Return()
			mockTimer.On("GetTiming").Return(time.Duration(0), time.Duration(0))

			// Create command
			cmd := &cobra.Command{Use: "status"}
			cmd.SetContext(context.Background())

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Create config manager
			cfgManager := createConfigManager(t, &output)

			// Create deps
			deps := StatusDeps{
				Timer:                   mockTimer,
				ClientProvider:          mockClient,
				ComponentStatusProvider: mockStatusProvider,
			}

			// Execute
			err := HandleStatusRunE(cmd, cfgManager, deps)

			// Verify
			if testCase.expectError {
				assertError(t, err, testCase.expectedErrorSubstring)
			} else {
				assertSuccess(t, err, output.String(), testCase.expectedOutputSubstring)
			}

			mockClient.AssertExpectations(t)
			mockStatusProvider.AssertExpectations(t)
			mockTimer.AssertExpectations(t)
		})
	}
}

func assertError(t *testing.T, err error, expectedSubstring string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Fatalf("expected error to contain %q, got %q", expectedSubstring, err.Error())
	}
}

func assertSuccess(t *testing.T, err error, output, expectedSubstring string) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(output, expectedSubstring) {
		t.Fatalf("expected output to contain %q, got %q", expectedSubstring, output)
	}
}

// Mock types for testing

type mockClientProvider struct {
	mock.Mock
}

func (m *mockClientProvider) CreateClient(kubeconfig, context string) (*kubernetes.Clientset, error) {
	args := m.Called(kubeconfig, context)

	if args.Get(0) == nil {
		//nolint:wrapcheck // Test mock - returning error as-is from mock
		return nil, args.Error(1)
	}

	//nolint:forcetypeassert,wrapcheck // Test mock - type assertion and error are safe
	return args.Get(0).(*kubernetes.Clientset), args.Error(1)
}

type mockComponentStatusProvider struct {
	mock.Mock
}

func (m *mockComponentStatusProvider) GetComponentStatuses(
	ctx context.Context,
	clientset *kubernetes.Clientset,
) ([]corev1.ComponentStatus, error) {
	args := m.Called(ctx, clientset)

	if args.Get(0) == nil {
		//nolint:wrapcheck // Test mock - returning error as-is from mock
		return nil, args.Error(1)
	}

	//nolint:forcetypeassert,wrapcheck // Test mock - type assertion and error are safe
	return args.Get(0).([]corev1.ComponentStatus), args.Error(1)
}

type mockTimer struct {
	mock.Mock
}

func (m *mockTimer) Start() {
	m.Called()
}

func (m *mockTimer) NewStage() {
	m.Called()
}

func (m *mockTimer) GetTiming() (time.Duration, time.Duration) {
	args := m.Called()
	//nolint:forcetypeassert // Test mock - type assertion is safe
	return args.Get(0).(time.Duration), args.Get(1).(time.Duration)
}

func (m *mockTimer) Stop() {
	m.Called()
}

