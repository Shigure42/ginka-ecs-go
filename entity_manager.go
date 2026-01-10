package ginka_ecs_go

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// EntityFactory is a function type that creates new entities of type T.
// This allows for dependency injection and flexible entity construction.
type EntityFactory[T Entity] func(id uint64, name string, typ EntityType, tags ...Tag) (T, error)

// MapEntityManager is a simple entity manager backed by a Go map.
//
// It only maintains an in-process index of entities. Loading/hydration from
// other sources should be done by the caller via Create/Add.
type MapEntityManager[T Entity] struct {
	mu      sync.RWMutex
	byId    map[uint64]T
	factory EntityFactory[T]
}

// NewMapEntityManager creates a new MapEntityManager with the given entity factory.
func NewMapEntityManager[T Entity](factory EntityFactory[T]) *MapEntityManager[T] {
	m := &MapEntityManager[T]{
		byId:    make(map[uint64]T),
		factory: factory,
	}
	return m
}

// Create allocates and registers a new entity with the given parameters.
func (m *MapEntityManager[T]) Create(ctx context.Context, id uint64, name string, typ EntityType, tags ...Tag) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if id == 0 {
		return zero, ErrInvalidEntityId
	}
	if m.factory == nil {
		return zero, fmt.Errorf("create entity: factory is nil")
	}

	m.mu.RLock()
	_, exists := m.byId[id]
	m.mu.RUnlock()
	if exists {
		return zero, fmt.Errorf("create entity %d: %w", id, ErrEntityAlreadyExists)
	}

	ent, err := m.factory(id, name, typ, tags...)
	if err != nil {
		return zero, err
	}
	if isNil(ent) {
		return zero, fmt.Errorf("create entity: nil entity")
	}
	if ent.Id() != id {
		return zero, fmt.Errorf("create entity: id mismatch: want %d got %d", id, ent.Id())
	}

	if err := m.Add(ctx, ent); err != nil {
		return zero, err
	}
	return ent, nil
}

// Add registers an existing entity with the manager.
func (m *MapEntityManager[T]) Add(ctx context.Context, ent T) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if isNil(ent) {
		return fmt.Errorf("add entity: nil entity")
	}
	id := ent.Id()
	if id == 0 {
		return fmt.Errorf("add entity: %w", ErrInvalidEntityId)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byId[id]; ok {
		return fmt.Errorf("add entity %d: %w", id, ErrEntityAlreadyExists)
	}
	m.byId[id] = ent
	return nil
}

// isNil safely checks if a value is nil, handling interface types correctly.
// This is needed because comparing interfaces to nil requires special handling.
func isNil[T any](v T) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}

	for rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}

// Get retrieves an entity by ID.
func (m *MapEntityManager[T]) Get(id uint64) (T, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ent, ok := m.byId[id]
	return ent, ok
}

// MustGet retrieves an entity by ID, panicking if not found.
func (m *MapEntityManager[T]) MustGet(id uint64) T {
	ent, ok := m.Get(id)
	if !ok {
		panic(ErrEntityNotFound)
	}
	return ent
}

// Remove deletes an entity by ID, returning true if the entity existed.
func (m *MapEntityManager[T]) Remove(id uint64) bool {
	m.mu.Lock()
	_, ok := m.byId[id]
	delete(m.byId, id)
	m.mu.Unlock()
	return ok
}

// Len returns the total number of managed entities.
func (m *MapEntityManager[T]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.byId)
}

// ForEach calls the provided function for each entity.
func (m *MapEntityManager[T]) ForEach(ctx context.Context, fn func(ent T) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.RLock()
	snapshot := make([]T, 0, len(m.byId))
	for _, ent := range m.byId {
		snapshot = append(snapshot, ent)
	}
	m.mu.RUnlock()

	for _, ent := range snapshot {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(ent); err != nil {
			return err
		}
	}
	return nil
}

// Range iterates entities under a read lock without allocating.
//
// The callback must not call methods that acquire the write lock (e.g. Add/Create/Remove),
// otherwise it may deadlock.
func (m *MapEntityManager[T]) Range(ctx context.Context, fn func(ent T) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ent := range m.byId {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(ent); err != nil {
			return err
		}
	}
	return nil
}

// ForEachWithComponent iterates entities that have the given component type.
//
// Implementations may use a scan or an index.
func (m *MapEntityManager[T]) ForEachWithComponent(ctx context.Context, t ComponentType, fn func(ent T) error) error {
	return m.ForEach(ctx, func(ent T) error {
		if ent.Has(t) {
			return fn(ent)
		}
		return nil
	})
}

// ForEachWithComponents iterates entities that have all given component types.
//
// Implementations may use a scan or an index.
func (m *MapEntityManager[T]) ForEachWithComponents(ctx context.Context, types []ComponentType, fn func(ent T) error) error {
	if len(types) == 0 {
		return m.ForEach(ctx, fn)
	}
	return m.ForEach(ctx, func(ent T) error {
		for _, t := range types {
			if !ent.Has(t) {
				return nil
			}
		}
		return fn(ent)
	})
}

var _ EntityManager[Entity] = (*MapEntityManager[Entity])(nil)
