package ginka_ecs_go

import "context"

// System represents a system that operates on entities with specific component sets.
// Systems contain the business logic that processes entities in the ECS architecture.
type System interface {
	// Name returns the unique name of this system.
	Name() string
	// Handle processes the given command.
	//
	// Systems that do not handle the command should return ErrUnhandledCommand.
	// Tick execution is expressed via CommandKindTick.
	Handle(ctx context.Context, w World, cmd Command) error
}
