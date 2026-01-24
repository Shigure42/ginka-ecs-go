package ginka_ecs_go

import (
	"fmt"
)

// DataEntityCore is a reusable DataEntity implementation.
//
// It is storage-agnostic and only tracks:
// - DataComponent attachment/mutation
// - Dirty component types
// - An entity version number (typically used for optimistic locking)
//
// Persistence (loading/flushing/version bumps) is expected to be implemented by
// the concrete World or an adapter layer.
type DataEntityCore struct {
	*EntityCore

	version    uint64
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

// GetData retrieves a data component by type.
func (e *DataEntityCore) GetData(t ComponentType) (DataComponent, bool) {
	c, ok := e.Get(t)
	if !ok {
		return nil, false
	}
	dc, ok := c.(DataComponent)
	return dc, ok
}

// SetData attaches or replaces a data component.
func (e *DataEntityCore) SetData(c DataComponent) error {
	if c == nil {
		return fmt.Errorf("set data: %w", ErrNilComponent)
	}
	t := c.ComponentType()

	// Replace existing data component of the same type.
	if existing, ok := e.Get(t); ok {
		if _, ok := existing.(DataComponent); !ok {
			return fmt.Errorf("set data %d: existing component is not a DataComponent", t)
		}
		// Keep insertion order stable: replace in-place.
		e.components[t] = c
		e.MarkDirty(t)
		return nil
	}
	if err := e.Add(c); err != nil {
		return err
	}
	e.MarkDirty(t)
	return nil
}

// MutateData applies fn to an existing data component.
func (e *DataEntityCore) MutateData(t ComponentType, fn func(c DataComponent) error) error {
	if fn == nil {
		return fmt.Errorf("mutate data %d: nil fn", t)
	}
	dc, ok := e.GetData(t)
	if !ok {
		return fmt.Errorf("mutate data %d: %w", t, ErrComponentNotFound)
	}
	if err := fn(dc); err != nil {
		return err
	}
	e.MarkDirty(t)
	return nil
}

// Version returns the current version of the entity data.
func (e *DataEntityCore) Version() uint64 {
	return e.version
}

// SetVersion sets the current version.
//
// This is not part of the DataEntity interface; it is intended for persistence layers.
func (e *DataEntityCore) SetVersion(v uint64) {
	e.version = v
}

// BumpVersion increments the version by 1.
//
// This is not part of the DataEntity interface; it is intended for persistence layers.
func (e *DataEntityCore) BumpVersion() uint64 {
	e.version++
	v := e.version
	return v
}

// DirtyTypes returns a copy of dirty component types.
func (e *DataEntityCore) DirtyTypes() []ComponentType {
	if len(e.dirtyTypes) == 0 {
		return nil
	}
	out := make([]ComponentType, len(e.dirtyTypes))
	copy(out, e.dirtyTypes)
	return out
}

// MarkDirty marks a component type as modified.
func (e *DataEntityCore) MarkDirty(t ComponentType) {
	if _, ok := e.dirty[t]; ok {
		return
	}
	e.dirty[t] = struct{}{}
	e.dirtyTypes = append(e.dirtyTypes, t)
}

// ClearDirty removes the dirty flag from the specified component types.
func (e *DataEntityCore) ClearDirty(types ...ComponentType) {
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

// ForEachDirtyType iterates dirty component types in stable mark order.
//
// It does not allocate.
func (e *DataEntityCore) ForEachDirtyType(fn func(t ComponentType) error) error {
	for _, t := range e.dirtyTypes {
		if _, ok := e.dirty[t]; !ok {
			continue
		}
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// Must satisfy DataEntity.
var _ DataEntity = (*DataEntityCore)(nil)
var _ Entity = (*DataEntityCore)(nil)
