package kubernetes

// RoleBindingGenerator generates RoleBinding manifests.
type RoleBindingGenerator struct {
	*Generator
}

// NewRoleBindingGenerator creates a new generator for RoleBinding resources.
func NewRoleBindingGenerator() *RoleBindingGenerator {
	return &RoleBindingGenerator{
		Generator: NewGenerator("rolebinding"),
	}
}
