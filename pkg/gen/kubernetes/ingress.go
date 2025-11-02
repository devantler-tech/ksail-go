package kubernetes

// IngressGenerator generates Ingress manifests.
type IngressGenerator struct {
	*Generator
}

// NewIngressGenerator creates a new generator for Ingress resources.
func NewIngressGenerator() *IngressGenerator {
	return &IngressGenerator{
		Generator: NewGenerator("ingress"),
	}
}
