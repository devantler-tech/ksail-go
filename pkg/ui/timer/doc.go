// Package timer provides timing functionality for tracking command execution duration.
//
// The timer package implements a simple, stateful timer that tracks total elapsed time
// and per-stage elapsed time for CLI command operations. It integrates with the notify
// package to display timing information in command output.
//
// Example usage for single-stage command:
//
//	timer := timer.New()
//	timer.Start()
//	// ... perform operation ...
//	total, stage := timer.GetTiming()
//	fmt.Printf("Operation completed [%s]\n", total)
//
// Example usage for multi-stage command:
//
//	timer := timer.New()
//	timer.Start()
//	// ... stage 1 ...
//	timer.NewStage()
//	// ... stage 2 ...
//	total, stage := timer.GetTiming()
//	fmt.Printf("Operation completed [%s total|%s stage]\n", total, stage)
package timer
