package ginka_ecs_go

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type WorldOption func(*CoreWorld)

const defaultEntityManagerName = "default"

func WithEntityManager(m EntityManager[DataEntity]) WorldOption {
	return func(w *CoreWorld) {
		if m != nil {
			w.entities = m
			w.setEntityManagerLocked(defaultEntityManagerName, m)
		}
	}
}

func WithEntityManagerNamed(name string, m EntityManager[DataEntity]) WorldOption {
	return func(w *CoreWorld) {
		if name == "" || m == nil {
			return
		}
		if name == defaultEntityManagerName {
			w.entities = m
		}
		w.setEntityManagerLocked(name, m)
	}
}

// CoreWorld is a simple in-process World.
// Submit is safe to call from multiple goroutines.
type CoreWorld struct {
	name string

	mu         sync.Mutex
	running    bool
	stopWeight int64
	stopping   atomic.Bool
	refs       atomic.Int64
	refsMu     sync.Mutex
	refsCond   *sync.Cond

	entities EntityManager[DataEntity]

	entityManagers sync.Map

	systemNames sync.Map

	systems []System
}

// NewCoreWorld creates a new world.
// If no EntityManager is provided, it creates a MapEntityManager with DataEntityCore.
func NewCoreWorld(name string, opts ...WorldOption) *CoreWorld {
	w := &CoreWorld{
		name: name,
	}
	w.refsCond = sync.NewCond(&w.refsMu)
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}
	if w.entities == nil {
		w.entities = NewEntityManager(func(id uint64, entityName string, typ EntityType, tags ...Tag) (DataEntity, error) {
			return NewDataEntityCore(id, entityName, typ, tags...), nil
		}, defaultEntityShardCount)
		w.setEntityManagerLocked(defaultEntityManagerName, w.entities)
	} else if _, ok := w.entityManagers.Load(defaultEntityManagerName); !ok {
		w.setEntityManagerLocked(defaultEntityManagerName, w.entities)
	}
	return w
}

func (w *CoreWorld) Run() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return ErrWorldAlreadyRunning
	}
	w.stopping.Store(false)
	w.running = true
	return nil
}

func (w *CoreWorld) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = false
	w.stopping.Store(true)
	w.mu.Unlock()

	w.refsMu.Lock()
	for w.refs.Load() != 0 {
		w.refsCond.Wait()
	}
	w.refsMu.Unlock()
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

func (w *CoreWorld) EntitiesByName(name string) (EntityManager[DataEntity], bool) {
	if name == "" {
		name = defaultEntityManagerName
	}
	if m, ok := w.entityManagers.Load(name); ok {
		manager, ok := m.(EntityManager[DataEntity])
		return manager, ok
	}
	return nil, false
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
		if _, loaded := w.systemNames.LoadOrStore(name, struct{}{}); loaded {
			return fmt.Errorf("register system %s: %w", name, ErrSystemAlreadyRegistered)
		}
		w.systems = append(w.systems, sys)
	}
	return nil
}

func (w *CoreWorld) Submit(ctx context.Context, cmd Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if cmd.Kind == 0 {
		return fmt.Errorf("submit command: nil")
	}

	w.mu.Lock()
	running := w.running
	w.mu.Unlock()
	if !running {
		return ErrWorldNotRunning
	}
	if !w.tryAcquire() {
		return ErrWorldNotRunning
	}
	defer w.release()
	return dispatchCommand(ctx, w, cmd)
}

func (w *CoreWorld) tryAcquire() bool {
	if w.stopping.Load() {
		return false
	}
	w.refs.Add(1)
	if w.stopping.Load() {
		w.release()
		return false
	}
	return true
}

func (w *CoreWorld) release() {
	if w.refs.Add(-1) != 0 {
		return
	}
	w.refsMu.Lock()
	w.refsCond.Broadcast()
	w.refsMu.Unlock()
}

func (w *CoreWorld) setEntityManagerLocked(name string, m EntityManager[DataEntity]) {
	if name == "" || m == nil {
		return
	}
	w.entityManagers.Store(name, m)
}

func dispatchCommand(ctx context.Context, w *CoreWorld, cmd Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	systems := w.systems
	for _, sys := range systems {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := sys.Handle(ctx, w, cmd)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrUnhandledCommand) {
			continue
		}
		return err
	}
	return ErrUnhandledCommand
}

var _ World = (*CoreWorld)(nil)
