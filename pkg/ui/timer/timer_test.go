package timer_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	start := time.Now()
	tmr := timer.New()

	// Timer should be created recently
	assert.WithinDuration(t, start, start.Add(tmr.Total()), 10*time.Millisecond)
}

func TestTotal(t *testing.T) {
	t.Parallel()

	tmr := timer.New()

	time.Sleep(50 * time.Millisecond)

	total := tmr.Total()
	assert.GreaterOrEqual(t, total, 50*time.Millisecond)
	assert.Less(t, total, 100*time.Millisecond)
}

func TestStage(t *testing.T) {
	t.Parallel()

	tmr := timer.New()

	time.Sleep(50 * time.Millisecond)

	// Start a new stage
	tmr.StartStage()
	time.Sleep(30 * time.Millisecond)

	// Stage duration should be ~30ms
	stage := tmr.Stage()
	assert.GreaterOrEqual(t, stage, 30*time.Millisecond)
	assert.Less(t, stage, 60*time.Millisecond)

	// Total duration should be ~80ms
	total := tmr.Total()
	assert.GreaterOrEqual(t, total, 80*time.Millisecond)
	assert.Less(t, total, 120*time.Millisecond)
}

func TestMultipleStages(t *testing.T) {
	t.Parallel()

	tmr := timer.New()

	// First stage
	tmr.StartStage()
	time.Sleep(20 * time.Millisecond)
	stage1 := tmr.Stage()
	assert.GreaterOrEqual(t, stage1, 20*time.Millisecond)

	// Second stage
	tmr.StartStage()
	time.Sleep(30 * time.Millisecond)
	stage2 := tmr.Stage()
	assert.GreaterOrEqual(t, stage2, 30*time.Millisecond)

	// Stage durations should be independent
	assert.NotEqual(t, stage1, stage2)

	// Total should be cumulative
	total := tmr.Total()
	assert.GreaterOrEqual(t, total, 50*time.Millisecond)
}
