package ginka_ecs_go

// DataEntity is an Entity with data management and versioning.
// It tracks dirty components for efficient persistence.
type DataEntity interface {
	Entity
	// DirtyTypes returns the component types that have been modified.
	DirtyTypes() []ComponentType
	// ClearDirty clears the dirty flag from the given types.
	ClearDirty(types ...ComponentType)
	// GetForUpdate fetches a component and marks it dirty if it is a DataComponent.
	GetForUpdate(t ComponentType) (Component, bool)
	// Tx runs fn with an exclusive lock.
	// The callback must use tx and not call entity methods (deadlock risk).
	Tx(fn func(tx DataEntity) error) error
}
