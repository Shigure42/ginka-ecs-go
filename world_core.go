package ginka_ecs_go

import (
	"sync"
)

// CoreWorld manages world runtime lifecycle.
type CoreWorld struct {
	name string

	mu         sync.RWMutex
	running    bool
	started    bool
	stopWeight int64
	stopOnce   sync.Once
	stopChan   chan struct{}
	stopAwait  chan struct{}
}

// NewCoreWorld creates a new CoreWorld.
func NewCoreWorld(name string) *CoreWorld {
	w := &CoreWorld{
		name:      name,
		stopChan:  make(chan struct{}),
		stopAwait: make(chan struct{}),
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

	close(w.stopAwait)
	return nil
}

func (w *CoreWorld) Stop() error {
	w.mu.RLock()
	started := w.started
	running := w.running
	stopAwait := w.stopAwait
	w.mu.RUnlock()
	if !started {
		return nil
	}

	if running {
		w.stopOnce.Do(func() {
			close(w.stopChan)
		})
	}
	<-stopAwait
	return nil
}

func (w *CoreWorld) GetName() string {
	return w.name
}

func (w *CoreWorld) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

func (w *CoreWorld) GetStopWeight() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stopWeight
}

func (w *CoreWorld) SetStopWeight(weight int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stopWeight = weight
}

var _ World = (*CoreWorld)(nil)
