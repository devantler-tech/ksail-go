package kubernetes

// DeploymentGenerator generates Deployment manifests.
type DeploymentGenerator struct {
	*Generator
}

// NewDeploymentGenerator creates a new generator for Deployment resources.
func NewDeploymentGenerator() *DeploymentGenerator {
	return &DeploymentGenerator{
		Generator: NewGenerator("deployment"),
	}
}
