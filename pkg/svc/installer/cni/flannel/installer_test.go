package flannel_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/flannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const manifestURL = "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml"

var (
	errNetworkUnavailable = errors.New("network unavailable")
	errInvalidManifest    = errors.New(
		"failed to decode manifest document 0: yaml: line 1: did not find expected key",
	)
	errForbidden        = errors.New("forbidden")
	errNetworkDelete    = errors.New("network error")
	errDeletePermission = errors.New("forbidden: insufficient permissions")
)

type installTestCase struct {
	name          string
	mockSetup     func(*kubectl.MockInterface)
	waitFactory   func() func(context.Context) error
	expectError   bool
	errorContains string
}

func TestInstallerInstall(t *testing.T) {
	t.Parallel()

	for _, testCase := range flannelInstallerInstallTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runInstallTestCase(t, testCase)
		})
	}
}

type uninstallTestCase struct {
	name          string
	mockSetup     func(*kubectl.MockInterface)
	expectError   bool
	errorContains string
}

func TestInstallerUninstall(t *testing.T) {
	t.Parallel()

	for _, testCase := range flannelInstallerUninstallTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			runUninstallTestCase(t, testCase)
		})
	}
}

func TestInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	clientMock := kubectl.NewMockInterface(t)
	clientMock.On("Apply", mock.Anything, mock.Anything).Return(nil)

	installer := flannel.NewFlannelInstaller(
		clientMock,
		"/tmp/kubeconfig",
		"test-context",
		5*time.Minute,
	)

	called := false
	customWait := func(context.Context) error {
		called = true

		return nil
	}

	installer.SetWaitForReadinessFunc(customWait)

	require.NoError(t, installer.Install(context.Background()))
	assert.True(t, called, "custom wait function should have been called")
	clientMock.AssertExpectations(t)
}

func TestNewFlannelInstaller(t *testing.T) {
	t.Parallel()

	clientMock := kubectl.NewMockInterface(t)
	kubeconfig := "/tmp/kubeconfig"
	clusterContext := "test-context"
	timeout := 5 * time.Minute

	installer := flannel.NewFlannelInstaller(clientMock, kubeconfig, clusterContext, timeout)

	assert.NotNil(t, installer)
	assert.IsType(t, &flannel.Installer{}, installer)
}

func TestNewFlannelInstaller_NilClient(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		flannel.NewFlannelInstaller(nil, "/tmp/kubeconfig", "test-context", 5*time.Minute)
	})
}

func runInstallTestCase(t *testing.T, testCase installTestCase) {
	t.Helper()

	clientMock := kubectl.NewMockInterface(t)
	if testCase.mockSetup != nil {
		testCase.mockSetup(clientMock)
	}

	installer := flannel.NewFlannelInstaller(
		clientMock,
		"/tmp/kubeconfig",
		"test-context",
		5*time.Minute,
	)

	waitFactory := testCase.waitFactory
	if waitFactory == nil {
		waitFactory = func() func(context.Context) error { return nil }
	}

	if waitFn := waitFactory(); waitFn != nil {
		installer.SetWaitForReadinessFunc(waitFn)
	}

	err := installer.Install(context.Background())

	assertOutcome(t, err, testCase.expectError, testCase.errorContains)

	clientMock.AssertExpectations(t)
}

func runUninstallTestCase(t *testing.T, testCase uninstallTestCase) {
	t.Helper()

	clientMock := kubectl.NewMockInterface(t)
	if testCase.mockSetup != nil {
		testCase.mockSetup(clientMock)
	}

	installer := flannel.NewFlannelInstaller(
		clientMock,
		"/tmp/kubeconfig",
		"test-context",
		5*time.Minute,
	)

	err := installer.Uninstall(context.Background())

	assertOutcome(t, err, testCase.expectError, testCase.errorContains)

	clientMock.AssertExpectations(t)
}

