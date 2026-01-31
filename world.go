package ginka_ecs_go

import "context"

// World holds entities and runs systems.
type World interface {
	// Run starts the world.
	Run() error
	// Stop shuts down.
	Stop() error
	// Name returns the world name.
	GetName() string
	// IsRunning reports whether the world is active.
	IsRunning() bool
	// StopWeight is used for graceful shutdown ordering.
	GetStopWeight() int64
	SetStopWeight(w int64)

	// Entities returns the default entity manager.
	Entities() EntityManager[DataEntity]
	// EntitiesByName returns a named entity manager if it exists.
	EntitiesByName(name string) (EntityManager[DataEntity], bool)
	// Register adds systems to the world.
	// The first system that handles a command wins.
	// Returns ErrSystemAlreadyRegistered if a system with the same name exists.
	Register(systems ...System) error
	// Submit sends a command to be processed.
	// Runs synchronously in the caller's goroutine.
	// Use CommandKindTick for tick commands.
	Submit(ctx context.Context, cmd Command) error
}
