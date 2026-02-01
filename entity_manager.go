package ginka_ecs_go

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// EntityFactory is a function type that creates new entities of type T.
// This allows for dependency injection and flexible entity construction.
type EntityFactory[T Entity] func(id string, name string, typ EntityType, tags ...Tag) (T, error)

// MapEntityManager is a simple entity manager backed by sharded Go maps.
//
// It only maintains an in-process index of entities. Loading/hydration from
// other sources should be done by the caller via Create/Add.
type MapEntityManager[T Entity] struct {
	shards    []entityShard[T]
	shardMask uint64
	factory   EntityFactory[T]
}

type entityShard[T Entity] struct {
	mu   sync.RWMutex
	byId map[string]T
}

const defaultEntityShardCount = 128

// NewEntityManager creates a new MapEntityManager with the given entity factory.
func NewEntityManager[T Entity](factory EntityFactory[T], shardCount int) *MapEntityManager[T] {
	if shardCount <= 0 {
		shardCount = defaultEntityShardCount
	}
	count := nextPow2(uint64(shardCount))
	if count == 0 {
		count = defaultEntityShardCount
	}
	shards := make([]entityShard[T], int(count))
	for i := range shards {
		shards[i].byId = make(map[string]T)
	}
	return &MapEntityManager[T]{
		shards:    shards,
		shardMask: count - 1,
		factory:   factory,
	}
}

// Create allocates and registers a new entity with the given parameters.
func (m *MapEntityManager[T]) Create(ctx context.Context, id string, name string, typ EntityType, tags ...Tag) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if id == "" {
		return zero, ErrInvalidEntityId
	}
	if m.factory == nil {
		return zero, fmt.Errorf("create entity: factory is nil")
	}

	ent, err := m.factory(id, name, typ, tags...)
	if err != nil {
		return zero, err
	}
	if isNil(ent) {
		return zero, fmt.Errorf("create entity: nil entity")
	}
	if ent.Id() != id {
		return zero, fmt.Errorf("create entity: id mismatch: want %s got %s", id, ent.Id())
	}

	if err := m.Add(ctx, ent); err != nil {
		if errors.Is(err, ErrEntityAlreadyExists) {
			return zero, fmt.Errorf("create entity %s: %w", id, ErrEntityAlreadyExists)
		}
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
	if id == "" {
		return fmt.Errorf("add entity: %w", ErrInvalidEntityId)
	}

	shard := m.shard(id)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	if _, ok := shard.byId[id]; ok {
		return fmt.Errorf("add entity %s: %w", id, ErrEntityAlreadyExists)
	}
	shard.byId[id] = ent
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
func (m *MapEntityManager[T]) Get(id string) (T, bool) {
	shard := m.shard(id)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	ent, ok := shard.byId[id]
	return ent, ok
}

// MustGet retrieves an entity by ID, panicking if not found.
func (m *MapEntityManager[T]) MustGet(id string) T {
	ent, ok := m.Get(id)
	if !ok {
		panic(ErrEntityNotFound)
	}
	return ent
}

// Remove deletes an entity by ID, returning true if the entity existed.
func (m *MapEntityManager[T]) Remove(id string) bool {
	shard := m.shard(id)
	shard.mu.Lock()
	_, ok := shard.byId[id]
	delete(shard.byId, id)
	shard.mu.Unlock()
	return ok
}

// Len returns the total number of managed entities.
func (m *MapEntityManager[T]) Len() int {
	count := 0
	for i := range m.shards {
		shard := &m.shards[i]
		shard.mu.RLock()
		count += len(shard.byId)
		shard.mu.RUnlock()
	}
	return count
}

// ForEach calls the provided function for each entity.
func (m *MapEntityManager[T]) ForEach(ctx context.Context, fn func(ent T) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for i := range m.shards {
		if err := ctx.Err(); err != nil {
			return err
		}
		shard := &m.shards[i]
		shard.mu.RLock()
		snapshot := make([]T, 0, len(shard.byId))
		for _, ent := range shard.byId {
			snapshot = append(snapshot, ent)
		}
		shard.mu.RUnlock()
		for _, ent := range snapshot {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := fn(ent); err != nil {
				return err
			}
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

// ForEachWithAllComponents iterates entities that have all given component types.
//
// Implementations may use a scan or an index.
func (m *MapEntityManager[T]) ForEachWithAllComponents(ctx context.Context, types []ComponentType, fn func(ent T) error) error {
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

func (m *MapEntityManager[T]) shard(id string) *entityShard[T] {
	idx := int(hashEntityId(id) & m.shardMask)
	return &m.shards[idx]
}

// hashEntityId hashes a string id for shard routing (cheap, non-cryptographic).
func hashEntityId(id string) uint64 {
	h := uint64(0)
	for i := 0; i < len(id); i++ {
		h = h*31 + uint64(id[i])
	}
	return h
}

func nextPow2(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

var _ EntityManager[Entity] = (*MapEntityManager[Entity])(nil)
