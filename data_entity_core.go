package ginka_ecs_go

import (
	"fmt"
)

// DataEntityCore is a DataEntity implementation with dirty tracking.
// Use Tx for consistent updates.
type DataEntityCore struct {
	*EntityCore
	dirty      map[ComponentType]struct{}
	dirtyTypes []ComponentType
}

// NewDataEntityCore creates a new DataEntityCore with the given parameters.
func NewDataEntityCore(id string, name string, typ EntityType, tags ...Tag) *DataEntityCore {
	return &DataEntityCore{
		EntityCore: NewEntityCore(id, name, typ, tags...),
		dirty:      make(map[ComponentType]struct{}),
		dirtyTypes: nil,
	}
}

// Tx executes fn with an exclusive lock for consistent updates.
func (e *DataEntityCore) Tx(fn func(tx DataEntity) error) error {
	if fn == nil {
		return fmt.Errorf("data entity tx: nil fn")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return fn(dataEntityTx{entity: e})
}

// GetForUpdate retrieves a component and marks it dirty if it is a DataComponent.
func (e *DataEntityCore) GetForUpdate(t ComponentType) (Component, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.getForUpdateUnlocked(t)
}

// DirtyTypes returns a copy of dirty component types.
func (e *DataEntityCore) DirtyTypes() []ComponentType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.dirtyTypesUnlocked()
}

// ClearDirty removes the dirty flag from the specified component types.
func (e *DataEntityCore) ClearDirty(types ...ComponentType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clearDirtyUnlocked(types...)
}

// RemoveComponent detaches a component by type and clears dirty state for that type.
func (e *DataEntityCore) RemoveComponent(t ComponentType) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.removeComponentUnlocked(t) {
		return false
	}
	e.clearDirtyUnlocked(t)
	return true
}

// RemoveComponents detaches multiple components by type and clears their dirty state.
func (e *DataEntityCore) RemoveComponents(types []ComponentType) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	removed := e.removeComponentsUnlocked(types)
	if removed == 0 {
		return 0
	}
	e.clearDirtyUnlocked(types...)
	return removed
}

func (e *DataEntityCore) getForUpdateUnlocked(t ComponentType) (Component, bool) {
	c, ok := e.getComponentUnlocked(t)
	if !ok {
		return nil, false
	}
	dc, ok := c.(DataComponent)
	if !ok {
		return c, true
	}
	dc.BumpVersion()
	e.markDirtyUnlocked(t)
	return c, true
}

func (e *DataEntityCore) dirtyTypesUnlocked() []ComponentType {
	if len(e.dirtyTypes) == 0 {
		return nil
	}
	out := make([]ComponentType, len(e.dirtyTypes))
	copy(out, e.dirtyTypes)
	return out
}

func (e *DataEntityCore) markDirtyUnlocked(t ComponentType) {
	if _, ok := e.dirty[t]; ok {
		return
	}
	e.dirty[t] = struct{}{}
	e.dirtyTypes = append(e.dirtyTypes, t)
}

func (e *DataEntityCore) clearDirtyUnlocked(types ...ComponentType) {
	if len(types) == 0 {
		clear(e.dirty)
		e.dirtyTypes = nil
		return
	}

	targets := make(map[ComponentType]struct{}, len(types))
	for _, t := range types {
		targets[t] = struct{}{}
	}

	removed := 0
	for t := range targets {
		if _, ok := e.dirty[t]; !ok {
			continue
		}
		delete(e.dirty, t)
		removed++
	}
	if removed == 0 {
		return
	}

	filtered := e.dirtyTypes[:0]
	for _, t := range e.dirtyTypes {
		if _, clearType := targets[t]; clearType {
			continue
		}
		filtered = append(filtered, t)
	}
	if len(filtered) == 0 {
		e.dirtyTypes = nil
	} else {
		e.dirtyTypes = filtered
	}
}

type dataEntityTx struct {
	entity *DataEntityCore
}

func (t dataEntityTx) Id() string {
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
	if !t.entity.removeComponentUnlocked(ct) {
		return false
	}
	t.entity.clearDirtyUnlocked(ct)
	return true
}

func (t dataEntityTx) RemoveComponents(types []ComponentType) int {
	removed := t.entity.removeComponentsUnlocked(types)
	if removed == 0 {
		return 0
	}
	t.entity.clearDirtyUnlocked(types...)
	return removed
}

func (t dataEntityTx) AllComponents() []Component {
	return t.entity.allComponentsUnlocked()
}

func (t dataEntityTx) DirtyTypes() []ComponentType {
	return t.entity.dirtyTypesUnlocked()
}

func (t dataEntityTx) ClearDirty(types ...ComponentType) {
	t.entity.clearDirtyUnlocked(types...)
}

func (t dataEntityTx) GetForUpdate(ct ComponentType) (Component, bool) {
	return t.entity.getForUpdateUnlocked(ct)
}

func (t dataEntityTx) Tx(fn func(tx DataEntity) error) error {
	return fmt.Errorf("data entity tx: nested tx not supported")
}

// Must satisfy DataEntity.
var _ DataEntity = (*DataEntityCore)(nil)
var _ Entity = (*DataEntityCore)(nil)
