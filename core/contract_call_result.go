// | KB @CerberRus00 - Nexus Invest Team
package core

import (
	"GND/types"
)

// Селекторы записи storage (первые 4 байта keccak256(signature)):
// - setGndToken(address)  = 0x89d9b190, слот 0 (gndToken)
// - setGaniToken(address) = 0xc7d9116f, слот 1 (ganiToken)
// - setOwner(address)     = 0x13af4035, слот 2 (если есть в контракте)
const (
	selectorSetGndToken  = "\x89\xd9\xb1\x90"
	selectorSetGaniToken = "\xc7\xd9\x11\x6f"
	selectorSetOwner     = "\x13\xaf\x40\x35"
)

// storageWriter по calldata вызова контракта возвращает список изменений storage (адрес контракта = tx.Recipient).
type storageWriter func(tx *Transaction) []*types.StateChange

// writeStorageSelectors — таблица селектор → функция формирования storage changes для applyBlock.
var writeStorageSelectors = map[string]storageWriter{
	selectorSetGndToken:  writeSetGndToken,
	selectorSetGaniToken: writeSetGaniToken,
	selectorSetOwner:     writeSetOwner,
}

// writeSetGndToken: setGndToken(address) — слот 0 = gndToken (32 байта, right-padded).
func writeSetGndToken(tx *Transaction) []*types.StateChange {
	if tx == nil || len(tx.Data) < 36 {
		return nil
	}
	addr20 := tx.Data[16:36]
	slotKey := make([]byte, 32)
	slotValue := make([]byte, 32)
	copy(slotValue[12:], addr20)
	return []*types.StateChange{
		types.NewStorageChange(types.Address(tx.Recipient.String()), slotKey, slotValue),
	}
}

// writeSetGaniToken: setGaniToken(address) — слот 1 = ganiToken (32 байта, right-padded).
func writeSetGaniToken(tx *Transaction) []*types.StateChange {
	if tx == nil || len(tx.Data) < 36 {
		return nil
	}
	addr20 := tx.Data[16:36]
	slotKey := make([]byte, 32)
	slotKey[31] = 1
	slotValue := make([]byte, 32)
	copy(slotValue[12:], addr20)
	return []*types.StateChange{
		types.NewStorageChange(types.Address(tx.Recipient.String()), slotKey, slotValue),
	}
}

// writeSetOwner: setOwner(address) — слот 2 = address (32 байта, right-padded). Для контрактов с owner в storage.
func writeSetOwner(tx *Transaction) []*types.StateChange {
	if tx == nil || len(tx.Data) < 36 {
		return nil
	}
	addr20 := tx.Data[16:36]
	slotKey := make([]byte, 32)
	slotKey[31] = 2
	slotValue := make([]byte, 32)
	copy(slotValue[12:], addr20)
	return []*types.StateChange{
		types.NewStorageChange(types.Address(tx.Recipient.String()), slotKey, slotValue),
	}
}

// buildContractCallExecutionResult по calldata вызова контракта строит ExecutionResult с изменениями storage,
// чтобы ApplyExecutionResult записал слоты в contract_storage при SaveToDB.
// Поддерживаемые селекторы: setGndToken(address) — слот 0; setGaniToken(address) — слот 1; setOwner(address) — слот 2.
func buildContractCallExecutionResult(tx *Transaction) *types.ExecutionResult {
	if tx == nil || len(tx.Data) < 4 {
		return nil
	}
	selector := string(tx.Data[:4])
	fn, ok := writeStorageSelectors[selector]
	if !ok {
		return nil
	}
	changes := fn(tx)
	if len(changes) == 0 {
		return nil
	}
	return &types.ExecutionResult{
		GasUsed:      0,
		StateChanges: changes,
		ReturnData:   nil,
		Error:        nil,
	}
}
