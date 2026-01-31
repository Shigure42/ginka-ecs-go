package ginka_ecs_go

import "fmt"

// DataEntityTx exposes entity operations for use within DataEntity.Tx.
// The entity lock is already held when these methods are called.
type DataEntityTx interface {
	Activatable
	Taggable

	// Id is the entity's stable identifier.
	Id() uint64
	// Name is the human-readable label.
	Name() string
	// Type categorizes the entity.
	Type() EntityType

	// Has reports whether the entity has a component of type t.
	Has(t ComponentType) bool
	// Get returns the component of type t, or (nil, false) if not present.
	Get(t ComponentType) (Component, bool)
	// MustGet returns the component of type t, panicking if missing.
	MustGet(t ComponentType) Component

	// Add attaches component c to the entity.
	// Returns ErrComponentAlreadyExists if a component of the same type exists.
	Add(c Component) error
	// RemoveComponent detaches the component of type t.
	// Returns true if a component was removed.
	RemoveComponent(t ComponentType) bool
	// RemoveComponents detaches multiple components.
	// Returns the count of components actually removed.
	RemoveComponents(types []ComponentType) int
	// AllComponents returns all attached components in insertion order.
	AllComponents() []Component

	// GetData fetches a data component by type.
	GetData(t ComponentType) (DataComponent, bool)
	// SetData attaches or replaces a data component and marks it dirty.
	SetData(c DataComponent) error
	// DirtyTypes returns the component types that have been modified.
	DirtyTypes() []ComponentType
	// ClearDirty clears the dirty flag from the given types.
	ClearDirty(types ...ComponentType)
}

type dataEntityTx struct {
	entity *DataEntityCore
}

func (t dataEntityTx) Id() uint64 {
	return t.entity.id
}

func (t dataEntityTx) Name() string {
	return t.entity.name
}

func (t dataEntityTx) Type() EntityType {
	return t.entity.typ
}

func (t dataEntityTx) Enabled() bool {
	return t.entity.enabledUnlocked()
}

func (t dataEntityTx) SetEnabled(enabled bool) {
	t.entity.setEnabledUnlocked(enabled)
}

func (t dataEntityTx) Tags() []Tag {
	return t.entity.tagsUnlocked()
}

func (t dataEntityTx) HasTag(tag Tag) bool {
	return t.entity.hasTagUnlocked(tag)
}

func (t dataEntityTx) AddTag(tag Tag) bool {
	return t.entity.addTagUnlocked(tag)
}

func (t dataEntityTx) RemoveTag(tag Tag) bool {
	return t.entity.removeTagUnlocked(tag)
}

func (t dataEntityTx) Has(ct ComponentType) bool {
	_, ok := t.entity.getComponentUnlocked(ct)
	return ok
}

func (t dataEntityTx) Get(ct ComponentType) (Component, bool) {
	return t.entity.getComponentUnlocked(ct)
}

func (t dataEntityTx) MustGet(ct ComponentType) Component {
	c, ok := t.entity.getComponentUnlocked(ct)
	if !ok {
		panic(fmt.Errorf("must get component %d: %w", ct, ErrComponentNotFound))
	}
	return c
}

func (t dataEntityTx) Add(c Component) error {
	return t.entity.addComponentUnlocked(c)
}

func (t dataEntityTx) RemoveComponent(ct ComponentType) bool {
	return t.entity.removeComponentUnlocked(ct)
}

func (t dataEntityTx) RemoveComponents(types []ComponentType) int {
	return t.entity.removeComponentsUnlocked(types)
}

func (t dataEntityTx) AllComponents() []Component {
	return t.entity.allComponentsUnlocked()
}

func (t dataEntityTx) GetData(ct ComponentType) (DataComponent, bool) {
	return t.entity.getDataUnlocked(ct)
}

func (t dataEntityTx) SetData(c DataComponent) error {
	return t.entity.setDataUnlocked(c)
}

func (t dataEntityTx) DirtyTypes() []ComponentType {
	return t.entity.dirtyTypesUnlocked()
}

func (t dataEntityTx) ClearDirty(types ...ComponentType) {
	t.entity.clearDirtyUnlocked(types...)
}

var _ DataEntityTx = (*dataEntityTx)(nil)
