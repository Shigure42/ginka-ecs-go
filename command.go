package ginka_ecs_go

// CommandType identifies a kind of command submitted to a World.
type CommandType int

// Command is an external trigger (e.g. TCP request, cron tick, MQ message)
// that is routed by EntityId() and handled serially per entity.
type Command interface {
	// Type returns the type of this command.
	Type() CommandType
	// EntityId returns the target entity ID for this command.
	EntityId() uint64
}
