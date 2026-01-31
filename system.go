package ginka_ecs_go

// System is a named system registered in a World.
// Execution is handled by the caller.
type System interface {
	// Name identifies the system.
	Name() string
}
