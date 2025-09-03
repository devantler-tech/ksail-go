// Package kubectlinstaller provides interfaces for kubectl installer dependencies.
package kubectlinstaller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DynamicClient defines the interface for dynamic client operations.
type DynamicClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error)
	Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}