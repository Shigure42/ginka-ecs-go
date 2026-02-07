package ginka_ecs_go

import (
	"errors"
	"testing"
)

const testDataComponentType ComponentType = 10001

type testDataComponent struct {
	DataComponentCore
}

type txFailDataEntity struct {
	*DataEntityCore
	getForUpdateCalls int
}

func (e *txFailDataEntity) Tx(fn func(tx DataEntity) error) error {
	return errors.New("tx denied")
}

func (e *txFailDataEntity) GetForUpdate(t ComponentType) (Component, bool) {
	e.getForUpdateCalls++
	return e.DataEntityCore.GetForUpdate(t)
}

func newTestDataComponent() *testDataComponent {
	return &testDataComponent{DataComponentCore: NewDataComponentCore(testDataComponentType)}
}

func (c *testDataComponent) StorageKey() string {
	return "test"
}

func TestGetForUpdate_TypeMismatchDoesNotDirtyOrBump(t *testing.T) {
	ent := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := ent.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	_, ok := GetForUpdate[*ComponentCore](ent, testDataComponentType)
	if ok {
		t.Fatalf("expected type mismatch")
	}
	if component.Version() != 0 {
		t.Fatalf("version bumped on mismatch: %d", component.Version())
	}
	if len(ent.DirtyTypes()) != 0 {
		t.Fatalf("dirty types should remain empty on mismatch")
	}
}

func TestGetForUpdate_SuccessMarksDirtyAndBumps(t *testing.T) {
	ent := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := ent.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	out, ok := GetForUpdate[*testDataComponent](ent, testDataComponentType)
	if !ok {
		t.Fatalf("expected successful typed update")
	}
	if out != component {
		t.Fatalf("unexpected component returned")
	}
	if component.Version() != 1 {
		t.Fatalf("expected version 1, got %d", component.Version())
	}
	dirty := ent.DirtyTypes()
	if len(dirty) != 1 || dirty[0] != testDataComponentType {
		t.Fatalf("unexpected dirty types: %v", dirty)
	}
}

func TestDataEntity_RemoveComponentClearsDirty(t *testing.T) {
	ent := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := ent.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	if _, ok := GetForUpdate[*testDataComponent](ent, testDataComponentType); !ok {
		t.Fatalf("expected successful typed update")
	}
	if !ent.RemoveComponent(testDataComponentType) {
		t.Fatalf("expected remove component success")
	}
	if len(ent.DirtyTypes()) != 0 {
		t.Fatalf("dirty types should be cleared after remove")
	}
}

func TestDataEntity_RemoveComponentsClearsDirty(t *testing.T) {
	ent := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := ent.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	if _, ok := GetForUpdate[*testDataComponent](ent, testDataComponentType); !ok {
		t.Fatalf("expected successful typed update")
	}
	if removed := ent.RemoveComponents([]ComponentType{testDataComponentType}); removed != 1 {
		t.Fatalf("expected one removed component, got %d", removed)
	}
	if len(ent.DirtyTypes()) != 0 {
		t.Fatalf("dirty types should be cleared after remove components")
	}
}

func TestGetForUpdate_InsideTx(t *testing.T) {
	ent := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := ent.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	err := ent.Tx(func(tx DataEntity) error {
		_, ok := GetForUpdate[*testDataComponent](tx, testDataComponentType)
		if !ok {
			t.Fatalf("expected successful typed update in tx")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx failed: %v", err)
	}
	if component.Version() != 1 {
		t.Fatalf("expected version 1, got %d", component.Version())
	}
}

func TestGetForUpdate_TxErrorDoesNotFallbackToDirectUpdate(t *testing.T) {
	base := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := base.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}
	ent := &txFailDataEntity{DataEntityCore: base}

	_, ok := GetForUpdate[*testDataComponent](ent, testDataComponentType)
	if ok {
		t.Fatalf("expected failed update when tx returns error")
	}
	if ent.getForUpdateCalls != 0 {
		t.Fatalf("expected no direct get-for-update fallback, got %d calls", ent.getForUpdateCalls)
	}
	if component.Version() != 0 {
		t.Fatalf("expected unchanged version, got %d", component.Version())
	}
}

func TestGetForUpdateE_ReturnsTxError(t *testing.T) {
	base := NewDataEntityCore("1", "entity", 1)
	component := newTestDataComponent()
	if err := base.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}
	ent := &txFailDataEntity{DataEntityCore: base}

	_, ok, err := GetForUpdateE[*testDataComponent](ent, testDataComponentType)
	if err == nil {
		t.Fatalf("expected tx error")
	}
	if ok {
		t.Fatalf("expected update to fail")
	}
	if ent.getForUpdateCalls != 0 {
		t.Fatalf("expected no direct get-for-update fallback, got %d calls", ent.getForUpdateCalls)
	}
}
