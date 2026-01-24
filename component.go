package ginka_ecs_go

// ComponentType uniquely identifies the type of a Component.
type ComponentType int

// Component is a piece of state attached to an Entity.
//
// A Component is identified by ComponentType. Implementations are expected to be
// safe for in-process use and should avoid exposing mutable internal state
// through returned slices.
type Component interface {
	Activatable
	Taggable

	// ComponentType returns the unique type identifier for this component.
	ComponentType() ComponentType
}
