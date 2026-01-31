package ginka_ecs_go

import "time"

// CommandKind indicates what kind of command this is.
type CommandKind int

const (
	CommandKindAction CommandKind = iota + 1
	CommandKindTick
)

// Command is an external event (TCP request, cron tick, MQ message, etc.)
// that gets dispatched to systems in registration order.
//
// If Kind is CommandKindAction, Payload contains the command data.
// If Kind is CommandKindTick, Dt contains the delta time and Payload is ignored.
type Command struct {
	Kind    CommandKind
	Payload any
	Dt      time.Duration
}
