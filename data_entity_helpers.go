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
	out, ok, err := GetForUpdateE[T](ent, t)
	if err != nil {
		return zero, false
	}
	return out, ok
}

// GetForUpdateE returns the component of type t, marks it dirty if needed, and casts it to T.
// It also returns a non-nil error when the entity transaction fails.
func GetForUpdateE[T Component](ent DataEntity, t ComponentType) (T, bool, error) {
	var zero T
	if isNil(ent) {
		return zero, false, nil
	}
	if _, inTx := ent.(dataEntityTx); inTx {
		out, ok := getForUpdateDirect[T](ent, t)
		return out, ok, nil
	}

	out, ok, err := getForUpdateViaTx[T](ent, t)
	if err != nil {
		return zero, false, err
	}
	if ok {
		return out, true, nil
	}
	return zero, false, nil
}

func getForUpdateViaTx[T Component](ent DataEntity, t ComponentType) (T, bool, error) {
	var zero T
	var out T
	ok := false
	err := ent.Tx(func(tx DataEntity) error {
		typed, exists := getForUpdateDirect[T](tx, t)
		if !exists {
			return nil
		}
		out = typed
		ok = true
		return nil
	})
	if err != nil {
		return zero, false, err
	}
	if !ok {
		return zero, false, nil
	}
	return out, true, nil
}

func getForUpdateDirect[T Component](ent DataEntity, t ComponentType) (T, bool) {
	var zero T
	c, ok := ent.Get(t)
	if !ok {
		return zero, false
	}
	out, ok := c.(T)
	if !ok {
		return zero, false
	}
	if _, ok := ent.GetForUpdate(t); !ok {
		return zero, false
	}
	return out, true
}
