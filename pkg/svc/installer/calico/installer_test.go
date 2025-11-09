package calicoinstaller //nolint:testpackage

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

func TestNewCalicoInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := helm.NewMockInterface(t)
	installer := NewCalicoInstaller(client, kubeconfig, context, timeout)

	testutils.ExpectNotNil(t, installer, "installer instance")
}

type installerScenario struct {
	name       string
	setup      func(*testing.T, *helm.MockInterface)
	actionName string
	action     func(context.Context, *CalicoInstaller) error
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

func TestCalicoInstallerInstall(t *testing.T) {
	t.Parallel()

	installAction := func(ctx context.Context, installer *CalicoInstaller) error {
		return installer.Install(ctx)
	}

	scenarios := []installerScenario{
		{
			name:       "Success",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCalicoInstallExpectations(t, client, nil)
			},
		},
		{
			name:       "InstallFailure",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCalicoInstallExpectations(t, client, installertestutils.ErrInstallFailed)
			},
			wantErr: "failed to install Calico",
		},
		{
			name:       "AddRepositoryFailure",
			actionName: "Install",
			action:     installAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCalicoAddRepository(t, client, installertestutils.ErrAddRepoFailed)
			},
			wantErr: "failed to add calico repository",
		},
	}

	runInstallerScenarios(t, scenarios)
}

func TestCalicoInstallerUninstall(t *testing.T) {
	t.Parallel()

	uninstallAction := func(ctx context.Context, installer *CalicoInstaller) error {
		return installer.Uninstall(ctx)
	}

	scenarios := []installerScenario{
		{
			name:       "Success",
			actionName: "Uninstall",
			action:     uninstallAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCalicoUninstall(t, client, nil)
			},
		},
		{
			name:       "UninstallFailure",
			actionName: "Uninstall",
			action:     uninstallAction,
			setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCalicoUninstall(t, client, installertestutils.ErrUninstallFailed)
			},
			wantErr: "failed to uninstall calico release",
		},
	}

	runInstallerScenarios(t, scenarios)
}

func TestCalicoInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	t.Run("InvokesCustomFunction", func(t *testing.T) {
		t.Parallel()

		client := helm.NewMockInterface(t)
		installer := NewCalicoInstaller(client, "kubeconfig", "", time.Second)
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
		installer := NewCalicoInstaller(client, "kubeconfig", "", time.Second)
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

func TestCalicoInstallerWaitForReadinessBuildConfigError(t *testing.T) {
	t.Parallel()

	installer := NewCalicoInstaller(helm.NewMockInterface(t), "", "", time.Second)
	err := installer.WaitForReadiness(context.Background())

	testutils.ExpectErrorContains(
		t,
		err,
		"build kubernetes client config",
		"WaitForReadiness error path",
	)
}

func TestCalicoInstallerWaitForReadinessNoOpWhenUnset(t *testing.T) {
	t.Parallel()

	installer := NewCalicoInstaller(helm.NewMockInterface(t), "kubeconfig", "", time.Second)
	installer.SetWaitFn(nil)

	err := installer.WaitForReadiness(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when waitFn unset, got %v", err)
	}
}

func TestCalicoInstallerWaitForReadinessSuccess(t *testing.T) {
	t.Parallel()

	server := newCalicoAPIServer(t, true)
	t.Cleanup(server.Close)

	kubeconfig := installertestutils.WriteServerBackedKubeconfig(t, server.URL)

	installer := NewCalicoInstaller(
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

func TestCalicoInstallerWaitForReadinessDetectsUnreadyComponents(t *testing.T) {
	t.Parallel()

	server := newCalicoAPIServer(t, false)
	t.Cleanup(server.Close)

	kubeconfig := installertestutils.WriteServerBackedKubeconfig(t, server.URL)

	installer := NewCalicoInstaller(
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

func TestCalicoInstallerBuildRESTConfig(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWhenKubeconfigMissing", testBuildRESTConfigErrorWhenKubeconfigMissing)
	t.Run("UsesCurrentContext", testBuildRESTConfigUsesCurrentContext)
	t.Run("OverridesContext", testBuildRESTConfigOverridesContext)
	t.Run("MissingContext", testBuildRESTConfigMissingContext)
}

func testBuildRESTConfigErrorWhenKubeconfigMissing(t *testing.T) {
	t.Helper()
	t.Parallel()

	installer := NewCalicoInstaller(helm.NewMockInterface(t), "", "", time.Second)
	_, err := installer.buildRESTConfig()

	testutils.ExpectErrorContains(t, err, "kubeconfig path is empty", "buildRESTConfig empty path")
}

func testBuildRESTConfigUsesCurrentContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := installertestutils.WriteKubeconfig(t, t.TempDir())
	installer := NewCalicoInstaller(helm.NewMockInterface(t), path, "", time.Second)

	restConfig, err := installer.buildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig current context")
	installertestutils.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-one.example.com",
		"rest config host",
	)
}

func testBuildRESTConfigOverridesContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := installertestutils.WriteKubeconfig(t, t.TempDir())
	installer := NewCalicoInstaller(helm.NewMockInterface(t), path, "alt", time.Second)

	restConfig, err := installer.buildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig override context")
	installertestutils.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-two.example.com",
		"rest config host override",
	)
}

func testBuildRESTConfigMissingContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := installertestutils.WriteKubeconfig(t, t.TempDir())
	installer := NewCalicoInstaller(helm.NewMockInterface(t), path, "missing", time.Second)
	_, err := installer.buildRESTConfig()

	testutils.ExpectErrorContains(
		t,
		err,
		"context \"missing\" does not exist",
		"buildRESTConfig missing context",
	)
}

func newCalicoAPIServer(t *testing.T, ready bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/apis/apps/v1/namespaces/tigera-operator/deployments/tigera-operator":
			serveTigeraOperatorDeployment(t, writer, ready)
		case "/apis/apps/v1/namespaces/calico-system/daemonsets/calico-node":
			serveCalicoNodeDaemonSet(t, writer, ready)
		case "/apis/apps/v1/namespaces/calico-system/deployments/calico-kube-controllers":
			serveCalicoKubeControllersDeployment(t, writer, ready)
		default:
			http.NotFound(writer, req)
		}
	}))
}

