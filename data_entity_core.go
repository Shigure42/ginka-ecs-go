package ginka_ecs_go

import (
	"fmt"
)

// DataEntityCore is a reusable DataEntity implementation.
//
// It is storage-agnostic and only tracks:
// - DataComponent attachment/mutation
// - Dirty component types
//
// Persistence (loading/flushing/component version bumps) is expected to be implemented by
// the concrete World or an adapter layer.
type DataEntityCore struct {
	*EntityCore
	dirty      map[ComponentType]struct{}
	dirtyTypes []ComponentType
}

// NewDataEntityCore creates a new DataEntityCore with the given parameters.
func NewDataEntityCore(id uint64, name string, typ EntityType, tags ...Tag) *DataEntityCore {
	return &DataEntityCore{
		EntityCore: NewEntityCore(id, name, typ, tags...),
		dirty:      make(map[ComponentType]struct{}),
		dirtyTypes: nil,
	}
}

// Tx executes fn with an exclusive lock for consistent updates.
func (e *DataEntityCore) Tx(fn func(tx DataEntityTx) error) error {
	if fn == nil {
		return fmt.Errorf("data entity tx: nil fn")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return fn(dataEntityTx{entity: e})
}

// GetData retrieves a data component by type.
func (e *DataEntityCore) GetData(t ComponentType) (DataComponent, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.getDataUnlocked(t)
}

// SetData attaches or replaces a data component.
func (e *DataEntityCore) SetData(c DataComponent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.setDataUnlocked(c)
}

// LoadData attaches a data component without bumping version or marking dirty.
//
// This is intended for hydration from persistence.
func (e *DataEntityCore) LoadData(c DataComponent) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.loadDataUnlocked(c)
}

// DirtyTypes returns a copy of dirty component types.
func (e *DataEntityCore) DirtyTypes() []ComponentType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.dirtyTypesUnlocked()
}

// DirtyDataComponents returns a copy of dirty data components in mark order.
func (e *DataEntityCore) DirtyDataComponents() []DataComponent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.dirtyDataComponentsUnlocked()
}

// ClearDirty removes the dirty flag from the specified component types.
func (e *DataEntityCore) ClearDirty(types ...ComponentType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clearDirtyUnlocked(types...)
}

func (e *DataEntityCore) getDataUnlocked(t ComponentType) (DataComponent, bool) {
	c, ok := e.getComponentUnlocked(t)
	if !ok {
		return nil, false
	}
	dc, ok := c.(DataComponent)
	return dc, ok
}

func (e *DataEntityCore) setDataUnlocked(c DataComponent) error {
	if isNil(c) {
		return fmt.Errorf("set data: %w", ErrNilComponent)
	}
	t := c.ComponentType()

	// Replace existing data component of the same type.
	if existing, ok := e.getComponentUnlocked(t); ok {
		existingData, ok := existing.(DataComponent)
		if !ok {
			return fmt.Errorf("set data %d: existing component is not a DataComponent", t)
		}
		// Keep insertion order stable: replace in-place.
		c.SetVersion(existingData.Version())
		c.BumpVersion()
		e.components[t] = c
		e.markDirtyUnlocked(t)
		return nil
	}
	if err := e.addComponentUnlocked(c); err != nil {
		return err
	}
	c.BumpVersion()
	e.markDirtyUnlocked(t)
	return nil
}

func (e *DataEntityCore) loadDataUnlocked(c DataComponent) error {
	if isNil(c) {
		return fmt.Errorf("load data: %w", ErrNilComponent)
	}
	t := c.ComponentType()

	if existing, ok := e.getComponentUnlocked(t); ok {
		if _, ok := existing.(DataComponent); !ok {
			return fmt.Errorf("load data %d: existing component is not a DataComponent", t)
		}
		e.components[t] = c
		return nil
	}
	return e.addComponentUnlocked(c)
}

func (e *DataEntityCore) dirtyTypesUnlocked() []ComponentType {
	if len(e.dirtyTypes) == 0 {
		return nil
	}
	out := make([]ComponentType, len(e.dirtyTypes))
	copy(out, e.dirtyTypes)
	return out
}

func (e *DataEntityCore) dirtyDataComponentsUnlocked() []DataComponent {
	if len(e.dirtyTypes) == 0 {
		return nil
	}
	out := make([]DataComponent, 0, len(e.dirtyTypes))
	for _, t := range e.dirtyTypes {
		c, ok := e.getDataUnlocked(t)
		if !ok {
			continue
		}
		out = append(out, c)
	}
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
	for _, t := range types {
		if _, ok := e.dirty[t]; !ok {
			continue
		}
		delete(e.dirty, t)
		for i, existing := range e.dirtyTypes {
			if existing == t {
				e.dirtyTypes = append(e.dirtyTypes[:i], e.dirtyTypes[i+1:]...)
				break
			}
		}
	}
}

// Must satisfy DataEntity.
var _ DataEntity = (*DataEntityCore)(nil)
var _ Entity = (*DataEntityCore)(nil)
