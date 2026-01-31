package ginka_ecs_go

// DataComponent is a Component that can be persisted.
// It adds serialization and versioning.
type DataComponent interface {
	Component
	// StorageKey is the storage mapping (table name, key prefix, etc).
	StorageKey() string
	// Version is the current version number.
	Version() uint64
	// SetVersion updates the version number.
	SetVersion(uint64)
	// BumpVersion increments the version by 1.
	BumpVersion() uint64
	// Marshal serializes the component to bytes.
	Marshal() ([]byte, error)
	// Unmarshal deserializes from bytes.
	Unmarshal([]byte) error
}
