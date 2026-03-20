// Package heartbeat provides the periodic dispatch + reap loop.
// The dumb, reliable background process — no AI, no decisions.
package missioncontrol

import (
	"context"
	"fmt"
	"time"
)

// CycleFunc is a function that performs one operation and returns a summary string.
type CycleFunc func() (string, error)

// Funcs holds the dispatch and reap functions injected by the cmd layer.
type Funcs struct {
	Dispatch CycleFunc
	Reap     CycleFunc
}

// CycleResult holds the outcome of one heartbeat cycle.
type CycleResult struct {
	DispatchResult string
	ReapResult     string
	DispatchErr    error
	ReapErr        error
}

// RunCycle executes one dispatch + reap cycle.
func RunCycle(fns Funcs) (*CycleResult, error) {
	result := &CycleResult{}

	if fns.Dispatch != nil {
		res, err := fns.Dispatch()
		result.DispatchResult = res
		result.DispatchErr = err
	}

	if fns.Reap != nil {
		res, err := fns.Reap()
		result.ReapResult = res
		result.ReapErr = err
	}

	return result, nil
}

// Loop runs dispatch + reap on a ticker until the context is cancelled.
func Loop(ctx context.Context, interval time.Duration, fns Funcs, log func(string)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start.
	runAndLog(fns, log)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runAndLog(fns, log)
		}
	}
}

func runAndLog(fns Funcs, log func(string)) {
	result, _ := RunCycle(fns)

	ts := time.Now().Format("15:04:05")

	if result.DispatchErr != nil {
		log(fmt.Sprintf("[%s] dispatch error: %v", ts, result.DispatchErr))
	} else if result.DispatchResult != "" {
		log(fmt.Sprintf("[%s] dispatch: %s", ts, result.DispatchResult))
	}

	if result.ReapErr != nil {
		log(fmt.Sprintf("[%s] reap error: %v", ts, result.ReapErr))
	} else if result.ReapResult != "" {
		log(fmt.Sprintf("[%s] reap: %s", ts, result.ReapResult))
	}
}
