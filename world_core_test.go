package ginka_ecs_go

import (
	"errors"
	"testing"
	"time"
)

type testSystem struct {
	name string
}

func (s *testSystem) Name() string {
	return s.name
}

func TestCoreWorld_RegisterDuplicateName(t *testing.T) {
	w := NewCoreWorld("test")
	sys1 := &testSystem{name: "dup"}
	sys2 := &testSystem{name: "dup"}
	if err := w.Register(sys1); err != nil {
		t.Fatalf("register sys1: %v", err)
	}
	if err := w.Register(sys2); !errors.Is(err, ErrSystemAlreadyRegistered) {
		t.Fatalf("expected ErrSystemAlreadyRegistered, got %v", err)
	}
}

func TestCoreWorld_RegisterWhileRunning(t *testing.T) {
	w := NewCoreWorld("test")
	runDone := make(chan error, 1)
	go func() {
		runDone <- w.Run()
	}()
	waitForRunning(t, w)
	if err := w.Register(&testSystem{name: "late"}); !errors.Is(err, ErrWorldAlreadyRunning) {
		t.Fatalf("expected ErrWorldAlreadyRunning, got %v", err)
	}
	if err := w.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := <-runDone; err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestCoreWorld_SystemsSnapshot(t *testing.T) {
	w := NewCoreWorld("test")
	sys1 := &testSystem{name: "a"}
	sys2 := &testSystem{name: "b"}
	if err := w.Register(sys1, sys2); err != nil {
		t.Fatalf("register: %v", err)
	}
	list := w.Systems()
	if len(list) != 2 {
		t.Fatalf("expected 2 systems, got %d", len(list))
	}
	if list[0] != sys1 || list[1] != sys2 {
		t.Fatalf("unexpected system order")
	}
	list[0] = sys2
	list2 := w.Systems()
	if list2[0] != sys1 {
		t.Fatalf("expected snapshot to be isolated")
	}
}

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
