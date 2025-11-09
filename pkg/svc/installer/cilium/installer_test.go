package ciliuminstaller //nolint:testpackage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	installertestutils "github.com/devantler-tech/ksail-go/pkg/svc/installer/testutils"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/stretchr/testify/mock"
)

func TestNewCiliumInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := helm.NewMockInterface(t)
	installer := NewCiliumInstaller(client, kubeconfig, context, timeout)

	testutils.ExpectNotNil(t, installer, "installer instance")
}

type installerScenario struct {
	name       string
	setup      func(*testing.T, *helm.MockInterface)
	actionName string
	action     func(context.Context, *CiliumInstaller) error
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

			installertestutils.ExpectInstallerResult(t, err, scenario.wantErr, scenario.actionName)
		})
	}
}

func TestCiliumInstallerInstall(t *testing.T) {
	t.Parallel()

	installAction := func(ctx context.Context, installer *CiliumInstaller) error {
		return installer.Install(ctx)
	}

	scenarios := []installerScenario{
		{
			name:       "Success",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCiliumInstallExpectations(t, client, nil)
			},
		},
		{
			name:       "InstallFailure",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCiliumInstallExpectations(t, client, installertestutils.ErrInstallFailed)
			},
			wantErr: "failed to install Cilium",
		},
		{
			name:       "AddRepositoryFailure",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCiliumAddRepository(t, client, installertestutils.ErrAddRepoFailed)
			},
			wantErr: "failed to add cilium repository",
		},
	}

	runInstallerScenarios(t, scenarios)
}

func TestCiliumInstallerUninstall(t *testing.T) {
	t.Parallel()

	uninstallAction := func(ctx context.Context, installer *CiliumInstaller) error {
		return installer.Uninstall(ctx)
	}

	scenarios := []installerScenario{
		{
			name:       "Success",
			actionName: "Uninstall",
			action:     uninstallAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCiliumUninstall(t, client, nil)
			},
		},
		{
			name:       "UninstallFailure",
			actionName: "Uninstall",
			action:     uninstallAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCiliumUninstall(t, client, installertestutils.ErrUninstallFailed)
			},
			wantErr: "failed to uninstall cilium release",
		},
	}

	runInstallerScenarios(t, scenarios)
}

func TestApplyDefaultValues(t *testing.T) {
	t.Parallel()

	t.Run("ReturnsDefaultValues", func(t *testing.T) {
		t.Parallel()

		vals := applyDefaultValues()

		testutils.ExpectNotNil(t, vals, "default values map")
		installertestutils.ExpectEqual(
			t,
			vals["operator.replicas"],
			"1",
			"operator replicas",
		)
	})
}

func TestCiliumInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	t.Run("InvokesCustomFunction", func(t *testing.T) {
		t.Parallel()

		client := helm.NewMockInterface(t)
		installer := NewCiliumInstaller(client, "kubeconfig", "", time.Second)
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
		installer := NewCiliumInstaller(client, "kubeconfig", "", time.Second)
		defaultFn := installer.GetWaitFn()
		testutils.ExpectNotNil(t, defaultFn, "default wait function")
		defaultPtr := reflect.ValueOf(defaultFn).Pointer()

		installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

		replacedPtr := reflect.ValueOf(installer.GetWaitFn()).Pointer()
		if replacedPtr == defaultPtr {
			t.Fatal("expected custom wait function to replace default")
		}

		installer.SetWaitForReadinessFunc(nil)
		restoredPtr := reflect.ValueOf(installer.GetWaitFn()).Pointer()
		installertestutils.ExpectEqual(
			t,
			restoredPtr,
			defaultPtr,
			"wait function pointer after restore",
		)
	})
}

func TestCiliumInstallerWaitForReadinessBuildConfigError(t *testing.T) {
	t.Parallel()

	installer := NewCiliumInstaller(helm.NewMockInterface(t), "", "", time.Second)
	err := installer.WaitForReadiness(context.Background())

	testutils.ExpectErrorContains(
		t,
		err,
		"build kubernetes client config",
		"WaitForReadiness error path",
	)
}

func TestCiliumInstallerWaitForReadinessNoOpWhenUnset(t *testing.T) {
	t.Parallel()

	installer := NewCiliumInstaller(helm.NewMockInterface(t), "kubeconfig", "", time.Second)
	installer.SetWaitFn(nil)

	err := installer.WaitForReadiness(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when waitFn unset, got %v", err)
	}
}

