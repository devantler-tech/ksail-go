package flannel //nolint:testpackage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
)

func TestNewFlannelInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	kubeContext := "test-context"
	timeout := 5 * time.Minute

	client := helm.NewMockInterface(t)
	installer := NewFlannelInstaller(client, kubeconfig, kubeContext, timeout)

	testutils.ExpectNotNil(t, installer, "installer instance")
}

func TestFlannelInstallerInstall(t *testing.T) {
	t.Parallel()

	installAction := func(ctx context.Context, installer *Installer) error {
		return installer.Install(ctx)
	}

	scenarios := []helm.InstallerScenario[*Installer]{
		{
			Name:       "Success",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupFlannelInstallExpectations(t, client, nil)
			},
		},
		{
			Name:       "InstallFailure",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				setupFlannelInstallExpectations(t, client, testutils.ErrInstallFailed)
			},
			WantErr: "failed to install Flannel",
		},
		{
			Name:       "AddRepositoryFailure",
			ActionName: "Install",
			Action:     installAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectFlannelAddRepository(t, client, testutils.ErrAddRepoFailed)
			},
			WantErr: "failed to add flannel repository",
		},
	}

	helm.RunInstallerScenarios(t, scenarios, newDefaultInstaller)
}

func TestFlannelInstallerUninstall(t *testing.T) {
	t.Parallel()

	uninstallAction := func(ctx context.Context, installer *Installer) error {
		return installer.Uninstall(ctx)
	}

	scenarios := []helm.InstallerScenario[*Installer]{
		{
			Name:       "Success",
			ActionName: "Uninstall",
			Action:     uninstallAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectFlannelUninstall(t, client, nil)
			},
		},
		{
			Name:       "UninstallFailure",
			ActionName: "Uninstall",
			Action:     uninstallAction,
			Setup: func(t *testing.T, client *helm.MockInterface) {
				t.Helper()

				expectFlannelUninstall(t, client, testutils.ErrUninstallFailed)
			},
			WantErr: "failed to uninstall flannel release",
		},
	}

	helm.RunInstallerScenarios(t, scenarios, newDefaultInstaller)
}

func TestFlannelInstallerSetWaitForReadinessFunc(t *testing.T) {
	t.Parallel()

	helm.TestSetWaitForReadinessFunc(t, func(t *testing.T) *Installer {
		t.Helper()
		client := helm.NewMockInterface(t)

		return NewFlannelInstaller(client, "kubeconfig", "", time.Second)
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

func TestFlannelInstallerWaitForReadyNoOpWhenUnset(t *testing.T) {
	t.Parallel()

	helm.TestWaitForReadinessNoOpWhenUnset(t, func(t *testing.T) *Installer {
		t.Helper()

		return NewFlannelInstaller(helm.NewMockInterface(t), "kubeconfig", "", time.Second)
	})
}

func TestFlannelInstallerWaitForReadinessSuccess(t *testing.T) {
	t.Parallel()

	server := newFlannelAPIServer(t, true)
	t.Cleanup(server.Close)

	kubeconfig := testutils.WriteServerBackedKubeconfig(t, server.URL)

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

func TestFlannelInstallerWaitForReadinessDetectsUnready(t *testing.T) {
	t.Parallel()

	server := newFlannelAPIServer(t, false)
	t.Cleanup(server.Close)

	kubeconfig := testutils.WriteServerBackedKubeconfig(t, server.URL)

	installer := NewFlannelInstaller(
		helm.NewMockInterface(t),
		kubeconfig,
		"default",
		75*time.Millisecond,
	)

	helm.TestWaitForReadinessDetectsUnready(t, installer.waitForReadiness)
}

func newDefaultInstaller(t *testing.T) (*Installer, *helm.MockInterface) {
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

func setupFlannelInstallExpectations(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()

	expectFlannelAddRepository(t, client, nil)
	expectFlannelInstallChart(t, client, installErr)
}

func expectFlannelAddRepository(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	helm.ExpectAddRepository(t, client, helm.RepoExpectation{
		RepoName: flannelRepoName,
		RepoURL:  flannelRepoURL,
	}, err)
}

func expectFlannelInstallChart(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	helm.ExpectInstallChart(t, client, helm.ChartExpectation{
		ReleaseName:     flannelReleaseName,
		ChartName:       flannelChartName,
		Namespace:       flannelNamespace,
		RepoURL:         flannelRepoURL,
		CreateNamespace: true,
	}, installErr)
}

func expectFlannelUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	helm.ExpectUninstall(t, client, flannelReleaseName, flannelNamespace, err)
}

func newFlannelAPIServer(t *testing.T, ready bool) *httptest.Server {
	t.Helper()

	return testutils.NewTestAPIServer(t, func(writer http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/apis/apps/v1/namespaces/kube-flannel/daemonsets/kube-flannel-ds":
			testutils.ServeDaemonSet(t, writer, ready)
		default:
			http.NotFound(writer, req)
		}
	})
}
