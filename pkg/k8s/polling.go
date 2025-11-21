package k8s

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	readinessPollInterval = 2 * time.Second
)

// PollForReadiness polls a check function until ready or timeout.
func PollForReadiness(
	ctx context.Context,
	deadline time.Duration,
	poll func(context.Context) (bool, error),
) error {
	pollErr := wait.PollUntilContextTimeout(
		ctx,
		readinessPollInterval,
		deadline,
		true,
		poll,
	)
	if pollErr != nil {
		return fmt.Errorf("poll for readiness: %w", pollErr)
	}

	return nil
}