func TestCiliumInstallerWaitForReadinessSuccess(t *testing.T) {
	t.Parallel()

	server := newCiliumAPIServer(t, true)
	t.Cleanup(server.Close)

	kubeconfig := installertestutils.WriteServerBackedKubeconfig(t, server.URL)

	installer := NewCiliumInstaller(
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

func TestCiliumInstallerWaitForReadinessDetectsUnreadyComponents(t *testing.T) {
	t.Parallel()

	server := newCiliumAPIServer(t, false)
	t.Cleanup(server.Close)

	kubeconfig := installertestutils.WriteServerBackedKubeconfig(t, server.URL)

	installer := NewCiliumInstaller(
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

func TestCiliumInstallerBuildRESTConfig(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWhenKubeconfigMissing", testCiliumBuildRESTConfigErrorWhenKubeconfigMissing)
	t.Run("UsesCurrentContext", testCiliumBuildRESTConfigUsesCurrentContext)
	t.Run("OverridesContext", testCiliumBuildRESTConfigOverridesContext)
	t.Run("MissingContext", testCiliumBuildRESTConfigMissingContext)
}

func testCiliumBuildRESTConfigErrorWhenKubeconfigMissing(t *testing.T) {
	t.Helper()
	t.Parallel()

	installer := NewCiliumInstaller(helm.NewMockInterface(t), "", "", time.Second)
	_, err := installer.buildRESTConfig()

	testutils.ExpectErrorContains(t, err, "kubeconfig path is empty", "buildRESTConfig empty path")
}

func testCiliumBuildRESTConfigUsesCurrentContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := installertestutils.WriteKubeconfig(t, t.TempDir())
	installer := NewCiliumInstaller(helm.NewMockInterface(t), path, "", time.Second)

	restConfig, err := installer.buildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig current context")
	installertestutils.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-one.example.com",
		"rest config host",
	)
}

func testCiliumBuildRESTConfigOverridesContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := installertestutils.WriteKubeconfig(t, t.TempDir())
	installer := NewCiliumInstaller(helm.NewMockInterface(t), path, "alt", time.Second)

	restConfig, err := installer.buildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig override context")
	installertestutils.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-two.example.com",
		"rest config host override",
	)
}

func testCiliumBuildRESTConfigMissingContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := installertestutils.WriteKubeconfig(t, t.TempDir())
	installer := NewCiliumInstaller(helm.NewMockInterface(t), path, "missing", time.Second)
	_, err := installer.buildRESTConfig()

	testutils.ExpectErrorContains(
		t,
		err,
		"context \"missing\" does not exist",
		"buildRESTConfig missing context",
	)
}

func newCiliumAPIServer(t *testing.T, ready bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/apis/apps/v1/namespaces/kube-system/daemonsets/cilium":
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

			installertestutils.EncodeJSON(t, writer, payload)

		case "/apis/apps/v1/namespaces/kube-system/deployments/cilium-operator":
			payload := map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"status": map[string]any{
					"replicas":          1,
					"updatedReplicas":   1,
					"availableReplicas": 1,
				},
			}

			status, ok := payload["status"].(map[string]any)
			if !ok {
				t.Fatalf("unexpected payload status type %T", payload["status"])
			}

			if !ready {
				status["updatedReplicas"] = 0
				status["availableReplicas"] = 0
			}

			installertestutils.EncodeJSON(t, writer, payload)

		default:
			http.NotFound(writer, req)
		}
	}))
}

func newDefaultInstaller(t *testing.T) (*CiliumInstaller, *helm.MockInterface) {
	t.Helper()
	client := helm.NewMockInterface(t)
	installer := NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer, client
}

func setupCiliumInstallExpectations(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()

	expectCiliumAddRepository(t, client, nil)
	expectCiliumInstallChart(t, client, installErr)
}

func expectCiliumAddRepository(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				t.Helper()
				installertestutils.ExpectEqual(t, entry.Name, "cilium", "repository name")
				installertestutils.ExpectEqual(
					t,
					entry.URL,
					"https://helm.cilium.io",
					"repository URL",
				)

				return true
			}),
		).
		Return(err)
}

func expectCiliumInstallChart(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				t.Helper()
				installertestutils.ExpectEqual(t, spec.ReleaseName, "cilium", "release name")
				installertestutils.ExpectEqual(t, spec.ChartName, "cilium/cilium", "chart name")
				installertestutils.ExpectEqual(t, spec.Namespace, "kube-system", "namespace")
				installertestutils.ExpectEqual(
					t,
					spec.RepoURL,
					"https://helm.cilium.io",
					"repository URL",
				)
				testutils.ExpectTrue(t, spec.Wait, "Wait flag")
				testutils.ExpectTrue(t, spec.WaitForJobs, "WaitForJobs flag")
				installertestutils.ExpectEqual(
					t,
					spec.SetJSONVals["operator.replicas"],
					"1",
					"operator replicas",
				)

				return true
			}),
		).
		Return(nil, installErr)
}

func expectCiliumUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "cilium", "kube-system").
		Return(err)
}
