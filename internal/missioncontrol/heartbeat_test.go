package missioncontrol

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRunCycle_callsDispatchThenReap(t *testing.T) {
	t.Parallel()

	var callOrder []string

	dispatchFn := func() (string, error) {
		callOrder = append(callOrder, "dispatch")
		return "dispatched #42", nil
	}

	reapFn := func() (string, error) {
		callOrder = append(callOrder, "reap")
		return "reaped worker-41", nil
	}

	result, err := RunCycle(Funcs{
		Dispatch: dispatchFn,
		Reap:     reapFn,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(callOrder) != 2 || callOrder[0] != "dispatch" || callOrder[1] != "reap" {
		t.Errorf("expected [dispatch, reap], got %v", callOrder)
	}

	if result.DispatchResult != "dispatched #42" {
		t.Errorf("expected dispatch result, got %q", result.DispatchResult)
	}
	if result.ReapResult != "reaped worker-41" {
		t.Errorf("expected reap result, got %q", result.ReapResult)
	}
}

func TestLoop_runsImmediatelyThenStopsOnCancel(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	cycleCount := 0

	dispatchFn := func() (string, error) {
		mu.Lock()
		cycleCount++
		mu.Unlock()
		return "ok", nil
	}

	reapFn := func() (string, error) {
		return "ok", nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	var logs []string
	logFn := func(msg string) {
		mu.Lock()
		logs = append(logs, msg)
		mu.Unlock()
	}

	done := make(chan struct{})
	go func() {
		Loop(ctx, 50*time.Millisecond, Funcs{
			Dispatch: dispatchFn,
			Reap:     reapFn,
		}, logFn)
		close(done)
	}()

	// Wait for at least one cycle.
	time.Sleep(80 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	count := cycleCount
	mu.Unlock()

	if count < 1 {
		t.Errorf("expected at least 1 cycle, got %d", count)
	}
}

func TestRunCycle_dispatchErrorDoesNotBlockReap(t *testing.T) {
	t.Parallel()

	dispatchFn := func() (string, error) {
		return "", fmt.Errorf("board fetch failed")
	}

	reapCalled := false
	reapFn := func() (string, error) {
		reapCalled = true
		return "reaped worker-41", nil
	}

	result, err := RunCycle(Funcs{
		Dispatch: dispatchFn,
		Reap:     reapFn,
	})
	if err != nil {
		t.Fatalf("RunCycle itself should not error: %v", err)
	}

	if !reapCalled {
		t.Error("expected reap to run even when dispatch fails")
	}
	if result.DispatchErr == nil {
		t.Error("expected dispatch error to be captured")
	}
	if result.ReapResult != "reaped worker-41" {
		t.Errorf("expected reap result, got %q", result.ReapResult)
	}
}
