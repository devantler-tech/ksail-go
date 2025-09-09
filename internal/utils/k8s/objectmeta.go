// Package k8s provides utilities for working with Kubernetes API objects.
package k8s

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewEmptyObjectMeta creates a metav1.ObjectMeta with all fields set to empty/nil values.
// This is useful for creating default ObjectMeta structures in constructors and tests.
func NewEmptyObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:                       "",
		GenerateName:               "",
		Namespace:                  "",
		SelfLink:                   "",
		UID:                        "",
		ResourceVersion:            "",
		Generation:                 0,
		CreationTimestamp:          metav1.Time{Time: time.Time{}},
		DeletionTimestamp:          nil,
		DeletionGracePeriodSeconds: nil,
		Labels:                     nil,
		Annotations:                nil,
		OwnerReferences:            nil,
		Finalizers:                 nil,
		ManagedFields:              nil,
	}
}
