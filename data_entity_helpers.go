package ginka_ecs_go

// Get returns the component of type t and casts it to T.
func Get[T Component](ent Entity, t ComponentType) (T, bool) {
	var zero T
	if isNil(ent) {
		return zero, false
	}
	c, ok := ent.Get(t)
	if !ok {
		return zero, false
	}
	out, ok := c.(T)
	return out, ok
}

// GetForUpdate returns the component of type t, marks it dirty if needed, and casts it to T.
func GetForUpdate[T Component](ent DataEntity, t ComponentType) (T, bool) {
	var zero T
	if isNil(ent) {
		return zero, false
	}
	c, ok := ent.GetForUpdate(t)
	if !ok {
		return zero, false
	}
	out, ok := c.(T)
	return out, ok
}
