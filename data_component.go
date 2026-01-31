package ginka_ecs_go

// DataComponent extends Component with persistence capabilities.
// It adds serialization and versioning features for components that need to be saved.
type DataComponent interface {
	Component
	// StorageKey identifies the storage mapping of this component (e.g. table name or key prefix).
	StorageKey() string
	// Version returns the current version of the component data.
	Version() uint64
	// SetVersion sets the current version.
	SetVersion(uint64)
	// BumpVersion increments the version by 1.
	BumpVersion() uint64
	// Marshal serializes the component data to bytes.
	Marshal() ([]byte, error)
	// Unmarshal deserializes component data from bytes.
	Unmarshal([]byte) error
}
