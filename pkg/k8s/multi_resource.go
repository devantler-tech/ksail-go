package k8s

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
)

// ErrTimeoutExceeded is returned when a timeout is exceeded.
var ErrTimeoutExceeded = errors.New("timeout exceeded")

var errUnknownResourceType = errors.New("unknown resource type")

// ReadinessCheck defines a check to perform for a Kubernetes resource.
type ReadinessCheck struct {
	// Type specifies the kind of Kubernetes resource to check for readiness.
	// Valid values are "deployment" or "daemonset".
	Type string
	// Namespace is the Kubernetes namespace where the resource resides.
	Namespace string
	// Name is the name of the resource to check for readiness.
	Name string
}

// WaitForMultipleResources waits for multiple Kubernetes resources to be ready.
func WaitForMultipleResources(
	ctx context.Context,
	clientset kubernetes.Interface,
	checks []ReadinessCheck,
	timeout time.Duration,
) error {
	start := time.Now()

	for _, check := range checks {
		elapsed := time.Since(start)

		remainingTimeout := timeout - elapsed
		if remainingTimeout <= 0 {
			return errTimeoutExceeded(check.Type, check.Namespace, check.Name)
		}

		resourceCtx, cancel := context.WithTimeout(ctx, remainingTimeout)
		defer cancel()

		var err error

		switch check.Type {
		case "deployment":
			err = WaitForDeploymentReady(
				resourceCtx, clientset, check.Namespace, check.Name, remainingTimeout,
			)
			if err != nil {
				return fmt.Errorf("%s deployment not ready: %w", check.Name, err)
			}
		case "daemonset":
			err = WaitForDaemonSetReady(
				resourceCtx, clientset, check.Namespace, check.Name, remainingTimeout,
			)
			if err != nil {
				return fmt.Errorf("%s daemonset not ready: %w", check.Name, err)
			}
		default:
			return fmt.Errorf("%w: %s", errUnknownResourceType, check.Type)
		}
	}

	return nil
}

func errTimeoutExceeded(resourceType, namespace, name string) error {
	return fmt.Errorf(
		"%w before checking %s %s/%s",
		ErrTimeoutExceeded, resourceType, namespace, name,
	)
}
