package flannel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	"github.com/devantler-tech/ksail-go/pkg/k8s"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// flannelManifestURL is the URL to the latest Flannel CNI manifest.
	flannelManifestURL = "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml"

	// flannelNamespace is the Kubernetes namespace where Flannel is deployed.
	flannelNamespace = "kube-flannel"

	// flannelDaemonSetName is the name of the Flannel DaemonSet.
	flannelDaemonSetName = "kube-flannel-ds"
)

// Installer implements the installer.Installer interface for Flannel CNI.
// Unlike other CNI installers, it uses kubectl directly instead of Helm.
type Installer struct {
	kubectlClient kubectl.Interface
	kubeconfig    string
	context       string
	timeout       time.Duration
	waitFn        func(context.Context) error
}

var _ installer.Installer = (*Installer)(nil)

// NewFlannelInstaller creates a new Flannel CNI installer.
//
// Parameters:
//   - client: kubectl client interface for applying manifests (must not be nil)
//   - kubeconfig: path to kubeconfig file
//   - context: Kubernetes context name
//   - timeout: maximum wait duration for readiness checks
//
// Panics if client is nil (defensive programming).
func NewFlannelInstaller(
	client kubectl.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *Installer {
	if client == nil {
		panic("kubectl client cannot be nil")
	}

	installer := &Installer{
		kubectlClient: client,
		kubeconfig:    kubeconfig,
		context:       context,
		timeout:       timeout,
	}

	// Set default readiness function
	installer.waitFn = installer.waitForReadiness

	return installer
}

// Install installs Flannel CNI by applying the official manifest and waiting for readiness.
//
// The method performs the following steps:
//  1. Applies the Flannel manifest from the official GitHub URL
//  2. Waits for the Flannel DaemonSet to become ready on all nodes
//
// Returns nil on success, or an error if installation fails.
func (f *Installer) Install(ctx context.Context) error {
	err := f.kubectlClient.Apply(ctx, flannelManifestURL)
	if err != nil {
		return f.wrapApplyError(err)
	}

	if f.waitFn != nil {
		err = f.waitFn(ctx)
		if err != nil {
			return f.wrapReadinessError(err)
		}
	}

	return nil
}

// WaitForReadiness waits for Flannel to become ready.
// This method delegates to the internal waitForReadiness function.
func (f *Installer) WaitForReadiness(ctx context.Context) error {
	if f.waitFn != nil {
		return f.waitFn(ctx)
	}

	return f.waitForReadiness(ctx)
}

// Uninstall removes Flannel CNI components from the cluster.
//
// The method deletes:
//   - Flannel DaemonSet
//   - Flannel namespace (which cascades deletion of other resources)
//
// The operation is idempotent - no error if resources don't exist.
// Returns nil on success, or an error if deletion fails.
func (f *Installer) Uninstall(ctx context.Context) error {
	err := f.kubectlClient.Delete(ctx, flannelNamespace, "daemonset", flannelDaemonSetName)
	if err != nil {
		return fmt.Errorf("failed to delete Flannel DaemonSet: %w", err)
	}

	err = f.kubectlClient.Delete(ctx, "", "namespace", flannelNamespace)
	if err != nil {
		return fmt.Errorf("failed to delete Flannel namespace: %w", err)
	}

	return nil
}

// SetWaitForReadinessFunc allows overriding the readiness check function.
// Primarily used for testing to mock readiness checks.
//
// If waitFunc is nil, this restores the default readiness check.
func (f *Installer) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	if waitFunc == nil {
		waitFunc = f.waitForReadiness
	}

	f.waitFn = waitFunc
}

// waitForReadiness waits for the Flannel DaemonSet to become ready on all nodes.
//
// Readiness criteria:
//   - DaemonSet status.desiredNumberScheduled == status.numberReady
//   - DaemonSet status.numberReady == status.numberAvailable
//   - All pods have passed readiness probes
//
// Returns nil when ready, or error on timeout/failure.
func (f *Installer) waitForReadiness(ctx context.Context) error {
	// Build REST config from kubeconfig
	config, err := k8s.BuildRESTConfig(f.kubeconfig, f.context)
	if err != nil {
		return fmt.Errorf("build REST config: %w", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	// Wait for DaemonSet to be ready
	err = k8s.WaitForDaemonSetReady(
		ctx,
		clientset,
		flannelNamespace,
		flannelDaemonSetName,
		f.timeout,
	)
	if err != nil {
		return fmt.Errorf("flannel daemonset not ready: %w", err)
	}

	return nil
}

func (f *Installer) wrapApplyError(err error) error {
	if err == nil {
		return nil
	}

	var (
		urlErr *url.Error
		netErr net.Error
	)

	switch {
	case errors.As(err, &urlErr), errors.As(err, &netErr):
		return fmt.Errorf(
			"flannel manifest download failed: verify network access to %s: %w",
			flannelManifestURL,
			err,
		)
	case strings.Contains(err.Error(), "unexpected status code"):
		return fmt.Errorf(
			"flannel manifest URL %s returned an unexpected response; ensure the URL is reachable and valid: %w",
			flannelManifestURL,
			err,
		)
	case strings.Contains(err.Error(), "failed to decode manifest"):
		return fmt.Errorf(
			"flannel manifest from %s could not be decoded; verify the download is a valid Kubernetes manifest: %w",
			flannelManifestURL,
			err,
		)
	case apierrors.IsUnauthorized(err):
		return fmt.Errorf(
			"flannel installation failed: authentication to the cluster was rejected; verify kubeconfig credentials: %w",
			err,
		)
	case apierrors.IsForbidden(err):
		return fmt.Errorf(
			"flannel installation failed: insufficient RBAC permissions; cluster-admin privileges are required: %w",
			err,
		)
	case apierrors.IsInvalid(err):
		return fmt.Errorf(
			"flannel manifest was rejected by the API server; validate resources fetched from %s: %w",
			flannelManifestURL,
			err,
		)
	default:
		return fmt.Errorf("failed to apply Flannel manifest: %w", err)
	}
}

func (f *Installer) wrapReadinessError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded), wait.Interrupted(err):
		return fmt.Errorf(
			"flannel readiness timed out after %s; inspect pods in namespace %q to diagnose: %w",
			f.timeout,
			flannelNamespace,
			err,
		)
	default:
		return fmt.Errorf("flannel readiness check failed: %w", err)
	}
}
