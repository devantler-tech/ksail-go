package kubernetes

// SecretGenerator generates Secret manifests.
type SecretGenerator struct {
	*Generator
}

// NewSecretGenerator creates a new generator for Secret resources.
func NewSecretGenerator() *SecretGenerator {
	return &SecretGenerator{
		Generator: NewGenerator("secret"),
	}
}
