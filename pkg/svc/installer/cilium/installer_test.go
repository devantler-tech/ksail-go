package ciliuminstaller //nolint:testpackage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/cnitesthelpers"
	installertestutils "github.com/devantler-tech/ksail-go/pkg/svc/installer/testutils"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
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

func TestCiliumInstallerInstall(t *testing.T) {
	t.Parallel()

	installAction := func(ctx context.Context, installer *CiliumInstaller) error {
		return installer.Install(ctx)
	}

	scenarios := []cnitesthelpers.InstallerScenario[*CiliumInstaller]{
		{
			Name:       "Success",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCiliumInstallExpectations(t, client, nil)
			},
		},
		{
			Name:       "InstallFailure",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCiliumInstallExpectations(t, client, installertestutils.ErrInstallFailed)
			},
			WantErr: "failed to install Cilium",
		},
		{
			Name:       "AddRepositoryFailure",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCiliumAddRepository(t, client, installertestutils.ErrAddRepoFailed)
			},
			WantErr: "failed to add cilium repository",
		},
	}

	cnitesthelpers.RunInstallerScenarios(t, scenarios, newDefaultInstaller)
}

func TestCiliumInstallerUninstall(t *testing.T) {
	t.Parallel()

	uninstallAction := func(ctx context.Context, installer *CiliumInstaller) error {
		return installer.Uninstall(ctx)
	}

	scenarios := []cnitesthelpers.InstallerScenario[*CiliumInstaller]{
		{
			Name:       "Success",
			ActionName: "Uninstall",
			Action:     uninstallAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCiliumUninstall(t, client, nil)
			},
		},
		{
			Name:       "UninstallFailure",
			ActionName: "Uninstall",
			Action:     uninstallAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCiliumUninstall(t, client, installertestutils.ErrUninstallFailed)
			},
			WantErr: "failed to uninstall cilium release",
		},
	}

	cnitesthelpers.RunInstallerScenarios(t, scenarios, newDefaultInstaller)
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

	cnitesthelpers.TestSetWaitForReadinessFunc(t, func(t *testing.T) *CiliumInstaller {
		client := helm.NewMockInterface(t)
		return NewCiliumInstaller(client, "kubeconfig", "", time.Second)
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

	cnitesthelpers.TestWaitForReadinessNoOpWhenUnset(t, func(t *testing.T) *CiliumInstaller {
		return NewCiliumInstaller(helm.NewMockInterface(t), "kubeconfig", "", time.Second)
	})
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

	cnitesthelpers.TestWaitForReadinessDetectsUnready(t, server.URL, installer.waitForReadiness)
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
	cnitesthelpers.ExpectAddRepository(t, client, cnitesthelpers.HelmRepoExpectation{
		RepoName: "cilium",
		RepoURL:  "https://helm.cilium.io",
	}, err)
}

func expectCiliumInstallChart(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	cnitesthelpers.ExpectInstallChart(t, client, cnitesthelpers.HelmChartExpectation{
		ReleaseName:     "cilium",
		ChartName:       "cilium/cilium",
		Namespace:       "kube-system",
		RepoURL:         "https://helm.cilium.io",
		CreateNamespace: false,
		SetJSONVals:     map[string]string{"operator.replicas": "1"},
	}, installErr)
}

func expectCiliumUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	cnitesthelpers.ExpectUninstall(t, client, "cilium", "kube-system", err)
}
