package ginka_ecs_go

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	defaultTickInterval = 50 * time.Millisecond
	defaultShardCount   = 256
)

type TickErrorHandler func(err error)

type WorldOption func(*CoreWorld)

func WithTickInterval(interval time.Duration) WorldOption {
	return func(w *CoreWorld) {
		w.tickInterval = interval
	}
}

func WithShardCount(n int) WorldOption {
	return func(w *CoreWorld) {
		if n > 0 {
			w.shardCount = n
		}
	}
}

func WithEntityManager(m EntityManager[DataEntity]) WorldOption {
	return func(w *CoreWorld) {
		if m != nil {
			w.entities = m
		}
	}
}

func WithTickErrorHandler(h TickErrorHandler) WorldOption {
	return func(w *CoreWorld) {
		w.tickErrHandler = h
	}
}

// CoreWorld is a simple in-process World implementation.
//
// Concurrency model:
//   - Submit is safe for concurrent use.
//   - Commands are handled serially per EntityId by routing to a sharded worker.
//   - Tick is driven by explicit TickOnce calls (e.g. external cron).
//     Optionally, an internal ticker can trigger ticks when WithTickInterval > 0.
type CoreWorld struct {
	name string

	mu         sync.Mutex
	running    bool
	stopWeight int64

	entities EntityManager[DataEntity]

	systemNames map[string]struct{}

	commandAll         []CommandSystem
	commandBy          map[CommandType][]CommandSystem
	shardedTickSystems []ShardedTickSystem

	shardCount     int
	tickInterval   time.Duration
	tickErrHandler TickErrorHandler

	rt *worldRuntime
}

// NewCoreWorld creates a new in-process world.
//
// If no EntityManager is provided, a MapEntityManager is used with DataEntityCore.
func NewCoreWorld(name string, opts ...WorldOption) *CoreWorld {
	w := &CoreWorld{
		name:         name,
		systemNames:  make(map[string]struct{}),
		commandBy:    make(map[CommandType][]CommandSystem),
		shardCount:   defaultShardCount,
		tickInterval: 0,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}
	if w.entities == nil {
		w.entities = NewMapEntityManager(func(id uint64, entityName string, typ EntityType, tags ...Tag) (DataEntity, error) {
			return NewDataEntityCore(id, entityName, typ, tags...), nil
		})
	}
	return w
}

func (w *CoreWorld) Run() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return ErrWorldAlreadyRunning
	}
	rt, err := w.finalizeLocked()
	if err != nil {
		return err
	}
	w.rt = rt
	w.running = true
	w.startLocked(rt)
	return nil
}

func (w *CoreWorld) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = false
	rt := w.rt
	w.rt = nil
	w.mu.Unlock()
	if rt != nil {
		rt.stop()
	}
	return nil
}

func (w *CoreWorld) GetName() string {
	return w.name
}

func (w *CoreWorld) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

func (w *CoreWorld) GetStopWeight() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stopWeight
}

func (w *CoreWorld) SetStopWeight(weight int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stopWeight = weight
}

func (w *CoreWorld) Entities() EntityManager[DataEntity] {
	return w.entities
}

func (w *CoreWorld) Register(systems ...System) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return fmt.Errorf("register systems: %w", ErrWorldAlreadyRunning)
	}

	for _, sys := range systems {
		if sys == nil {
			return fmt.Errorf("register system: nil")
		}
		name := sys.Name()
		if _, exists := w.systemNames[name]; exists {
			return fmt.Errorf("register system %s: %w", name, ErrSystemAlreadyRegistered)
		}
		w.systemNames[name] = struct{}{}

		if cmdSys, ok := sys.(CommandSystem); ok {
			w.registerCommandSystemLocked(cmdSys)
		}
		if tickSys, ok := sys.(ShardedTickSystem); ok {
			w.shardedTickSystems = append(w.shardedTickSystems, tickSys)
		}
	}
	return nil
}

func (w *CoreWorld) Submit(ctx context.Context, cmd Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if cmd == nil {
		return fmt.Errorf("submit command: nil")
	}
	id := cmd.EntityId()
	if id == 0 {
		return ErrInvalidEntityId
	}

	w.mu.Lock()
	rt := w.rt
	w.mu.Unlock()
	if rt == nil {
		return ErrWorldNotRunning
	}
	if !rt.tryAcquire() {
		return ErrWorldNotRunning
	}
	defer rt.release()
	idx := int(hashEntityId(id) & rt.shardMask)

	resp := make(chan error, 1)
	req := shardRequest{kind: shardRequestCommand, ctx: ctx, cmd: cmd, resp: resp}

	select {
	case rt.shards[idx].ch <- req:
		// ok
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-resp:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TickOnce executes a single tick pass. It is primarily useful for tests.
func (w *CoreWorld) TickOnce(ctx context.Context, dt time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	w.mu.Lock()
	rt := w.rt
	w.mu.Unlock()
	if rt == nil {
		return ErrWorldNotRunning
	}
	if !rt.tryAcquire() {
		return ErrWorldNotRunning
	}
	defer rt.release()
	return rt.tickOnce(ctx, dt, w)
}

func (w *CoreWorld) registerCommandSystemLocked(sys CommandSystem) {
	if subscriber, ok := sys.(CommandSubscriber); ok {
		for _, typ := range subscriber.SubscribedCommands() {
			w.commandBy[typ] = append(w.commandBy[typ], sys)
		}
		return
	}
	w.commandAll = append(w.commandAll, sys)
}

var _ World = (*CoreWorld)(nil)
