package ginka_ecs_go

// DataEntity is an Entity with data management and versioning.
// It tracks dirty components for efficient persistence.
type DataEntity interface {
	Entity

	// GetData fetches a data component by type.
	GetData(t ComponentType) (DataComponent, bool)
	// SetData attaches or replaces a data component and marks it dirty.
	SetData(c DataComponent) error
	// DirtyTypes returns the component types that have been modified.
	DirtyTypes() []ComponentType
	// DirtyDataComponents returns the dirty components in mark order.
	DirtyDataComponents() []DataComponent
	// ClearDirty clears the dirty flag from the given types.
	ClearDirty(types ...ComponentType)
	// Tx runs fn with an exclusive lock.
	// The callback must use tx and not call entity methods (deadlock risk).
	Tx(fn func(tx DataEntityTx) error) error
}
