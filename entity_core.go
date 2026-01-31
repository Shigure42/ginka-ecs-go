package ginka_ecs_go

import (
	"fmt"
	"sync"
)

// EntityCore is a basic Entity implementation with enabled flag and tags.
// It stores one component per ComponentType.
type EntityCore struct {
	EnabledFlag
	TagSet

	mu sync.RWMutex

	id   uint64
	name string
	typ  EntityType

	components     map[ComponentType]Component
	componentTypes []ComponentType
}

func NewEntityCore(id uint64, name string, typ EntityType, tags ...Tag) *EntityCore {
	e := &EntityCore{
		id:             id,
		name:           name,
		typ:            typ,
		components:     make(map[ComponentType]Component),
		componentTypes: nil,
	}
	e.SetTags(tags...)
	return e
}

// Id returns the entity's unique identifier.
func (e *EntityCore) Id() uint64 {
	return e.id
}

// Name returns the entity's human-readable name.
func (e *EntityCore) Name() string {
	return e.name
}

// Type returns the entity's category type.
func (e *EntityCore) Type() EntityType {
	return e.typ
}

// Enabled returns whether this entity is enabled.
func (e *EntityCore) Enabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enabledUnlocked()
}

// SetEnabled enables or disables this entity.
func (e *EntityCore) SetEnabled(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.setEnabledUnlocked(enabled)
}

// Tags returns a copy of tags.
func (e *EntityCore) Tags() []Tag {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tagsUnlocked()
}

// HasTag checks if the tag exists.
func (e *EntityCore) HasTag(tag Tag) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.hasTagUnlocked(tag)
}

// AddTag adds a tag, returning true if the tag was added.
func (e *EntityCore) AddTag(tag Tag) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.addTagUnlocked(tag)
}

// RemoveTag removes a tag, returning true if the tag was removed.
func (e *EntityCore) RemoveTag(tag Tag) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.removeTagUnlocked(tag)
}

// ClearTags removes all tags.
func (e *EntityCore) ClearTags() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clearTagsUnlocked()
}

// SetTags replaces all tags with the provided set.
// Duplicates are removed, keeping the first occurrence.
func (e *EntityCore) SetTags(tags ...Tag) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.setTagsUnlocked(tags...)
}

// Has checks if the entity has a component of the given type.
func (e *EntityCore) Has(t ComponentType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.getComponentUnlocked(t)
	return ok
}

// Get retrieves a component by type.
func (e *EntityCore) Get(t ComponentType) (Component, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.getComponentUnlocked(t)
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
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.addComponentUnlocked(c)
}

// RemoveComponent detaches a component by type.
func (e *EntityCore) RemoveComponent(t ComponentType) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.removeComponentUnlocked(t)
}

// RemoveComponents detaches multiple components by type.
func (e *EntityCore) RemoveComponents(types []ComponentType) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.removeComponentsUnlocked(types)
}

// AllComponents returns a copy of the stored components.
func (e *EntityCore) AllComponents() []Component {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.allComponentsUnlocked()
}

func (e *EntityCore) enabledUnlocked() bool {
	return e.EnabledFlag.Enabled()
}

func (e *EntityCore) setEnabledUnlocked(enabled bool) {
	e.EnabledFlag.SetEnabled(enabled)
}

func (e *EntityCore) tagsUnlocked() []Tag {
	return e.TagSet.Tags()
}

func (e *EntityCore) hasTagUnlocked(tag Tag) bool {
	return e.TagSet.HasTag(tag)
}

func (e *EntityCore) addTagUnlocked(tag Tag) bool {
	return e.TagSet.AddTag(tag)
}

func (e *EntityCore) removeTagUnlocked(tag Tag) bool {
	return e.TagSet.RemoveTag(tag)
}

func (e *EntityCore) clearTagsUnlocked() {
	e.TagSet.ClearTags()
}

func (e *EntityCore) setTagsUnlocked(tags ...Tag) {
	e.TagSet.SetTags(tags...)
}

func (e *EntityCore) getComponentUnlocked(t ComponentType) (Component, bool) {
	c, ok := e.components[t]
	return c, ok
}

func (e *EntityCore) addComponentUnlocked(c Component) error {
	if isNil(c) {
		return ErrNilComponent
	}
	t := c.ComponentType()
	if _, ok := e.components[t]; ok {
		return ErrComponentAlreadyExists
	}
	if e.components == nil {
		e.components = make(map[ComponentType]Component)
	}
	e.components[t] = c
	e.componentTypes = append(e.componentTypes, t)
	return nil
}

func (e *EntityCore) removeComponentUnlocked(t ComponentType) bool {
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

func (e *EntityCore) removeComponentsUnlocked(types []ComponentType) int {
	if len(types) == 0 {
		return 0
	}
	count := 0
	for _, t := range types {
		if e.removeComponentUnlocked(t) {
			count++
		}
	}
	return count
}

func (e *EntityCore) allComponentsUnlocked() []Component {
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

// Compile-time interface checks.
var _ Entity = (*EntityCore)(nil)
