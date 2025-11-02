package kubernetes

// ServiceGenerator generates Service manifests.
type ServiceGenerator struct {
	*Generator
}

// NewServiceGenerator creates a new generator for Service resources.
func NewServiceGenerator() *ServiceGenerator {
	return &ServiceGenerator{
		Generator: NewGenerator("service"),
	}
}
