package ginka_ecs_go

import "time"

// NewAction creates a command with action payload.
func NewAction(payload any) Command {
	return Command{Kind: CommandKindAction, Payload: payload}
}

// NewTick creates a tick command with the provided delta time.
func NewTick(dt time.Duration) Command {
	return Command{Kind: CommandKindTick, Dt: dt}
}
