package kubernetes

// ServiceAccountGenerator generates ServiceAccount manifests.
type ServiceAccountGenerator struct {
	*Generator
}

// NewServiceAccountGenerator creates a new generator for ServiceAccount resources.
func NewServiceAccountGenerator() *ServiceAccountGenerator {
	return &ServiceAccountGenerator{
		Generator: NewGenerator("serviceaccount"),
	}
}
