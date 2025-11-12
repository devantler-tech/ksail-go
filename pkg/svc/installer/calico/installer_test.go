package calicoinstaller //nolint:testpackage

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

func TestNewCalicoInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := helm.NewMockInterface(t)
	installer := NewCalicoInstaller(client, kubeconfig, context, timeout)

	testutils.ExpectNotNil(t, installer, "installer instance")
}

func TestCalicoInstallerInstall(t *testing.T) {
	t.Parallel()

	installAction := func(ctx context.Context, installer *CalicoInstaller) error {
		return installer.Install(ctx)
	}

	scenarios := []cnitesthelpers.InstallerScenario[*CalicoInstaller]{
		{
			Name:       "Success",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCalicoInstallExpectations(t, client, nil)
			},
		},
		{
			Name:       "InstallFailure",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupCalicoInstallExpectations(t, client, installertestutils.ErrInstallFailed)
			},
			WantErr: "failed to install Calico",
		},
		{
			Name:       "AddRepositoryFailure",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCalicoAddRepository(t, client, installertestutils.ErrAddRepoFailed)
			},
			WantErr: "failed to add calico repository",
		},
	}

	cnitesthelpers.RunInstallerScenarios(t, scenarios, newDefaultInstaller)
}

func TestCalicoInstallerUninstall(t *testing.T) {
	t.Parallel()

	uninstallAction := func(ctx context.Context, installer *CalicoInstaller) error {
		return installer.Uninstall(ctx)
	}

	scenarios := []cnitesthelpers.InstallerScenario[*CalicoInstaller]{
		{
			Name:       "Success",
			ActionName: "Uninstall",
			Action:     uninstallAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCalicoUninstall(t, client, nil)
			},
		},
		{
			Name:       "UninstallFailure",
			ActionName: "Uninstall",
			Action:     uninstallAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectCalicoUninstall(t, client, installertestutils.ErrUninstallFailed)
			},
			WantErr: "failed to uninstall calico release",
		},
	}

	cnitesthelpers.RunInstallerScenarios(t, scenarios, newDefaultInstaller)
}

func TestCalicoInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	cnitesthelpers.TestSetWaitForReadinessFunc(t, func(t *testing.T) *CalicoInstaller {
		client := helm.NewMockInterface(t)
		return NewCalicoInstaller(client, "kubeconfig", "", time.Second)
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

	cnitesthelpers.TestWaitForReadinessNoOpWhenUnset(t, func(t *testing.T) *CalicoInstaller {
		return NewCalicoInstaller(helm.NewMockInterface(t), "kubeconfig", "", time.Second)
	})
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

	cnitesthelpers.TestWaitForReadinessDetectsUnready(t, server.URL, installer.waitForReadiness)
}

func newCalicoAPIServer(t *testing.T, ready bool) *httptest.Server {
	t.Helper()

	return cnitesthelpers.NewTestAPIServer(t, func(writer http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/apis/apps/v1/namespaces/tigera-operator/deployments/tigera-operator":
			cnitesthelpers.ServeDeployment(t, writer, ready)
		case "/apis/apps/v1/namespaces/calico-system/daemonsets/calico-node":
			cnitesthelpers.ServeDaemonSet(t, writer, ready)
		case "/apis/apps/v1/namespaces/calico-system/deployments/calico-kube-controllers":
			cnitesthelpers.ServeDeployment(t, writer, ready)
		default:
			http.NotFound(writer, req)
		}
	})
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
	cnitesthelpers.ExpectAddRepository(t, client, cnitesthelpers.HelmRepoExpectation{
		RepoName: "projectcalico",
		RepoURL:  "https://docs.tigera.io/calico/charts",
	}, err)
}

func expectCalicoInstallChart(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	cnitesthelpers.ExpectInstallChart(t, client, cnitesthelpers.HelmChartExpectation{
		ReleaseName:     "calico",
		ChartName:       "projectcalico/tigera-operator",
		Namespace:       "tigera-operator",
		RepoURL:         "https://docs.tigera.io/calico/charts",
		CreateNamespace: true,
	}, installErr)
}

func expectCalicoUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	cnitesthelpers.ExpectUninstall(t, client, "calico", "tigera-operator", err)
}
