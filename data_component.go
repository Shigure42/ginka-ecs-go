package ginka_ecs_go

// DataComponent extends Component with persistence capabilities.
// It adds serialization and versioning features for components that need to be saved.
type DataComponent interface {
	Component
	// PersistKey identifies the persistence mapping of this component (e.g. table name or key).
	PersistKey() string
	// Marshal serializes the component data to bytes.
	Marshal() ([]byte, error)
	// Unmarshal deserializes component data from bytes.
	Unmarshal([]byte) error
}

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
	// MutateData applies fn to an existing data component.
	//
	// Implementations should mark the component type as dirty if fn succeeds.
	MutateData(t ComponentType, fn func(c DataComponent) error) error

	// Version returns the current version of the entity data (commonly used for optimistic locking).
	Version() uint64
	// DirtyTypes returns a copy of dirty component types.
	DirtyTypes() []ComponentType
	// MarkDirty marks a component type as modified.
	MarkDirty(t ComponentType)
	// ClearDirty removes the dirty flag from the specified component types.
	ClearDirty(types ...ComponentType)
}
