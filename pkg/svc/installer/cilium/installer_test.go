package ciliuminstaller //nolint:testpackage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func expectEqual[T comparable](t *testing.T, got, want T, description string) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected %s: got %v want %v", description, got, want)
	}
}

var (
	errInstallFailed   = errors.New("install failed")
	errAddRepoFailed   = errors.New("add repo failed")
	errUninstallFailed = errors.New("uninstall failed")
	errDaemonSetBoom   = errors.New("boom")
	errDeploymentFail  = errors.New("fail")
	errPollBoom        = errors.New("boom")
)

func TestNewCiliumInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := NewMockHelmClient(t)
	installer := NewCiliumInstaller(client, kubeconfig, context, timeout)

	expectNotNil(t, installer, "installer instance")
}

func TestCiliumInstallerInstall(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		setup   func(*testing.T, *MockHelmClient)
		wantErr string
	}{
		{
			name: "Success",
			setup: func(t *testing.T, client *MockHelmClient) {
				t.Helper()

				expectCiliumAddRepository(t, client, nil)
				expectCiliumInstallChart(t, client, nil)
			},
		},
		{
			name: "InstallFailure",
			setup: func(t *testing.T, client *MockHelmClient) {
				t.Helper()

				expectCiliumAddRepository(t, client, nil)
				expectCiliumInstallChart(t, client, errInstallFailed)
			},
			wantErr: "failed to install Cilium",
		},
		{
			name: "AddRepositoryFailure",
			setup: func(t *testing.T, client *MockHelmClient) {
				t.Helper()

				expectCiliumAddRepository(t, client, errAddRepoFailed)
			},
			wantErr: "failed to add cilium repository",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			installer, client := newDefaultInstaller(t)
			testCase.setup(t, client)

			err := installer.Install(context.Background())

			if testCase.wantErr == "" {
				expectNoError(t, err, "Install")

				return
			}

			expectErrorContains(t, err, testCase.wantErr, "Install error")
		})
	}
}

func TestCiliumInstallerUninstall(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		setup   func(*testing.T, *MockHelmClient)
		wantErr string
	}{
		{
			name: "Success",
			setup: func(t *testing.T, client *MockHelmClient) {
				t.Helper()

				expectCiliumUninstall(t, client, nil)
			},
		},
		{
			name: "UninstallFailure",
			setup: func(t *testing.T, client *MockHelmClient) {
				t.Helper()

				expectCiliumUninstall(t, client, errUninstallFailed)
			},
			wantErr: "failed to uninstall cilium release",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			installer, client := newDefaultInstaller(t)
			testCase.setup(t, client)

			err := installer.Uninstall(context.Background())

			if testCase.wantErr == "" {
				expectNoError(t, err, "Uninstall")

				return
			}

			expectErrorContains(t, err, testCase.wantErr, "Uninstall error")
		})
	}
}

func TestApplyDefaultValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		spec          *helm.ChartSpec
		expectedValue string
	}{
		{
			name:          "SetsDefaultWhenMissing",
			spec:          &helm.ChartSpec{},
			expectedValue: "1",
		},
		{
			name: "PreservesExisting",
			spec: &helm.ChartSpec{
				SetJSONVals: map[string]string{"operator.replicas": "3"},
			},
			expectedValue: "3",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			applyDefaultValues(testCase.spec)

			expectNotNil(t, testCase.spec.SetJSONVals, "SetJSONVals map")
			expectEqual(
				t,
				testCase.spec.SetJSONVals["operator.replicas"],
				testCase.expectedValue,
				"operator replicas",
			)
		})
	}
}

func TestCiliumInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	t.Run("InvokesCustomFunction", func(t *testing.T) {
		t.Parallel()

		client := NewMockHelmClient(t)
		installer := NewCiliumInstaller(client, "kubeconfig", "", time.Second)
		called := false

		installer.SetWaitForReadinessFunc(func(context.Context) error {
			called = true

			return nil
		})

		expectNoError(
			t,
			installer.WaitForReadiness(context.Background()),
			"WaitForReadiness with custom func",
		)
		expectTrue(t, called, "custom wait function invocation")
	})

	t.Run("RestoresDefaultWhenNil", func(t *testing.T) {
		t.Parallel()

		client := NewMockHelmClient(t)
		installer := NewCiliumInstaller(client, "kubeconfig", "", time.Second)
		defaultFn := installer.waitFn
		expectNotNil(t, defaultFn, "default wait function")
		defaultPtr := reflect.ValueOf(defaultFn).Pointer()

		installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

		replacedPtr := reflect.ValueOf(installer.waitFn).Pointer()
		if replacedPtr == defaultPtr {
			t.Fatal("expected custom wait function to replace default")
		}

		installer.SetWaitForReadinessFunc(nil)
		restoredPtr := reflect.ValueOf(installer.waitFn).Pointer()
		expectEqual(t, restoredPtr, defaultPtr, "wait function pointer after restore")
	})
}

