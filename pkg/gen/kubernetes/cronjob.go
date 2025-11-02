package kubernetes

// CronJobGenerator generates CronJob manifests.
type CronJobGenerator struct {
	*Generator
}

// NewCronJobGenerator creates a new generator for CronJob resources.
func NewCronJobGenerator() *CronJobGenerator {
	return &CronJobGenerator{
		Generator: NewGenerator("cronjob"),
	}
}
