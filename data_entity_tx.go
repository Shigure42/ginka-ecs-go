package ginka_ecs_go

import "fmt"

// DataEntityTx exposes entity operations for use within DataEntity.Tx.
//
// Methods on this interface assume the entity lock is already held.
type DataEntityTx interface {
	Activatable
	Taggable

	// Id is the stable identifier of the entity (e.g. player id).
	Id() uint64
	// Name returns the human-readable name of the entity.
	Name() string
	// Type returns the category type of the entity.
	Type() EntityType

	// Has checks if the entity has a component of the given type.
	Has(t ComponentType) bool
	// Get retrieves a component by type, returning (nil, false) if not found.
	Get(t ComponentType) (Component, bool)
	// MustGet retrieves a component by type, panicking if not found.
	//
	// The panic value should wrap ErrComponentNotFound.
	MustGet(t ComponentType) Component

	// Add attaches c to the entity.
	//
	// If a component with the same ComponentType already exists, implementations
	// should return ErrComponentAlreadyExists (possibly wrapped).
	Add(c Component) error
	// RemoveComponent detaches the component for t and returns whether it existed.
	RemoveComponent(t ComponentType) bool
	// RemoveComponents detaches multiple components and returns the count of removed.
	RemoveComponents(types []ComponentType) int
	// AllComponents returns a copy of components in stable insertion order.
	AllComponents() []Component

	// GetData retrieves a data component by type.
	GetData(t ComponentType) (DataComponent, bool)
	// SetData attaches or replaces a data component.
	//
	// Implementations should mark the component type as dirty.
	SetData(c DataComponent) error
	// DirtyTypes returns a copy of dirty component types.
	DirtyTypes() []ComponentType
	// ClearDirty removes the dirty flag from the specified component types.
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
