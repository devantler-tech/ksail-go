package flannelinstaller //nolint:testpackage

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
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
	errPollBoom        = errors.New("boom")
)

func TestNewFlannelInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := helm.NewMockInterface(t)
	installer := NewFlannelInstaller(client, kubeconfig, context, timeout)

	testutils.ExpectNotNil(t, installer, "installer instance")
}

type installerScenario struct {
	name       string
	setup      func(*testing.T, *helm.MockInterface)
	actionName string
	action     func(context.Context, *FlannelInstaller) error
	wantErr    string
}

func runInstallerScenarios(t *testing.T, scenarios []installerScenario) {
	t.Helper()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			installer, client := newDefaultInstaller(t)
			scenario.setup(t, client)

			err := scenario.action(context.Background(), installer)

			expectInstallerResult(t, err, scenario.wantErr, scenario.actionName)
		})
	}
}

func TestFlannelInstallerInstall(t *testing.T) {
	t.Parallel()

	installAction := func(ctx context.Context, installer *FlannelInstaller) error {
		return installer.Install(ctx)
	}

	scenarios := []installerScenario{
		{
			name:       "Success",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupFlannelInstallExpectations(t, client, nil)
			},
		},
		{
			name:       "InstallFailure",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupFlannelInstallExpectations(t, client, errInstallFailed)
			},
			wantErr: "failed to install Flannel",
		},
		{
			name:       "AddRepositoryFailure",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectFlannelAddRepository(t, client, errAddRepoFailed)
			},
			wantErr: "failed to add flannel repository",
		},
	}

	runInstallerScenarios(t, scenarios)
}

func TestFlannelInstallerUninstall(t *testing.T) {
	t.Parallel()

	uninstallAction := func(ctx context.Context, installer *FlannelInstaller) error {
		return installer.Uninstall(ctx)
	}

	scenarios := []installerScenario{
		{
			name:       "Success",
			actionName: "Uninstall",
			action:     uninstallAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectFlannelUninstall(t, client, nil)
			},
		},
		{
			name:       "UninstallFailure",
			actionName: "Uninstall",
			action:     uninstallAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectFlannelUninstall(t, client, errUninstallFailed)
			},
			wantErr: "failed to uninstall flannel release",
		},
	}

	runInstallerScenarios(t, scenarios)
}

func TestFlannelInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	t.Run("InvokesCustomFunction", func(t *testing.T) {
		t.Parallel()

		client := helm.NewMockInterface(t)
		installer := NewFlannelInstaller(client, "kubeconfig", "", time.Second)
		called := false

		installer.SetWaitForReadinessFunc(func(context.Context) error {
			called = true

			return nil
		})

		testutils.ExpectNoError(
			t,
			installer.WaitForReadiness(context.Background()),
			"WaitForReadiness with custom func",
		)
		testutils.ExpectTrue(t, called, "custom wait function invocation")
	})

	t.Run("RestoresDefaultWhenNil", func(t *testing.T) {
		t.Parallel()

		client := helm.NewMockInterface(t)
		installer := NewFlannelInstaller(client, "kubeconfig", "", time.Second)
		defaultFn := installer.waitFn
		testutils.ExpectNotNil(t, defaultFn, "default wait function")
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

func TestFlannelInstallerWaitForReadinessBuildConfigError(t *testing.T) {
	t.Parallel()

	installer := NewFlannelInstaller(helm.NewMockInterface(t), "", "", time.Second)
	err := installer.WaitForReadiness(context.Background())

	testutils.ExpectErrorContains(
		t,
		err,
		"build kubernetes client config",
		"WaitForReadiness error path",
	)
}

func TestFlannelInstallerWaitForReadinessNoOpWhenUnset(t *testing.T) {
	t.Parallel()

	installer := NewFlannelInstaller(helm.NewMockInterface(t), "kubeconfig", "", time.Second)
	installer.waitFn = nil

	err := installer.WaitForReadiness(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when waitFn unset, got %v", err)
	}
}

func TestFlannelInstallerWaitForReadinessSuccess(t *testing.T) {
	t.Parallel()

	server := newFlannelAPIServer(t, true)
	t.Cleanup(server.Close)

	kubeconfig := writeServerBackedKubeconfig(t, server.URL)

	installer := NewFlannelInstaller(
		helm.NewMockInterface(t),
		kubeconfig,
		"default",
		100*time.Millisecond,
	)

	err := installer.waitForReadiness(context.Background())
	if err != nil {
		t.Fatalf("expected readiness checks to succeed, got %v", err)
	}
}

func TestFlannelInstallerWaitForReadinessDetectsUnreadyComponents(t *testing.T) {
	t.Parallel()

	server := newFlannelAPIServer(t, false)
	t.Cleanup(server.Close)

	kubeconfig := writeServerBackedKubeconfig(t, server.URL)

	installer := NewFlannelInstaller(
		helm.NewMockInterface(t),
		kubeconfig,
		"default",
		75*time.Millisecond,
	)

	err := installer.waitForReadiness(context.Background())
	if err == nil {
		t.Fatal("expected readiness failure when components are unready")
	}

	if !strings.Contains(err.Error(), "not ready") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFlannelInstallerBuildRESTConfig(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWhenKubeconfigMissing", func(t *testing.T) {
		t.Parallel()

		installer := NewFlannelInstaller(helm.NewMockInterface(t), "", "", time.Second)
		_, err := installer.buildRESTConfig()

		testutils.ExpectErrorContains(
			t,
			err,
			"kubeconfig path is empty",
			"buildRESTConfig empty path",
		)
	})

	t.Run("UsesCurrentContext", func(t *testing.T) {
		t.Parallel()

		path := writeKubeconfig(t, t.TempDir())
		installer := NewFlannelInstaller(helm.NewMockInterface(t), path, "", time.Second)

		restConfig, err := installer.buildRESTConfig()

		testutils.ExpectNoError(t, err, "buildRESTConfig current context")
		expectEqual(t, restConfig.Host, "https://cluster-one.example.com", "rest config host")
	})

	t.Run("OverridesContext", func(t *testing.T) {
		t.Parallel()

		path := writeKubeconfig(t, t.TempDir())
		installer := NewFlannelInstaller(helm.NewMockInterface(t), path, "alt", time.Second)

		restConfig, err := installer.buildRESTConfig()

		testutils.ExpectNoError(t, err, "buildRESTConfig override context")
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
		installer := NewFlannelInstaller(helm.NewMockInterface(t), path, "missing", time.Second)
		_, err := installer.buildRESTConfig()

		testutils.ExpectErrorContains(
			t,
			err,
			"context \"missing\" does not exist",
			"buildRESTConfig missing context",
		)
	})
}

func newFlannelAPIServer(t *testing.T, ready bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/apis/apps/v1/namespaces/kube-flannel/daemonsets/kube-flannel-ds":
			payload := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "DaemonSet",
				"status": map[string]any{
					"desiredNumberScheduled": 1,
					"numberUnavailable":      0,
					"updatedNumberScheduled": 1,
				},
			}

			status, ok := payload["status"].(map[string]any)
			if !ok {
				t.Fatalf("unexpected payload status type %T", payload["status"])
			}

			if !ready {
				status["numberUnavailable"] = 1
				status["updatedNumberScheduled"] = 0
			}

			encodeJSON(t, writer, payload)

		default:
			http.NotFound(writer, req)
		}
	}))
}

func encodeJSON(t *testing.T, writer http.ResponseWriter, payload any) {
	t.Helper()

	writer.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(writer)

	err := encoder.Encode(payload)
	if err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}

