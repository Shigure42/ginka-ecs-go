package ginka_ecs_go

import "context"

// ComponentForEacher is an optional extension for entities that can iterate
// components without allocating.
type ComponentForEacher interface {
	ForEachComponent(fn func(t ComponentType, c Component) error) error
}

// DirtyTypeForEacher is an optional extension for data entities that can iterate
// dirty component types without allocating.
type DirtyTypeForEacher interface {
	ForEachDirtyType(fn func(t ComponentType) error) error
}

// EntityManagerRanger is an optional extension for managers that can iterate
// entities without allocating.
type EntityManagerRanger[T Entity] interface {
	EntityManager[T]
	Range(ctx context.Context, fn func(ent T) error) error
}

// ComponentQueryableEntityManager is an optional extension for managers that can
// iterate entities by component constraints.
//
// Implementations may use a scan or an index.
type ComponentQueryableEntityManager[T Entity] interface {
	EntityManager[T]
	ForEachWithComponent(ctx context.Context, t ComponentType, fn func(ent T) error) error
	ForEachWithComponents(ctx context.Context, types []ComponentType, fn func(ent T) error) error
}
