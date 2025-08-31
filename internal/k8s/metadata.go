// Package k8s provides utilities for working with Kubernetes objects.
package k8s

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewObjectMeta creates a new ObjectMeta with the given name.
func NewObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:                       name,
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