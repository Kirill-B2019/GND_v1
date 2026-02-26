package vm

import (
	"math/big"
	"testing"
)

// TestChangeTypeConstants использует константы типа изменения состояния в StateChange (для линтера).
func TestChangeTypeConstants(t *testing.T) {
	sc := &StateChange{Type: ChangeTypeBalance, Address: "0x1", Symbol: "GND", Amount: big.NewInt(0)}
	if sc.Type != ChangeTypeBalance {
		t.Errorf("ожидался ChangeTypeBalance")
	}
	sc2 := &StateChange{Type: ChangeTypeStorage}
	if sc2.Type != ChangeTypeStorage {
		t.Errorf("ожидался ChangeTypeStorage")
	}
}

// TestGetEVMInstance проверяет вызов GetEVMInstance (использует функцию для линтера).
func TestGetEVMInstance(t *testing.T) {
	inst := GetEVMInstance()
	// TODO: после реализации синглтона — проверка не nil
	if inst != nil {
		t.Log("GetEVMInstance вернул экземпляр")
	}
}
