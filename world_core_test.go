package ginka_ecs_go

import (
	"errors"
	"testing"
	"time"
)

func TestCoreWorld_RunStop(t *testing.T) {
	w := NewCoreWorld("test")
	if w.IsRunning() {
		t.Fatalf("expected not running")
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- w.Run()
	}()
	waitForRunning(t, w)
	if !w.IsRunning() {
		t.Fatalf("expected running")
	}
	if err := w.Run(); !errors.Is(err, ErrWorldAlreadyRunning) {
		t.Fatalf("expected ErrWorldAlreadyRunning, got %v", err)
	}
	if err := w.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := <-runDone; err != nil {
		t.Fatalf("run: %v", err)
	}
	if w.IsRunning() {
		t.Fatalf("expected stopped")
	}
}

func TestCoreWorld_StopIdempotent(t *testing.T) {
	w := NewCoreWorld("test")
	runDone := make(chan error, 1)
	go func() {
		runDone <- w.Run()
	}()
	waitForRunning(t, w)

	if err := w.Stop(); err != nil {
		t.Fatalf("first stop: %v", err)
	}
	if err := w.Stop(); err != nil {
		t.Fatalf("second stop: %v", err)
	}

	if err := <-runDone; err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestCoreWorld_StopConcurrent(t *testing.T) {
	w := NewCoreWorld("test")
	runDone := make(chan error, 1)
	go func() {
		runDone <- w.Run()
	}()
	waitForRunning(t, w)

	stopErr := make(chan error, 2)
	go func() {
		stopErr <- w.Stop()
	}()
	go func() {
		stopErr <- w.Stop()
	}()

	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-stopErr:
			if err != nil {
				t.Fatalf("stop: %v", err)
			}
		case <-deadline.C:
			t.Fatalf("concurrent stop timed out")
		}
	}

	if err := <-runDone; err != nil {
		t.Fatalf("run: %v", err)
	}
}

func waitForRunning(t *testing.T, w *CoreWorld) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		if w.IsRunning() {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("world did not start")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
