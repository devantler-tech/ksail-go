package configmanager

import (
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

// FieldSelector defines a field and its metadata for configuration management.
type FieldSelector[T any] struct {
	Selector     func(*T) any // Function that returns a pointer to the field
	Description  string       // Human-readable description for CLI flags
	DefaultValue any          // Default value for the field
}

// AddFlagFromField returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
// Requires a default value as the second parameter, optionally accepts a description as the third parameter.
//
// Usage:
//
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, v1alpha1.DistributionKind)
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//		v1alpha1.DistributionKind, "Custom description")
func AddFlagFromField(
	selector func(*v1alpha1.Cluster) any,
	defaultValue any,
	description ...string,
) FieldSelector[v1alpha1.Cluster] {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return FieldSelector[v1alpha1.Cluster]{
		Selector:     selector,
		Description:  desc,
		DefaultValue: defaultValue,
	}
}
