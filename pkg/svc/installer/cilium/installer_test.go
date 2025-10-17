package ciliuminstaller //nolint:testpackage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func expectNoError(t *testing.T, err error, description string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: unexpected error: %v", description, err)
	}
}

func expectErrorContains(t *testing.T, err error, substr, description string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected error containing %q but got nil", description, substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("%s: expected error to contain %q, got %q", description, substr, err.Error())
	}
}

func expectNotNil(t *testing.T, value any, description string) {
	t.Helper()

	if value == nil {
		t.Fatalf("expected %s to be non-nil", description)
	}
}

func expectTrue(t *testing.T, condition bool, description string) {
	t.Helper()

	if !condition {
		t.Fatalf("expected %s to be true", description)
	}
}

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
				expectCiliumAddRepository(t, client, nil)
				expectCiliumInstallChart(t, client, nil)
			},
		},
		{
			name: "InstallFailure",
			setup: func(t *testing.T, client *MockHelmClient) {
				expectCiliumAddRepository(t, client, nil)
				expectCiliumInstallChart(t, client, errors.New("install failed"))
			},
			wantErr: "failed to install Cilium",
		},
		{
			name: "AddRepositoryFailure",
			setup: func(t *testing.T, client *MockHelmClient) {
				expectCiliumAddRepository(t, client, errors.New("add repo failed"))
			},
			wantErr: "failed to add cilium repository",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			installer, client := newDefaultInstaller(t)
			tc.setup(t, client)

			err := installer.Install(context.Background())

			if tc.wantErr == "" {
				expectNoError(t, err, "Install")

				return
			}

			expectErrorContains(t, err, tc.wantErr, "Install error")
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
				expectCiliumUninstall(t, client, nil)
			},
		},
		{
			name: "UninstallFailure",
			setup: func(t *testing.T, client *MockHelmClient) {
				expectCiliumUninstall(t, client, errors.New("uninstall failed"))
			},
			wantErr: "failed to uninstall cilium release",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			installer, client := newDefaultInstaller(t)
			tc.setup(t, client)

			err := installer.Uninstall(context.Background())

			if tc.wantErr == "" {
				expectNoError(t, err, "Uninstall")

				return
			}

			expectErrorContains(t, err, tc.wantErr, "Uninstall error")
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

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			applyDefaultValues(tc.spec)

			expectNotNil(t, tc.spec.SetJSONVals, "SetJSONVals map")
			expectEqual(
				t,
				tc.spec.SetJSONVals["operator.replicas"],
				tc.expectedValue,
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

	t.Run("ReadyOnFirstPoll", func(t *testing.T) {
		t.Parallel()

		client := fake.NewSimpleClientset(&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "cilium", Namespace: "kube-system"},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 1,
				NumberUnavailable:      0,
				UpdatedNumberScheduled: 1,
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := waitForDaemonSetReady(ctx, client, "kube-system", "cilium", 200*time.Millisecond)

		expectNoError(t, err, "waitForDaemonSetReady ready state")
	})

	t.Run("PropagatesAPIError", func(t *testing.T) {
		t.Parallel()

		client := fake.NewSimpleClientset()
		client.PrependReactor(
			"get",
			"daemonsets",
			func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("boom")
			},
		)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := waitForDaemonSetReady(ctx, client, "kube-system", "cilium", 200*time.Millisecond)

		expectErrorContains(
			t,
			err,
			"get daemonset kube-system/cilium: boom",
			"waitForDaemonSetReady api error",
		)
	})

	t.Run("TimesOutWhenNotReady", func(t *testing.T) {
		t.Parallel()

		client := fake.NewSimpleClientset()

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		err := waitForDaemonSetReady(ctx, client, "kube-system", "cilium", 150*time.Millisecond)

		expectErrorContains(t, err, "poll for readiness", "waitForDaemonSetReady timeout")
	})
}

func TestWaitForDeploymentReady(t *testing.T) {
	t.Parallel()

	t.Run("ReadyOnFirstPoll", func(t *testing.T) {
		t.Parallel()

		client := fake.NewSimpleClientset(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "cilium-operator", Namespace: "kube-system"},
			Status: appsv1.DeploymentStatus{
				Replicas:          1,
				UpdatedReplicas:   1,
				AvailableReplicas: 1,
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := waitForDeploymentReady(
			ctx,
			client,
			"kube-system",
			"cilium-operator",
			200*time.Millisecond,
		)

		expectNoError(t, err, "waitForDeploymentReady ready state")
	})

	t.Run("PropagatesAPIError", func(t *testing.T) {
		t.Parallel()

		client := fake.NewSimpleClientset()
		client.PrependReactor(
			"get",
			"deployments",
			func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("fail")
			},
		)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := waitForDeploymentReady(
			ctx,
			client,
			"kube-system",
			"cilium-operator",
			200*time.Millisecond,
		)

		expectErrorContains(
			t,
			err,
			"get deployment kube-system/cilium-operator: fail",
			"waitForDeploymentReady api error",
		)
	})

	t.Run("TimesOutWhenNotReady", func(t *testing.T) {
		t.Parallel()

		client := fake.NewSimpleClientset(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "cilium-operator", Namespace: "kube-system"},
			Status: appsv1.DeploymentStatus{
				Replicas:        2,
				UpdatedReplicas: 1,
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		err := waitForDeploymentReady(
			ctx,
			client,
			"kube-system",
			"cilium-operator",
			150*time.Millisecond,
		)

		expectErrorContains(t, err, "poll for readiness", "waitForDeploymentReady timeout")
	})
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
			return false, errors.New("boom")
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
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write kubeconfig file: %v", err)
	}

	return path
}
