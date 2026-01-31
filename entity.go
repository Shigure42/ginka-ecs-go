package ginka_ecs_go

import "context"

// EntityType identifies the category of an entity.
type EntityType int

// Entity is a container of components.
//
// Notes:
// - Implementations are expected to have at most one Component per ComponentType.
// - Disabled entities are still expected to be accessible via the APIs (e.g. Get/Has).
// - Tags() returns a copy.
type Entity interface {
	Activatable
	Taggable

	// Id is the stable identifier of the entity (e.g. player id).
	Id() uint64
	// Name returns the human-readable name of the entity.
	Name() string
	// Type returns the category type of the entity.
	Type() EntityType

	// Has checks if the entity has a component of the given type.
	Has(t ComponentType) bool
	// Get retrieves a component by type, returning (nil, false) if not found.
	Get(t ComponentType) (Component, bool)
	// MustGet retrieves a component by type, panicking if not found.
	//
	// The panic value should wrap ErrComponentNotFound.
	MustGet(t ComponentType) Component

	// Add attaches c to the entity.
	//
	// If a component with the same ComponentType already exists, implementations
	// should return ErrComponentAlreadyExists (possibly wrapped).
	Add(c Component) error
	// RemoveComponent detaches the component for t and returns whether it existed.
	RemoveComponent(t ComponentType) bool
	// RemoveComponents detaches multiple components and returns the count of removed.
	RemoveComponents(types []ComponentType) int
	// AllComponents returns a copy of components in stable insertion order.
	AllComponents() []Component
}

// EntityManager provides lifecycle management for entities of type T.
type EntityManager[T Entity] interface {
	// Create allocates and registers a new entity with a caller-provided id.
	//
	// This is the primary creation path for server-side entities where ids are
	// assigned externally (e.g. player id).
	Create(ctx context.Context, id uint64, name string, typ EntityType, tags ...Tag) (T, error)
	// Add registers an existing entity (e.g. hydrated from persistence).
	Add(ctx context.Context, ent T) error
	// Get retrieves an entity by ID.
	Get(id uint64) (T, bool)
	// MustGet retrieves an entity by ID, panicking if not found.
	MustGet(id uint64) T
	// Remove deletes an entity by ID, returning true if the entity existed.
	Remove(id uint64) bool
	// Len returns the total number of managed entities.
	Len() int
	// ForEach calls the provided function for each entity.
	ForEach(ctx context.Context, fn func(ent T) error) error
	// ForEachWithComponent iterates entities that have the given component type.
	ForEachWithComponent(ctx context.Context, t ComponentType, fn func(ent T) error) error
	// ForEachWithAllComponents iterates entities that have all given component types.
	ForEachWithAllComponents(ctx context.Context, types []ComponentType, fn func(ent T) error) error
}
