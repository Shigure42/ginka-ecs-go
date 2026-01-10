package ginka_ecs_go

// ComponentType uniquely identifies the type of a Component.
type ComponentType int

// Activatable indicates whether an entity/component should participate in systems.
//
// Implementations should treat Enabled as an in-memory runtime flag. Disabled
// entities/components are still expected to be accessible via the APIs (e.g.
// Get/Has). The concrete World implementation defines how Enabled affects
// command/tick execution and persistence.
type Activatable interface {
	// Enabled returns whether this component is currently active.
	Enabled() bool
	// SetEnabled enables or disables this component.
	SetEnabled(bool)
}

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
