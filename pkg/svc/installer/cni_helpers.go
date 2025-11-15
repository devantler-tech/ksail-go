package installer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/k8s"
	"k8s.io/client-go/rest"
)

// CNIInstallerBase provides common fields and methods for CNI installers.
type CNIInstallerBase struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
	waitFn     func(context.Context) error
}

// NewCNIInstallerBase creates a new base installer instance.
func NewCNIInstallerBase(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
	waitFn func(context.Context) error,
) *CNIInstallerBase {
	return &CNIInstallerBase{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
		waitFn:     waitFn,
	}
}

// WaitForReadiness waits for the CNI components to become ready.
func (b *CNIInstallerBase) WaitForReadiness(ctx context.Context) error {
	if b.waitFn == nil {
		return nil
	}

	err := b.waitFn(ctx)
	if err != nil {
		return fmt.Errorf("wait for readiness: %w", err)
	}

	return nil
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (b *CNIInstallerBase) SetWaitForReadinessFunc(
	waitFunc func(context.Context) error,
	defaultWaitFn func(context.Context) error,
) {
	if waitFunc == nil {
		b.waitFn = defaultWaitFn

		return
	}

	b.waitFn = waitFunc
}

// BuildRESTConfig builds a Kubernetes REST configuration.
func (b *CNIInstallerBase) BuildRESTConfig() (*rest.Config, error) {
	config, err := k8s.BuildRESTConfig(b.kubeconfig, b.context)
	if err != nil {
		return nil, fmt.Errorf("build REST config: %w", err)
	}

	return config, nil
}

var errHelmClientNil = errors.New("helm client is nil")

// GetClient returns the Helm client.
//
//nolint:ireturn // Method returns interface by design for flexibility.
func (b *CNIInstallerBase) GetClient() (helm.Interface, error) {
	if b.client == nil {
		return nil, errHelmClientNil
	}

	return b.client, nil
}

// GetTimeout returns the timeout duration.
func (b *CNIInstallerBase) GetTimeout() time.Duration {
	return b.timeout
}

// GetKubeconfig returns the kubeconfig path.
func (b *CNIInstallerBase) GetKubeconfig() string {
	return b.kubeconfig
}

// GetContext returns the kubeconfig context.
func (b *CNIInstallerBase) GetContext() string {
	return b.context
}

// GetWaitFn returns the wait function for testing purposes.
// This method is primarily used in tests to verify wait function behavior.
func (b *CNIInstallerBase) GetWaitFn() func(context.Context) error {
	return b.waitFn
}

// SetWaitFn sets the wait function directly for testing purposes.
// This is a low-level method used primarily in tests. Prefer using SetWaitForReadinessFunc for production code.
func (b *CNIInstallerBase) SetWaitFn(fn func(context.Context) error) {
	b.waitFn = fn
}
