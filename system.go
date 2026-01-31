package ginka_ecs_go

import "context"

// System processes entities. A system is named and handles commands.
type System interface {
	// Name identifies the system.
	Name() string
	// Handle processes a command.
	// Return ErrUnhandledCommand if the system doesn't handle this command.
	Handle(ctx context.Context, w World, cmd Command) error
}
