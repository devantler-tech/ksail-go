package kubernetes

// NamespaceGenerator generates Namespace manifests.
type NamespaceGenerator struct {
	*Generator
}

// NewNamespaceGenerator creates a new generator for Namespace resources.
func NewNamespaceGenerator() *NamespaceGenerator {
	return &NamespaceGenerator{
		Generator: NewGenerator("namespace"),
	}
}
