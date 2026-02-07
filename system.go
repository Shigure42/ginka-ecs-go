package ginka_ecs_go

// System is a named business logic unit.
// Execution and organization are handled by the caller.
type System interface {
	// Name identifies the system.
	Name() string
}
