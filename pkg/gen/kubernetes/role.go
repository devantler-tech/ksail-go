package kubernetes

// RoleGenerator generates Role manifests.
type RoleGenerator struct {
	*Generator
}

// NewRoleGenerator creates a new generator for Role resources.
func NewRoleGenerator() *RoleGenerator {
	return &RoleGenerator{
		Generator: NewGenerator("role"),
	}
}
