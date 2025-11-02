package kubernetes

// ClusterRoleBindingGenerator generates ClusterRoleBinding manifests.
type ClusterRoleBindingGenerator struct {
	*Generator
}

// NewClusterRoleBindingGenerator creates a new generator for ClusterRoleBinding resources.
func NewClusterRoleBindingGenerator() *ClusterRoleBindingGenerator {
	return &ClusterRoleBindingGenerator{
		Generator: NewGenerator("clusterrolebinding"),
	}
}
