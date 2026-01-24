package ginka_ecs_go

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testCommandType CommandType = 1

type testCommand struct {
	entityId uint64
}

func (c testCommand) Type() CommandType {
	return testCommandType
}

func (c testCommand) EntityId() uint64 {
	return c.entityId
}

type serialCheckSystem struct {
	t        *testing.T
	inflight int32
}

func (s *serialCheckSystem) Name() string { return "serial-check" }

func (s *serialCheckSystem) SubscribedCommands() []CommandType {
	return []CommandType{testCommandType}
}

func (s *serialCheckSystem) Handle(ctx context.Context, w World, cmd Command) error {
	n := atomic.AddInt32(&s.inflight, 1)
	if n != 1 {
		s.t.Errorf("concurrent handle: inflight=%d", n)
	}
	time.Sleep(2 * time.Millisecond)
	atomic.AddInt32(&s.inflight, -1)
	return nil
}

type tickCounterSystem struct {
	count int32
}

func (s *tickCounterSystem) Name() string { return "tick-counter" }

func (s *tickCounterSystem) TickShard(ctx context.Context, w World, dt time.Duration, shardIdx int, shardCount int) error {
	atomic.AddInt32(&s.count, 1)
	return nil
}

func TestCoreWorld_RegisterDuplicateName(t *testing.T) {
	w := NewCoreWorld("test")
	sys1 := &serialCheckSystem{t: t}
	sys2 := &serialCheckSystem{t: t}
	if err := w.Register(sys1); err != nil {
		t.Fatalf("register sys1: %v", err)
	}
	if err := w.Register(sys2); !errors.Is(err, ErrSystemAlreadyRegistered) {
		t.Fatalf("expected ErrSystemAlreadyRegistered, got %v", err)
	}
}

func TestCoreWorld_SubmitNotRunning(t *testing.T) {
	w := NewCoreWorld("test")
	if err := w.Submit(context.Background(), testCommand{entityId: 1}); !errors.Is(err, ErrWorldNotRunning) {
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

	err := w.Submit(ctx, testCommand{entityId: 1})
	if !errors.Is(err, ErrUnhandledCommand) {
		t.Fatalf("expected ErrUnhandledCommand, got %v", err)
	}
}

func TestCoreWorld_SerialPerEntity(t *testing.T) {
	ctx := context.Background()
	sys := &serialCheckSystem{t: t}
	w := NewCoreWorld("test", WithShardCount(8))
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
			if err := w.Submit(ctx, testCommand{entityId: 42}); err != nil {
				t.Errorf("submit: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestCoreWorld_ParallelAcrossShards(t *testing.T) {
	ctx := context.Background()
	const shardCount = 2

	id1, id2 := findIdsDifferentShard(shardCount)
	if id1 == 0 || id2 == 0 {
		t.Fatalf("failed to find ids for shardCount=%d", shardCount)
	}

	started := make(chan uint64, 2)
	release := make(chan struct{})

	sys := &parallelSystem{
		t:       t,
		started: started,
		release: release,
	}
	w := NewCoreWorld("test", WithShardCount(shardCount))
	if err := w.Register(sys); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := w.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	defer func() { _ = w.Stop() }()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = w.Submit(ctx, testCommand{entityId: id1})
	}()
	go func() {
		defer wg.Done()
		_ = w.Submit(ctx, testCommand{entityId: id2})
	}()

	// If id1/id2 are on different shards, both should enter Handle without
	// blocking on each other.
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()

	seen := map[uint64]struct{}{}
	for len(seen) < 2 {
		select {
		case id := <-started:
			seen[id] = struct{}{}
		case <-timer.C:
			t.Fatalf("expected parallel handle for ids %d/%d", id1, id2)
		}
	}

	close(release)
	wg.Wait()
}

type parallelSystem struct {
	t       *testing.T
	started chan<- uint64
	release <-chan struct{}
}

func (s *parallelSystem) Name() string { return "parallel-check" }

func (s *parallelSystem) SubscribedCommands() []CommandType {
	return []CommandType{testCommandType}
}

func (s *parallelSystem) Handle(ctx context.Context, w World, cmd Command) error {
	s.started <- cmd.EntityId()
	<-s.release
	return nil
}

func TestCoreWorld_TickLoopRuns(t *testing.T) {
	ctx := context.Background()
	sys := &tickCounterSystem{}
	w := NewCoreWorld("test", WithTickInterval(10*time.Millisecond))
	if err := w.Register(sys); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := w.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	defer func() { _ = w.Stop() }()

	deadline := time.NewTimer(200 * time.Millisecond)
	defer deadline.Stop()
	for {
		if atomic.LoadInt32(&sys.count) > 0 {
			break
		}
		select {
		case <-deadline.C:
			t.Fatalf("expected tick to run")
		case <-time.After(5 * time.Millisecond):
		}
	}

	if err := w.TickOnce(ctx, time.Second); err != nil {
		t.Fatalf("tick once: %v", err)
	}
}

func findIdsDifferentShard(shardCount int) (uint64, uint64) {
	if shardCount <= 1 {
		return 0, 0
	}
	byShard := make(map[int]uint64)
	for id := uint64(1); id < 10_000; id++ {
		idx := ShardIndex(id, shardCount)
		if existing, ok := byShard[idx]; ok {
			_ = existing
			continue
		}
		byShard[idx] = id
		if len(byShard) >= 2 {
			var a, b uint64
			for _, v := range byShard {
				if a == 0 {
					a = v
					continue
				}
				b = v
				break
			}
			return a, b
		}
	}
	return 0, 0
}
