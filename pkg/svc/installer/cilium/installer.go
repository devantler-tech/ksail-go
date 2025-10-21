package ciliuminstaller

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	commandrunner "github.com/devantler-tech/ksail-go/pkg/svc/commandrunner"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// CiliumInstaller implements the installer.Installer interface for Cilium.
type CiliumInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
	runner     commandrunner.CommandRunner
	waitFn     func(context.Context) error
	stdout     io.Writer
	stderr     io.Writer
}

var errKubeconfigPathEmpty = stderrors.New("ciliuminstaller: kubeconfig path is empty")

const readinessPollInterval = 3 * time.Second

// NewCiliumInstaller creates a new Cilium installer instance.
func NewCiliumInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *CiliumInstaller {
	return NewCiliumInstallerWithRunner(
		client,
		kubeconfig,
		context,
		timeout,
		commandrunner.NewCobraCommandRunner(os.Stdout, os.Stderr),
		os.Stdout,
		os.Stderr,
	)
}

// NewCiliumInstallerWithRunner creates a new Cilium installer instance with a custom command runner.
// This is primarily used for testing.
func NewCiliumInstallerWithRunner(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
	runner commandrunner.CommandRunner,
	stdout, stderr io.Writer,
) *CiliumInstaller {
	installer := &CiliumInstaller{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
		runner:     runner,
		stdout:     stdout,
		stderr:     stderr,
	}

	installer.waitFn = installer.waitForReadiness

	return installer
}

// Install installs or upgrades Cilium via its Helm chart using in-process Cobra commands.
func (c *CiliumInstaller) Install(ctx context.Context) error {
	//nolint:contextcheck // Context is passed via cobra command runner
	cmd := NewInstallCommand(
		c.client,
		c.kubeconfig,
		c.context,
		c.timeout,
		c.stdout,
	)

	result, err := c.runner.Run(ctx, cmd, []string{})
	if err != nil {
		mergedErr := commandrunner.MergeCommandError(
			fmt.Errorf("failed to install Cilium: %w", err),
			result,
		)

		return fmt.Errorf("install command failed: %w", mergedErr)
	}

	return nil
}

// WaitForReadiness waits for the Cilium components to become ready.
func (c *CiliumInstaller) WaitForReadiness(ctx context.Context) error {
	if c.waitFn == nil {
		return nil
	}

	return c.waitFn(ctx)
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (c *CiliumInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	if waitFunc == nil {
		c.waitFn = c.waitForReadiness

		return
	}

	c.waitFn = waitFunc
}

// Uninstall removes the Helm release for Cilium using in-process Cobra commands.
func (c *CiliumInstaller) Uninstall(ctx context.Context) error {
	//nolint:contextcheck // Context is passed via cobra command runner
	cmd := NewUninstallCommand(
		c.client,
		c.kubeconfig,
		c.context,
		c.stdout,
	)

	result, err := c.runner.Run(ctx, cmd, []string{})
	if err != nil {
		mergedErr := commandrunner.MergeCommandError(
			fmt.Errorf("failed to uninstall Cilium: %w", err),
			result,
		)

		return fmt.Errorf("uninstall command failed: %w", mergedErr)
	}

	return nil
}

// --- internals ---

func (c *CiliumInstaller) waitForReadiness(ctx context.Context) error {
	restConfig, err := c.buildRESTConfig()
	if err != nil {
		return fmt.Errorf("build kubernetes client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	daemonSetErr := waitForDaemonSetReady(waitCtx, clientset, "kube-system", "cilium", c.timeout)
	if daemonSetErr != nil {
		return fmt.Errorf("cilium daemonset not ready: %w", daemonSetErr)
	}

	deploymentErr := waitForDeploymentReady(
		waitCtx,
		clientset,
		"kube-system",
		"cilium-operator",
		c.timeout,
	)
	if deploymentErr != nil {
		return fmt.Errorf("cilium operator not ready: %w", deploymentErr)
	}

	return nil
}

func (c *CiliumInstaller) buildRESTConfig() (*rest.Config, error) {
	if c.kubeconfig == "" {
		return nil, errKubeconfigPathEmpty
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: c.kubeconfig}

	overrides := &clientcmd.ConfigOverrides{}
	if c.context != "" {
		overrides.CurrentContext = c.context
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

//nolint:unparam // name remains for future flexibility and consistent signatures.
func waitForDeploymentReady(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace, name string,
	deadline time.Duration,
) error {
	return pollForReadiness(ctx, deadline, func(ctx context.Context) (bool, error) {
		deployment, err := clientset.AppsV1().
			Deployments(namespace).
			Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return false, fmt.Errorf("get deployment %s/%s: %w", namespace, name, err)
		}

		if deployment.Status.Replicas == 0 {
			return false, nil
		}

		if deployment.Status.UpdatedReplicas < deployment.Status.Replicas {
			return false, nil
		}

		if deployment.Status.AvailableReplicas < deployment.Status.Replicas {
			return false, nil
		}

		return true, nil
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