func assertOutcome(t *testing.T, err error, expectError bool, errorContains string) {
	t.Helper()

	if expectError {
		require.Error(t, err)

		if errorContains != "" {
			assert.Contains(t, err.Error(), errorContains)
		}
	} else {
		require.NoError(t, err)
	}
}

func flannelInstallerInstallTestCases() []installTestCase {
	return []installTestCase{
		newSuccessfulInstallCase(),
		newNetworkErrorInstallCase(),
		newTimeoutInstallCase(),
		newInvalidManifestInstallCase(),
		newPermissionDeniedInstallCase(),
	}
}

func newSuccessfulInstallCase() installTestCase {
	return installTestCase{
		name: "successful installation",
		mockSetup: func(clientMock *kubectl.MockInterface) {
			clientMock.On(
				"Apply",
				mock.Anything,
				mock.MatchedBy(func(appliedURL string) bool {
					return appliedURL == manifestURL
				}),
			).Return(nil)
		},
		waitFactory: func() func(context.Context) error {
			return func(context.Context) error {
				return nil
			}
		},
	}
}

func newNetworkErrorInstallCase() installTestCase {
	return installTestCase{
		name: "network error during apply",
		mockSetup: func(clientMock *kubectl.MockInterface) {
			clientMock.On(
				"Apply",
				mock.Anything,
				mock.Anything,
			).Return(&url.Error{Op: "Get", URL: manifestURL, Err: errNetworkUnavailable})
		},
		expectError:   true,
		errorContains: "verify network access",
	}
}

func newTimeoutInstallCase() installTestCase {
	return installTestCase{
		name: "timeout during readiness",
		mockSetup: func(clientMock *kubectl.MockInterface) {
			clientMock.On("Apply", mock.Anything, mock.Anything).Return(nil)
		},
		waitFactory: func() func(context.Context) error {
			return func(context.Context) error {
				return context.DeadlineExceeded
			}
		},
		expectError:   true,
		errorContains: "timed out",
	}
}

func newInvalidManifestInstallCase() installTestCase {
	return installTestCase{
		name: "invalid manifest content",
		mockSetup: func(clientMock *kubectl.MockInterface) {
			clientMock.On("Apply", mock.Anything, mock.Anything).Return(errInvalidManifest)
		},
		expectError:   true,
		errorContains: "could not be decoded",
	}
}

func newPermissionDeniedInstallCase() installTestCase {
	return installTestCase{
		name: "permission denied error",
		mockSetup: func(clientMock *kubectl.MockInterface) {
			statusErr := apierrors.NewForbidden(
				schema.GroupResource{Group: "apps", Resource: "daemonsets"},
				"kube-flannel-ds",
				errForbidden,
			)

			clientMock.On("Apply", mock.Anything, mock.Anything).Return(statusErr)
		},
		expectError:   true,
		errorContains: "insufficient RBAC permissions",
	}
}

func flannelInstallerUninstallTestCases() []uninstallTestCase {
	return []uninstallTestCase{
		{
			name: "successful uninstall",
			mockSetup: func(clientMock *kubectl.MockInterface) {
				clientMock.On(
					"Delete",
					mock.Anything,
					"kube-flannel",
					"daemonset",
					"kube-flannel-ds",
				).Return(nil)
				clientMock.On(
					"Delete",
					mock.Anything,
					"",
					"namespace",
					"kube-flannel",
				).Return(nil)
			},
		},
		{
			name: "network error during delete",
			mockSetup: func(clientMock *kubectl.MockInterface) {
				clientMock.On(
					"Delete",
					mock.Anything,
					"kube-flannel",
					"daemonset",
					"kube-flannel-ds",
				).Return(errNetworkDelete)
			},
			expectError:   true,
			errorContains: errNetworkDelete.Error(),
		},
		{
			name: "permission denied",
			mockSetup: func(clientMock *kubectl.MockInterface) {
				clientMock.On(
					"Delete",
					mock.Anything,
					"kube-flannel",
					"daemonset",
					"kube-flannel-ds",
				).Return(errDeletePermission)
			},
			expectError:   true,
			errorContains: "forbidden",
		},
	}
}