func TestCiliumInstallerWaitForReadinessBuildConfigError(t *testing.T) {
	t.Parallel()

	installer := NewCiliumInstaller(NewMockHelmClient(t), "", "", time.Second)
	err := installer.WaitForReadiness(context.Background())

	expectErrorContains(t, err, "build kubernetes client config", "WaitForReadiness error path")
}

func TestCiliumInstallerBuildRESTConfig(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWhenKubeconfigMissing", func(t *testing.T) {
		t.Parallel()

		installer := NewCiliumInstaller(NewMockHelmClient(t), "", "", time.Second)
		_, err := installer.buildRESTConfig()

		expectErrorContains(t, err, "kubeconfig path is empty", "buildRESTConfig empty path")
	})

	t.Run("UsesCurrentContext", func(t *testing.T) {
		t.Parallel()

		path := writeKubeconfig(t, t.TempDir())
		installer := NewCiliumInstaller(NewMockHelmClient(t), path, "", time.Second)

		restConfig, err := installer.buildRESTConfig()

		expectNoError(t, err, "buildRESTConfig current context")
		expectEqual(t, restConfig.Host, "https://cluster-one.example.com", "rest config host")
	})

	t.Run("OverridesContext", func(t *testing.T) {
		t.Parallel()

		path := writeKubeconfig(t, t.TempDir())
		installer := NewCiliumInstaller(NewMockHelmClient(t), path, "alt", time.Second)

		restConfig, err := installer.buildRESTConfig()

		expectNoError(t, err, "buildRESTConfig override context")
		expectEqual(
			t,
			restConfig.Host,
			"https://cluster-two.example.com",
			"rest config host override",
		)
	})

	t.Run("MissingContext", func(t *testing.T) {
		t.Parallel()

		path := writeKubeconfig(t, t.TempDir())
		installer := NewCiliumInstaller(NewMockHelmClient(t), path, "missing", time.Second)
		_, err := installer.buildRESTConfig()

		expectErrorContains(
			t,
			err,
			"context \"missing\" does not exist",
			"buildRESTConfig missing context",
		)
	})
}

func TestWaitForDaemonSetReady(t *testing.T) {
	t.Parallel()

	t.Run("ReadyOnFirstPoll", testWaitForDaemonSetReadyReady)
	t.Run("PropagatesAPIError", testWaitForDaemonSetReadyAPIError)
	t.Run("TimesOutWhenNotReady", testWaitForDaemonSetReadyTimeout)
}

func testWaitForDaemonSetReadyReady(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "kube-system"
		name      = "cilium"
	)

	client := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberUnavailable:      0,
			UpdatedNumberScheduled: 1,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitForDaemonSetReady(ctx, client, namespace, name, 200*time.Millisecond)

	expectNoError(t, err, "waitForDaemonSetReady ready state")
}

func testWaitForDaemonSetReadyAPIError(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "observability"
		name      = "cilium-agent"
	)

	client := fake.NewSimpleClientset()
	client.PrependReactor(
		"get",
		"daemonsets",
		func(_ k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errDaemonSetBoom
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitForDaemonSetReady(ctx, client, namespace, name, 200*time.Millisecond)

	expectErrorContains(
		t,
		err,
		"get daemonset observability/cilium-agent: boom",
		"waitForDaemonSetReady api error",
	)
}

func testWaitForDaemonSetReadyTimeout(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "networking"
		name      = "cilium"
	)

	client := fake.NewSimpleClientset()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := waitForDaemonSetReady(ctx, client, namespace, name, 150*time.Millisecond)

	expectErrorContains(t, err, "poll for readiness", "waitForDaemonSetReady timeout")
}

