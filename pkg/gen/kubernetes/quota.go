package kubernetes

// QuotaGenerator generates ResourceQuota manifests.
type QuotaGenerator struct {
	*Generator
}

// NewQuotaGenerator creates a new generator for ResourceQuota resources.
func NewQuotaGenerator() *QuotaGenerator {
	return &QuotaGenerator{
		Generator: NewGenerator("quota"),
	}
}
