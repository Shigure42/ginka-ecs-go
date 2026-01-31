package ginka_ecs_go

// Activatable indicates whether an entity/component should be processed by systems.
// Enabled is a runtime flag. Disabled entities are still accessible via Get/Has.
// The World implementation decides how Enabled affects execution and persistence.
type Activatable interface {
	// Enabled reports whether this is active.
	Enabled() bool
	// SetEnabled turns the entity/component on or off.
	SetEnabled(bool)
}