func writeServerBackedKubeconfig(t *testing.T, serverURL string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "kubeconfig.yaml")

	content := "apiVersion: v1\n" +
		"clusters:\n" +
		"- cluster:\n" +
		"    server: " + serverURL + "\n" +
		"    insecure-skip-tls-verify: true\n" +
		"  name: local\n" +
		"contexts:\n" +
		"- context:\n" +
		"    cluster: local\n" +
		"    user: default\n" +
		"  name: default\n" +
		"current-context: default\n" +
		"kind: Config\n" +
		"preferences: {}\n" +
		"users:\n" +
		"- name: default\n" +
		"  user: {}\n"

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}

	return path
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
		namespace = "kube-flannel"
		name      = "kube-flannel-ds"
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

	testutils.ExpectNoError(t, err, "waitForDaemonSetReady ready state")
}

func testWaitForDaemonSetReadyAPIError(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "observability"
		name      = "flannel-agent"
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

	testutils.ExpectErrorContains(
		t,
		err,
		"get daemonset observability/flannel-agent: boom",
		"waitForDaemonSetReady api error",
	)
}

func testWaitForDaemonSetReadyTimeout(t *testing.T) {
	t.Helper()
	t.Parallel()

	const (
		namespace = "networking"
		name      = "kube-flannel-ds"
	)

	client := fake.NewSimpleClientset()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := waitForDaemonSetReady(ctx, client, namespace, name, 150*time.Millisecond)

	testutils.ExpectErrorContains(t, err, "poll for readiness", "waitForDaemonSetReady timeout")
}

func TestPollForReadiness(t *testing.T) {
	t.Parallel()

	t.Run("ReturnsNilWhenReady", func(t *testing.T) {
		t.Parallel()

		err := pollForReadinessWithDefaultTimeout(t, func(context.Context) (bool, error) {
			return true, nil
		})

		testutils.ExpectNoError(t, err, "pollForReadiness success")
	})

	t.Run("WrapsErrors", func(t *testing.T) {
		t.Parallel()

		err := pollForReadinessWithDefaultTimeout(t, func(context.Context) (bool, error) {
			return false, errPollBoom
		})

		testutils.ExpectErrorContains(
			t,
			err,
			"poll for readiness: boom",
			"pollForReadiness error wrap",
		)
	})
}

func newDefaultInstaller(t *testing.T) (*FlannelInstaller, *helm.MockInterface) {
	t.Helper()
	client := helm.NewMockInterface(t)
	installer := NewFlannelInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer, client
}

func expectInstallerResult(t *testing.T, err error, wantErr, operation string) {
	t.Helper()

	if wantErr == "" {
		testutils.ExpectNoError(t, err, operation)

		return
	}

	message := operation + " error"
	testutils.ExpectErrorContains(t, err, wantErr, message)
}

func setupFlannelInstallExpectations(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()

	expectFlannelAddRepository(t, client, nil)
	expectFlannelInstallChart(t, client, installErr)
}

func pollForReadinessWithDefaultTimeout(
	t *testing.T,
	checker func(context.Context) (bool, error),
) error {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	return pollForReadiness(ctx, 200*time.Millisecond, checker)
}

func expectFlannelAddRepository(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				t.Helper()
				expectEqual(t, entry.Name, "flannel", "repository name")
				expectEqual(t, entry.URL, "https://flannel-io.github.io/flannel", "repository URL")

				return true
			}),
		).
		Return(err)
}

func expectFlannelInstallChart(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				t.Helper()
				expectEqual(t, spec.ReleaseName, "flannel", "release name")
				expectEqual(t, spec.ChartName, "flannel/flannel", "chart name")
				expectEqual(t, spec.Namespace, "kube-flannel", "namespace")
				expectEqual(t, spec.RepoURL, "https://flannel-io.github.io/flannel", "repository URL")
				testutils.ExpectTrue(t, spec.Wait, "Wait flag")
				testutils.ExpectTrue(t, spec.WaitForJobs, "WaitForJobs flag")
				testutils.ExpectTrue(t, spec.CreateNamespace, "CreateNamespace flag")

				return true
			}),
		).
		Return(nil, installErr)
}

func expectFlannelUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "flannel", "kube-flannel").
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
