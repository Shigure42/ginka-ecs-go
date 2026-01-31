package ginka_ecs_go

import "context"

// EntityType categorizes entities (e.g. player, npc, item).
type EntityType int

// Entity groups components and tags together.
type Entity interface {
	Activatable
	Taggable

	// Id is the entity's stable identifier.
	Id() uint64
	// Name is the human-readable label.
	Name() string
	// Type categorizes the entity.
	Type() EntityType

	// Has reports whether the entity has a component of type t.
	Has(t ComponentType) bool
	// Get returns the component of type t, or (nil, false) if not present.
	Get(t ComponentType) (Component, bool)
	// MustGet returns the component of type t, panicking if missing.
	MustGet(t ComponentType) Component

	// Add attaches component c to the entity.
	// Returns ErrComponentAlreadyExists if a component of the same type exists.
	Add(c Component) error
	// RemoveComponent detaches the component of type t.
	// Returns true if a component was removed.
	RemoveComponent(t ComponentType) bool
	// RemoveComponents detaches multiple components.
	// Returns the count of components actually removed.
	RemoveComponents(types []ComponentType) int
	// AllComponents returns all attached components in insertion order.
	AllComponents() []Component
}

// EntityManager keeps track of entities.
type EntityManager[T Entity] interface {
	// Create makes a new entity with the given id.
	// The id typically comes from somewhere else (e.g. player id from auth).
	Create(ctx context.Context, id uint64, name string, typ EntityType, tags ...Tag) (T, error)
	// Add registers an existing entity (e.g. loaded from DB).
	Add(ctx context.Context, ent T) error
	// Get fetches an entity by id.
	Get(id uint64) (T, bool)
	// MustGet fetches an entity by id, panicking if not found.
	MustGet(id uint64) T
	// Remove deletes an entity by id.
	Remove(id uint64) bool
	// Len returns the total count.
	Len() int
	// ForEach runs fn on every entity.
	ForEach(ctx context.Context, fn func(ent T) error) error
	// ForEachWithComponent runs fn on entities that have component type t.
	ForEachWithComponent(ctx context.Context, t ComponentType, fn func(ent T) error) error
	// ForEachWithAllComponents runs fn on entities that have all the given types.
	ForEachWithAllComponents(ctx context.Context, types []ComponentType, fn func(ent T) error) error
}
