package ginka_ecs_go

import "time"

// AsCommand extracts a typed payload from an action command.
// Returns ErrUnhandledCommand for non-action or type mismatch.
func AsCommand[T any](cmd Command) (T, error) {
	var zero T
	if cmd.Kind != CommandKindAction {
		return zero, ErrUnhandledCommand
	}
	payload, ok := cmd.Payload.(T)
	if !ok {
		return zero, ErrUnhandledCommand
	}
	return payload, nil
}

// TickEvent extracts the tick duration from a tick command.
// Returns ErrUnhandledCommand for non-tick commands.
func TickEvent(cmd Command) (time.Duration, error) {
	if cmd.Kind != CommandKindTick {
		return 0, ErrUnhandledCommand
	}
	return cmd.Dt, nil
}