func TestWaitForDeploymentReady(t *testing.T) {
	t.Parallel()

	t.Run("ReadyOnFirstPoll", testWaitForDeploymentReadyReady)
	t.Run("PropagatesAPIError", testWaitForDeploymentReadyAPIError)
	t.Run("TimesOutWhenNotReady", testWaitForDeploymentReadyTimeout)
}

func testWaitForDeploymentReadyReady(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "kube-system"
		name      = "cilium-operator"
	)

	client := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DeploymentStatus{
			Replicas:          1,
			UpdatedReplicas:   1,
			AvailableReplicas: 1,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitForDeploymentReady(ctx, client, namespace, name, 200*time.Millisecond)

	expectNoError(t, err, "waitForDeploymentReady ready state")
}

func testWaitForDeploymentReadyAPIError(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "platform-system"
		name      = "cilium-operator"
	)

	client := fake.NewSimpleClientset()
	client.PrependReactor(
		"get",
		"deployments",
		func(_ k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errDeploymentFail
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitForDeploymentReady(ctx, client, namespace, name, 200*time.Millisecond)

	expectErrorContains(
		t,
		err,
		"get deployment platform-system/cilium-operator: fail",
		"waitForDeploymentReady api error",
	)
}

func testWaitForDeploymentReadyTimeout(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "observability"
		name      = "cilium-operator"
	)

	client := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DeploymentStatus{
			Replicas:        2,
			UpdatedReplicas: 1,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := waitForDeploymentReady(ctx, client, namespace, name, 150*time.Millisecond)

	expectErrorContains(t, err, "poll for readiness", "waitForDeploymentReady timeout")
}

func TestPollForReadiness(t *testing.T) {
	t.Parallel()

	t.Run("ReturnsNilWhenReady", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := pollForReadiness(ctx, 200*time.Millisecond, func(context.Context) (bool, error) {
			return true, nil
		})

		expectNoError(t, err, "pollForReadiness success")
	})

	t.Run("WrapsErrors", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := pollForReadiness(ctx, 200*time.Millisecond, func(context.Context) (bool, error) {
			return false, errPollBoom
		})

		expectErrorContains(t, err, "poll for readiness: boom", "pollForReadiness error wrap")
	})
}

func newDefaultInstaller(t *testing.T) (*CiliumInstaller, *MockHelmClient) {
	t.Helper()
	client := NewMockHelmClient(t)
	installer := NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer, client
}

func expectCiliumAddRepository(t *testing.T, client *MockHelmClient, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				t.Helper()
				expectEqual(t, entry.Name, "cilium", "repository name")
				expectEqual(t, entry.URL, "https://helm.cilium.io", "repository URL")

				return true
			}),
		).
		Return(err)
}

func expectCiliumInstallChart(t *testing.T, client *MockHelmClient, installErr error) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				t.Helper()
				expectEqual(t, spec.ReleaseName, "cilium", "release name")
				expectEqual(t, spec.ChartName, "cilium/cilium", "chart name")
				expectEqual(t, spec.Namespace, "kube-system", "namespace")
				expectEqual(t, spec.RepoURL, "https://helm.cilium.io", "repository URL")
				expectTrue(t, spec.Wait, "Wait flag")
				expectTrue(t, spec.WaitForJobs, "WaitForJobs flag")
				expectEqual(t, spec.SetJSONVals["operator.replicas"], "1", "operator replicas")

				return true
			}),
		).
		Return(nil, installErr)
}

func expectCiliumUninstall(t *testing.T, client *MockHelmClient, err error) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "cilium", "kube-system").
		Return(err)
}

func writeKubeconfig(t *testing.T, dir string) string {
	t.Helper()

	contents := `apiVersion: v1
kind: Config
clusters:
- name: cluster-one
	cluster:
		server: https://cluster-one.example.com
- name: cluster-two
	cluster:
		server: https://cluster-two.example.com
contexts:
- name: primary
	context:
		cluster: cluster-one
		user: user-one
- name: alt
	context:
		cluster: cluster-two
		user: user-two
current-context: primary
users:
- name: user-one
	user:
		token: token-one
- name: user-two
	user:
		token: token-two
`

	path := filepath.Join(dir, "config")

	err := os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("write kubeconfig file: %v", err)
	}

	return path
}
