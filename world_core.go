package ginka_ecs_go

import (
	"fmt"
	"sync"
)

// CoreWorld manages system registration and runtime lifecycle.
type CoreWorld struct {
	name string

	mu         sync.Mutex
	running    bool
	started    bool
	stopWeight int64
	stopOnce   sync.Once
	stopChan   chan struct{}
	stopAwait  chan struct{}

	systemNames sync.Map

	systems []System
}

// NewCoreWorld creates a new CoreWorld.
func NewCoreWorld(name string) *CoreWorld {
	w := &CoreWorld{
		name:      name,
		stopChan:  make(chan struct{}, 1),
		stopAwait: make(chan struct{}, 1),
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

var _ World = (*CoreWorld)(nil)
