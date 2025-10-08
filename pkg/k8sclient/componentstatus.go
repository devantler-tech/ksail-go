package k8sclient

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ComponentStatusProvider describes a provider that can retrieve component statuses.
type ComponentStatusProvider interface {
	// GetComponentStatuses retrieves all component statuses from the cluster.
	GetComponentStatuses(
		ctx context.Context,
		clientset *kubernetes.Clientset,
	) ([]corev1.ComponentStatus, error)
}

// DefaultComponentStatusProvider is the default implementation of ComponentStatusProvider.
type DefaultComponentStatusProvider struct{}

// NewDefaultComponentStatusProvider creates a new DefaultComponentStatusProvider.
func NewDefaultComponentStatusProvider() *DefaultComponentStatusProvider {
	return &DefaultComponentStatusProvider{}
}

// GetComponentStatuses retrieves all component statuses from the cluster.
func (p *DefaultComponentStatusProvider) GetComponentStatuses(
	ctx context.Context,
	clientset *kubernetes.Clientset,
) ([]corev1.ComponentStatus, error) {
	componentStatuses, err := clientset.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list component statuses: %w", err)
	}

	return componentStatuses.Items, nil
}
