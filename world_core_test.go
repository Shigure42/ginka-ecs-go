package ginka_ecs_go

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testCommand struct{}

type serialCheckSystem struct{}

func (s *serialCheckSystem) Name() string { return "serial-check" }

func (s *serialCheckSystem) Handle(ctx context.Context, w World, cmd Command) error {
	_, err := AsCommand[testCommand](cmd)
	return err
}

type tickCounterSystem struct {
	count int32
}

func (s *tickCounterSystem) Name() string { return "tick-counter" }

func (s *tickCounterSystem) Handle(ctx context.Context, w World, cmd Command) error {
	if _, err := TickEvent(cmd); err != nil {
		return err
	}
	atomic.AddInt32(&s.count, 1)
	return nil
}

func TestCoreWorld_RegisterDuplicateName(t *testing.T) {
	w := NewCoreWorld("test")
	sys1 := &serialCheckSystem{}
	sys2 := &serialCheckSystem{}
	if err := w.Register(sys1); err != nil {
		t.Fatalf("register sys1: %v", err)
	}
	if err := w.Register(sys2); !errors.Is(err, ErrSystemAlreadyRegistered) {
		t.Fatalf("expected ErrSystemAlreadyRegistered, got %v", err)
	}
}

func TestCoreWorld_SubmitNotRunning(t *testing.T) {
	w := NewCoreWorld("test")
	if err := w.Submit(context.Background(), NewAction(testCommand{})); !errors.Is(err, ErrWorldNotRunning) {
		t.Fatalf("expected ErrWorldNotRunning, got %v", err)
	}
}

func TestCoreWorld_UnhandledCommand(t *testing.T) {
	ctx := context.Background()
	w := NewCoreWorld("test")
	if err := w.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	defer func() { _ = w.Stop() }()

	err := w.Submit(ctx, NewAction(testCommand{}))
	if !errors.Is(err, ErrUnhandledCommand) {
		t.Fatalf("expected ErrUnhandledCommand, got %v", err)
	}
}

func TestCoreWorld_ConcurrentSubmit(t *testing.T) {
	ctx := context.Background()
	sys := &serialCheckSystem{}
	w := NewCoreWorld("test")
	if err := w.Register(sys); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := w.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	defer func() { _ = w.Stop() }()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if err := w.Submit(ctx, NewAction(testCommand{})); err != nil {
				t.Errorf("submit: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestCoreWorld_TickRuns(t *testing.T) {
	ctx := context.Background()
	sys := &tickCounterSystem{}
	w := NewCoreWorld("test")
	if err := w.Register(sys); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := w.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	defer func() { _ = w.Stop() }()

	if err := w.Submit(ctx, NewTick(time.Second)); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if atomic.LoadInt32(&sys.count) == 0 {
		t.Fatalf("expected tick to run")
	}
}
