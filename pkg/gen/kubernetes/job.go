package kubernetes

// JobGenerator generates Job manifests.
type JobGenerator struct {
	*Generator
}

// NewJobGenerator creates a new generator for Job resources.
func NewJobGenerator() *JobGenerator {
	return &JobGenerator{
		Generator: NewGenerator("job"),
	}
}
