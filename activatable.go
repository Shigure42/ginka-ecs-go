package ginka_ecs_go

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
