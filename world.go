package ginka_ecs_go

import "context"

// World is the central container that manages all entities, components, and systems.
// It provides the execution environment for the ECS architecture and handles
// command routing and system execution.
type World interface {
	// Run starts the world and begins processing commands and ticks.
	Run() error
	// Stop shuts down the world and stops all processing.
	Stop() error
	// GetName returns the name of this world instance.
	GetName() string
	// IsRunning checks if the world is currently running.
	IsRunning() bool
	// GetStopWeight returns the current stop weight for graceful shutdown ordering.
	GetStopWeight() int64
	// SetStopWeight sets the stop weight for graceful shutdown ordering.
	SetStopWeight(w int64)

	// Entities returns the entity manager used by systems to access entities.
	Entities() EntityManager[DataEntity]
	// EntitiesByName returns a named entity manager, if registered.
	EntitiesByName(name string) (EntityManager[DataEntity], bool)
	// Register installs systems into the world.
	//
	// Order matters: the first system that handles an event wins.
	//
	// Implementations should return ErrSystemAlreadyRegistered (possibly wrapped)
	// if a system with the same name is registered more than once.
	Register(systems ...System) error
	// Submit routes and executes cmd.
	//
	// Submit executes synchronously in the caller goroutine.
	// Use CommandKindTick for tick execution.
	//
	// Dispatch contract:
	// - Systems are invoked in registration order.
	// - A system that does not handle the command should return ErrUnhandledCommand.
	// - The first system that returns nil or a non-ErrUnhandledCommand error stops dispatch.
	// - If no system handles the command, implementations should return ErrUnhandledCommand
	//   (possibly wrapped).
	Submit(ctx context.Context, cmd Command) error
}
