package ginka_ecs_go

import "fmt"

// EntityCore is a reusable Entity implementation.
//
// It provides Enabled/tag behavior and enforces at most one Component per
// ComponentType.
type EntityCore struct {
	EnabledFlag
	TagSet

	entityId   uint64
	entityName string
	entityType EntityType

	components     map[ComponentType]Component
	componentTypes []ComponentType
}

func (e *EntityCore) Enabled() bool {
	return e.EnabledFlag.Enabled()
}

func (e *EntityCore) SetEnabled(enabled bool) {
	e.EnabledFlag.SetEnabled(enabled)
}

func (e *EntityCore) Tags() []Tag {
	return e.TagSet.Tags()
}

func (e *EntityCore) HasTag(tag Tag) bool {
	return e.TagSet.HasTag(tag)
}

func (e *EntityCore) AddTag(tag Tag) bool {
	return e.TagSet.AddTag(tag)
}

func (e *EntityCore) RemoveTag(tag Tag) bool {
	return e.TagSet.RemoveTag(tag)
}

// NewEntityCore creates a new EntityCore with the given id, name, type, and tags.
func NewEntityCore(id uint64, name string, typ EntityType, tags ...Tag) *EntityCore {
	e := &EntityCore{
		entityId:       id,
		entityName:     name,
		entityType:     typ,
		components:     make(map[ComponentType]Component),
		componentTypes: nil,
	}
	e.SetTags(tags...)
	return e
}

// Id returns the entity's unique identifier.
func (e *EntityCore) Id() uint64 {
	return e.entityId
}

// Name returns the entity's human-readable name.
func (e *EntityCore) Name() string {
	return e.entityName
}

// Type returns the entity's category type.
func (e *EntityCore) Type() EntityType {
	return e.entityType
}

// Has checks if the entity has a component of the given type.
func (e *EntityCore) Has(t ComponentType) bool {
	_, ok := e.Get(t)
	return ok
}

// Get retrieves a component by type.
func (e *EntityCore) Get(t ComponentType) (Component, bool) {
	if len(e.components) == 0 {
		return nil, false
	}
	c, ok := e.components[t]
	return c, ok
}

// MustGet retrieves a component by type, panicking if not found.
func (e *EntityCore) MustGet(t ComponentType) Component {
	c, ok := e.Get(t)
	if !ok {
		panic(fmt.Errorf("must get component %d: %w", t, ErrComponentNotFound))
	}
	return c
}

// Add attaches a component to the entity.
func (e *EntityCore) Add(c Component) error {
	if isNil(c) {
		return ErrNilComponent
	}
	t := c.ComponentType()
	if _, ok := e.Get(t); ok {
		return ErrComponentAlreadyExists
	}
	if e.components == nil {
		e.components = make(map[ComponentType]Component)
	}
	e.components[t] = c
	e.componentTypes = append(e.componentTypes, t)
	return nil
}

// Remove detaches a component by type.
func (e *EntityCore) Remove(t ComponentType) bool {
	if len(e.components) == 0 {
		return false
	}
	_, ok := e.components[t]
	delete(e.components, t)
	if ok {
		for i, existing := range e.componentTypes {
			if existing == t {
				e.componentTypes = append(e.componentTypes[:i], e.componentTypes[i+1:]...)
				break
			}
		}
	}
	return ok
}

// Components returns a copy of the stored components.
func (e *EntityCore) Components() []Component {
	if len(e.componentTypes) == 0 {
		return nil
	}

	out := make([]Component, 0, len(e.componentTypes))
	for _, t := range e.componentTypes {
		if c, ok := e.components[t]; ok {
			out = append(out, c)
		}
	}
	return out
}

// ForEachComponent iterates components in stable insertion order.
//
// It does not allocate.
func (e *EntityCore) ForEachComponent(fn func(t ComponentType, c Component) error) error {
	for _, t := range e.componentTypes {
		c, ok := e.components[t]
		if !ok {
			continue
		}
		if err := fn(t, c); err != nil {
			return err
		}
	}
	return nil
}

// Compile-time interface checks.
var _ Activatable = (*EntityCore)(nil)
var _ Taggable = (*EntityCore)(nil)
var _ Entity = (*EntityCore)(nil)
