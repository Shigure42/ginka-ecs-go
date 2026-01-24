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

// ShardedTickSystem is a tick system that is executed once per shard.
//
// When a CoreWorld tick is triggered (via TickOnce or the internal ticker), the
// world enqueues a tick request onto every shard worker. Each shard then invokes
// TickShard serially with other commands in that shard.
//
// Contract:
//   - TickShard is invoked exactly once per shard for a tick event.
//   - Implementations must only operate on data that belongs to the given shard.
//     Use ShardIndex(entityId, shardCount) to filter entities.
//   - TickShard may read and write entity/component state; CoreWorld guarantees it
//     will not run concurrently with command handling in the same shard.
type ShardedTickSystem interface {
	System
	TickShard(ctx context.Context, w World, dt time.Duration, shardIdx int, shardCount int) error
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
