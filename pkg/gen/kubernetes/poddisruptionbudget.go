package kubernetes

// PodDisruptionBudgetGenerator generates PodDisruptionBudget manifests.
type PodDisruptionBudgetGenerator struct {
	*Generator
}

// NewPodDisruptionBudgetGenerator creates a new generator for PodDisruptionBudget resources.
func NewPodDisruptionBudgetGenerator() *PodDisruptionBudgetGenerator {
	return &PodDisruptionBudgetGenerator{
		Generator: NewGenerator("poddisruptionbudget"),
	}
}
