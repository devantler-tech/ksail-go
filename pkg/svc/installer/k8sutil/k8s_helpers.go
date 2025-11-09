// Package k8sutil provides shared Kubernetes utilities for CNI installers.
package k8sutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	readinessPollInterval = 2 * time.Second
)

// ErrKubeconfigPathEmpty is returned when kubeconfig path is empty.
var ErrKubeconfigPathEmpty = errors.New("kubeconfig path is empty")

// BuildRESTConfig builds a Kubernetes REST config from kubeconfig path and optional context.
func BuildRESTConfig(kubeconfig, context string) (*rest.Config, error) {
	if kubeconfig == "" {
		return nil, ErrKubeconfigPathEmpty
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}

	overrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		overrides.CurrentContext = context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	return restConfig, nil
}

// WaitForDaemonSetReady waits for a DaemonSet to be ready.
func WaitForDaemonSetReady(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace, name string,
	deadline time.Duration,
) error {
	return PollForReadiness(ctx, deadline, func(ctx context.Context) (bool, error) {
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

// WaitForDeploymentReady waits for a Deployment to be ready.
func WaitForDeploymentReady(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace, name string,
	deadline time.Duration,
) error {
	return PollForReadiness(ctx, deadline, func(ctx context.Context) (bool, error) {
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

// PollForReadiness polls a check function until ready or timeout.
func PollForReadiness(
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
