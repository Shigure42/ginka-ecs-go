package ginka_ecs_go

import (
	"fmt"
	"sync"
)

type WorldOption func(*CoreWorld)

const defaultEntityManagerName = "default"

func WithEntityManager(m EntityManager[DataEntity]) WorldOption {
	return func(w *CoreWorld) {
		if m != nil {
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
			w.setEntityManagerLocked(defaultEntityManagerName, m)
			return
		}
		w.setEntityManagerLocked(name, m)
	}
}

// CoreWorld is a simple in-process World.
type CoreWorld struct {
	name string

	mu         sync.Mutex
	running    bool
	started    bool
	stopWeight int64
	stopOnce   sync.Once
	stopChan   chan struct{}
	stopAwait  chan struct{}

	entityManagers sync.Map

	systemNames sync.Map

	systems []System
}

// NewCoreWorld creates a new world.
// If no EntityManager is provided, it creates a MapEntityManager with DataEntityCore.
func NewCoreWorld(name string, opts ...WorldOption) *CoreWorld {
	w := &CoreWorld{
		name:      name,
		stopChan:  make(chan struct{}, 1),
		stopAwait: make(chan struct{}, 1),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}
	if _, ok := w.entityManagers.Load(defaultEntityManagerName); !ok {
		defaultManager := NewEntityManager(func(id string, entityName string, typ EntityType, tags ...Tag) (DataEntity, error) {
			return NewDataEntityCore(id, entityName, typ, tags...), nil
		}, defaultEntityShardCount)
		w.setEntityManagerLocked(defaultEntityManagerName, defaultManager)
	}
	return w
}

func (w *CoreWorld) Run() error {
	w.mu.Lock()
	if w.running || w.started {
		w.mu.Unlock()
		return ErrWorldAlreadyRunning
	}
	w.running = true
	w.started = true
	w.mu.Unlock()

	<-w.stopChan

	w.mu.Lock()
	w.running = false
	w.mu.Unlock()

	select {
	case w.stopAwait <- struct{}{}:
	default:
	}
	return nil
}

func (w *CoreWorld) Stop() error {
	w.mu.Lock()
	started := w.started
	w.mu.Unlock()
	if !started {
		return nil
	}

	w.stopOnce.Do(func() {
		select {
		case w.stopChan <- struct{}{}:
		default:
		}
	})
	<-w.stopAwait
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
	if m, ok := w.entityManagers.Load(defaultEntityManagerName); ok {
		manager, ok := m.(EntityManager[DataEntity])
		if ok {
			return manager
		}
	}
	return nil
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

func (w *CoreWorld) Systems() []System {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.systems) == 0 {
		return nil
	}
	out := make([]System, len(w.systems))
	copy(out, w.systems)
	return out
}

func (w *CoreWorld) setEntityManagerLocked(name string, m EntityManager[DataEntity]) {
	if name == "" || m == nil {
		return
	}
	w.entityManagers.Store(name, m)
}

var _ World = (*CoreWorld)(nil)
