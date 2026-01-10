package ginka_ecs_go

import (
	"context"
	"time"
)

// System represents a system that operates on entities with specific component sets.
// Systems contain the business logic that processes entities in the ECS architecture.
type System interface {
	// Name returns the unique name of this system.
	Name() string
}

// TickSystem is a system that runs periodically, typically once per frame or update cycle.
// The dt parameter represents the time elapsed since the last tick.
type TickSystem interface {
	System
	// Tick is called periodically to update the system.
	// The dt parameter is the delta time since the last tick.
	Tick(ctx context.Context, w World, dt time.Duration) error
}

// CommandSystem handles commands submitted to a World.
type CommandSystem interface {
	System
	Handle(ctx context.Context, w World, cmd Command) error
}

// CommandSubscriber declares which command types a system is interested in.
type CommandSubscriber interface {
	SubscribedCommands() []CommandType
}
