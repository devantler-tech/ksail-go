package flannelinstaller

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// FlannelInstaller implements the installer.Installer interface for Flannel.
type FlannelInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
	waitFn     func(context.Context) error
}

var errKubeconfigPathEmpty = stderrors.New("flannelinstaller: kubeconfig path is empty")

const readinessPollInterval = 3 * time.Second

// NewFlannelInstaller creates a new Flannel installer instance.
func NewFlannelInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *FlannelInstaller {
	installer := &FlannelInstaller{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}

	installer.waitFn = installer.waitForReadiness

	return installer
}

// Install installs or upgrades Flannel via its Helm chart.
func (f *FlannelInstaller) Install(ctx context.Context) error {
	err := f.helmInstallOrUpgradeFlannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Flannel: %w", err)
	}

	return nil
}

// WaitForReadiness waits for the Flannel components to become ready.
func (f *FlannelInstaller) WaitForReadiness(ctx context.Context) error {
	if f.waitFn == nil {
		return nil
	}

	return f.waitFn(ctx)
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (f *FlannelInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	if waitFunc == nil {
		f.waitFn = f.waitForReadiness

		return
	}

	f.waitFn = waitFunc
}

// Uninstall removes the Helm release for Flannel.
func (f *FlannelInstaller) Uninstall(ctx context.Context) error {
	err := f.client.UninstallRelease(ctx, "flannel", "kube-flannel")
	if err != nil {
		return fmt.Errorf("failed to uninstall flannel release: %w", err)
	}

	return nil
}

// --- internals ---

func (f *FlannelInstaller) helmInstallOrUpgradeFlannel(ctx context.Context) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "flannel",
		URL:  "https://flannel-io.github.io/flannel",
	}

	addRepoErr := f.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add flannel repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     "flannel",
		ChartName:       "flannel/flannel",
		Namespace:       "kube-flannel",
		RepoURL:         "https://flannel-io.github.io/flannel",
		CreateNamespace: true,
		Atomic:          true,
		Silent:          true,
		UpgradeCRDs:     true,
		Timeout:         f.timeout,
		Wait:            true,
		WaitForJobs:     true,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	_, err := f.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install flannel chart: %w", err)
	}

	return nil
}

func (f *FlannelInstaller) waitForReadiness(ctx context.Context) error {
	restConfig, err := f.buildRESTConfig()
	if err != nil {
		return fmt.Errorf("build kubernetes client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	daemonSetErr := waitForDaemonSetReady(
		waitCtx,
		clientset,
		"kube-flannel",
		"kube-flannel-ds",
		f.timeout,
	)
	if daemonSetErr != nil {
		return fmt.Errorf("flannel daemonset not ready: %w", daemonSetErr)
	}

	return nil
}

func (f *FlannelInstaller) buildRESTConfig() (*rest.Config, error) {
	if f.kubeconfig == "" {
		return nil, errKubeconfigPathEmpty
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: f.kubeconfig}

	overrides := &clientcmd.ConfigOverrides{}
	if f.context != "" {
		overrides.CurrentContext = f.context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	return restConfig, nil
}

func waitForDaemonSetReady(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace, name string,
	deadline time.Duration,
) error {
	return pollForReadiness(ctx, deadline, func(ctx context.Context) (bool, error) {
		daemonSet, err := clientset.AppsV1().
			DaemonSets(namespace).
			Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return false, fmt.Errorf("get daemonset %s/%s: %w", namespace, name, err)
		}

		if daemonSet.Status.DesiredNumberScheduled == 0 {
			return false, nil
		}

		ready := daemonSet.Status.NumberUnavailable == 0 &&
			daemonSet.Status.UpdatedNumberScheduled == daemonSet.Status.DesiredNumberScheduled

		return ready, nil
	})
}

func pollForReadiness(
	ctx context.Context,
	deadline time.Duration,
	poll func(context.Context) (bool, error),
) error {
	pollErr := wait.PollUntilContextTimeout(
		ctx,
		readinessPollInterval,
		deadline,
		true,
		poll,
	)
	if pollErr != nil {
		return fmt.Errorf("poll for readiness: %w", pollErr)
	}

	return nil
}
