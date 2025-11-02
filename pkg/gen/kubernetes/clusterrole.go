package kubernetes

// ClusterRoleGenerator generates ClusterRole manifests.
type ClusterRoleGenerator struct {
	*Generator
}

// NewClusterRoleGenerator creates a new generator for ClusterRole resources.
func NewClusterRoleGenerator() *ClusterRoleGenerator {
	return &ClusterRoleGenerator{
		Generator: NewGenerator("clusterrole"),
	}
}
