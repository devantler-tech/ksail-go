package kubernetes

// PriorityClassGenerator generates PriorityClass manifests.
type PriorityClassGenerator struct {
	*Generator
}

// NewPriorityClassGenerator creates a new generator for PriorityClass resources.
func NewPriorityClassGenerator() *PriorityClassGenerator {
	return &PriorityClassGenerator{
		Generator: NewGenerator("priorityclass"),
	}
}
