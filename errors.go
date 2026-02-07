package ginka_ecs_go

import "errors"

var (
	// ErrComponentAlreadyExists indicates an entity already has a component for the given ComponentType.
	ErrComponentAlreadyExists = errors.New("component already exists")
	// ErrComponentNotFound indicates an entity does not have a component for the given ComponentType.
	ErrComponentNotFound = errors.New("component not found")
	// ErrNilComponent indicates a nil component was provided.
	ErrNilComponent = errors.New("nil component")
	// ErrEntityAlreadyExists indicates the entity manager already contains an entity for the given id.
	ErrEntityAlreadyExists = errors.New("entity already exists")
	// ErrEntityNotFound indicates the entity manager does not contain an entity for the given id.
	ErrEntityNotFound = errors.New("entity not found")
	// ErrInvalidEntityId indicates an operation received an invalid entity id (e.g. empty string).
	ErrInvalidEntityId = errors.New("invalid entity id")
	// ErrWorldAlreadyRunning indicates an operation requires the world to be stopped.
	ErrWorldAlreadyRunning = errors.New("world already running")
)
