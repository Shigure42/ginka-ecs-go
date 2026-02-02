package ginka_ecs_go

// World coordinates runtime state and registered systems.
// Scheduling and execution are handled by the caller.
type World interface {
	// Run starts the world and blocks until Stop is called.
	// Run can only be called once.
	Run() error
	// Stop signals Run to exit.
	Stop() error
	// Name returns the world name.
	GetName() string
	// IsRunning reports whether the world is active.
	IsRunning() bool
	// StopWeight is used for graceful shutdown ordering.
	GetStopWeight() int64
	SetStopWeight(w int64)

	// Register adds systems to the world.
	// Returns ErrSystemAlreadyRegistered if a system with the same name exists.
	Register(systems ...System) error
	// Systems returns the registered systems in order.
	// The returned slice is a snapshot.
	Systems() []System
}