func serveTigeraOperatorDeployment(t *testing.T, writer http.ResponseWriter, ready bool) {
	t.Helper()

	payload := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"status": map[string]any{
			"replicas":          1,
			"updatedReplicas":   1,
			"availableReplicas": 1,
		},
	}

	if !ready {
		installertestutils.UpdateDeploymentStatusToUnready(t, payload)
	}

	installertestutils.EncodeJSON(t, writer, payload)
}

func serveCalicoNodeDaemonSet(t *testing.T, writer http.ResponseWriter, ready bool) {
	t.Helper()

	payload := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "DaemonSet",
		"status": map[string]any{
			"desiredNumberScheduled": 1,
			"numberUnavailable":      0,
			"updatedNumberScheduled": 1,
		},
	}

	if !ready {
		installertestutils.UpdateDaemonSetStatusToUnready(t, payload)
	}

	installertestutils.EncodeJSON(t, writer, payload)
}

func serveCalicoKubeControllersDeployment(t *testing.T, writer http.ResponseWriter, ready bool) {
	t.Helper()

	payload := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"status": map[string]any{
			"replicas":          1,
			"updatedReplicas":   1,
			"availableReplicas": 1,
		},
	}

	if !ready {
		installertestutils.UpdateDeploymentStatusToUnready(t, payload)
	}

	installertestutils.EncodeJSON(t, writer, payload)
}

func newDefaultInstaller(t *testing.T) (*CalicoInstaller, *helm.MockInterface) {
	t.Helper()
	client := helm.NewMockInterface(t)
	installer := NewCalicoInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer, client
}

func setupCalicoInstallExpectations(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()

	expectCalicoAddRepository(t, client, nil)
	expectCalicoInstallChart(t, client, installErr)
}

func expectCalicoAddRepository(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				t.Helper()
				installertestutils.ExpectEqual(t, entry.Name, "projectcalico", "repository name")
				installertestutils.ExpectEqual(
					t,
					entry.URL,
					"https://docs.tigera.io/calico/charts",
					"repository URL",
				)

				return true
			}),
		).
		Return(err)
}

func expectCalicoInstallChart(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				t.Helper()
				installertestutils.ExpectEqual(t, spec.ReleaseName, "calico", "release name")
				installertestutils.ExpectEqual(
					t,
					spec.ChartName,
					"projectcalico/tigera-operator",
					"chart name",
				)
				installertestutils.ExpectEqual(t, spec.Namespace, "tigera-operator", "namespace")
				installertestutils.ExpectEqual(
					t,
					spec.RepoURL,
					"https://docs.tigera.io/calico/charts",
					"repository URL",
				)
				testutils.ExpectTrue(t, spec.Wait, "Wait flag")
				testutils.ExpectTrue(t, spec.WaitForJobs, "WaitForJobs flag")
				testutils.ExpectTrue(t, spec.CreateNamespace, "CreateNamespace flag")

				return true
			}),
		).
		Return(nil, installErr)
}

func expectCalicoUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "calico", "tigera-operator").
		Return(err)
}
