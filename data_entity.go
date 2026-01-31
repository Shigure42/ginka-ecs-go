package ginka_ecs_go

// DataEntity extends Entity with data management and versioning capabilities.
// It provides dirty tracking for efficient persistence operations.
type DataEntity interface {
	Entity

	// GetData retrieves a data component by type.
	GetData(t ComponentType) (DataComponent, bool)
	// SetData attaches or replaces a data component.
	//
	// Implementations should mark the component type as dirty.
	SetData(c DataComponent) error
	// DirtyTypes returns a copy of dirty component types.
	DirtyTypes() []ComponentType
	// DirtyDataComponents returns a copy of dirty data components in mark order.
	DirtyDataComponents() []DataComponent
	// ClearDirty removes the dirty flag from the specified component types.
	ClearDirty(types ...ComponentType)
	// Tx executes fn with an exclusive lock for consistent updates.
	//
	// The callback must use the provided tx and must not call methods on the
	// original entity to avoid deadlocks.
	Tx(fn func(tx DataEntityTx) error) error
}
