package kubernetes

// ConfigMapGenerator generates ConfigMap manifests.
type ConfigMapGenerator struct {
	*Generator
}

// NewConfigMapGenerator creates a new generator for ConfigMap resources.
func NewConfigMapGenerator() *ConfigMapGenerator {
	return &ConfigMapGenerator{
		Generator: NewGenerator("configmap"),
	}
}
