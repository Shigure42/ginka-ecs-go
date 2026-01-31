package ginka_ecs_go

import "fmt"

// UpdateData loads an existing data component, applies fn, and stores it back.
//
// It uses Tx to ensure consistent updates and to bump component version.
func UpdateData(ent DataEntity, t ComponentType, fn func(c DataComponent) error) error {
	if isNil(ent) {
		return fmt.Errorf("update data %d: nil entity", t)
	}
	if fn == nil {
		return fmt.Errorf("update data %d: nil fn", t)
	}
	return ent.Tx(func(tx DataEntityTx) error {
		c, ok := tx.GetData(t)
		if !ok {
			return fmt.Errorf("update data %d: %w", t, ErrComponentNotFound)
		}
		if err := fn(c); err != nil {
			return err
		}
		return tx.SetData(c)
	})
}
