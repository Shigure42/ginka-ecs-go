package ginka_ecs_go

import "time"

// CommandKind identifies the kind of command submitted to a World.
type CommandKind int

const (
	CommandKindAction CommandKind = iota + 1
	CommandKindTick
)

// Command is an external trigger (e.g. TCP request, cron tick, MQ message)
// that is handled by systems in registration order.
//
// When Kind is CommandKindAction, Payload contains the command data.
// When Kind is CommandKindTick, Dt is used and Payload is ignored.
type Command struct {
	Kind    CommandKind
	Payload any
	Dt      time.Duration
}
