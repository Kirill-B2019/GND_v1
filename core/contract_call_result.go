// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"GND/types"
)

// Селекторы записи storage (первые 4 байта keccak256(signature)):
// - setGaniToken(address) = 0xc7d9116f, слот 0
// - setOwner(address)     = 0x13af4035, слот 1
const (
	selectorSetGaniToken = "\xc7\xd9\x11\x6f"
	selectorSetOwner     = "\x13\xaf\x40\x35"
)

// storageWriter по calldata вызова контракта возвращает список изменений storage (адрес контракта = tx.Recipient).
type storageWriter func(tx *Transaction) []*types.StateChange

// writeStorageSelectors — таблица селектор → функция формирования storage changes для applyBlock.
var writeStorageSelectors = map[string]storageWriter{
	selectorSetGaniToken: writeSetGaniToken,
	selectorSetOwner:     writeSetOwner,
}

// writeSetGaniToken: setGaniToken(address) — слот 0 = address (32 байта, right-padded).
func writeSetGaniToken(tx *Transaction) []*types.StateChange {
	if tx == nil || len(tx.Data) < 36 {
		return nil
	}
	addr20 := tx.Data[16:36] // последние 20 байт первого аргумента ABI
	slotKey := make([]byte, 32)
	slotValue := make([]byte, 32)
	copy(slotValue[12:], addr20)
	return []*types.StateChange{
		types.NewStorageChange(types.Address(tx.Recipient.String()), slotKey, slotValue),
	}
}

// writeSetOwner: setOwner(address) — слот 1 = address (32 байта, right-padded).
func writeSetOwner(tx *Transaction) []*types.StateChange {
	if tx == nil || len(tx.Data) < 36 {
		return nil
	}
	addr20 := tx.Data[16:36]
	slotKey := make([]byte, 32)
	slotKey[31] = 1 // слот 1
	slotValue := make([]byte, 32)
	copy(slotValue[12:], addr20)
	return []*types.StateChange{
		types.NewStorageChange(types.Address(tx.Recipient.String()), slotKey, slotValue),
	}
}

// buildContractCallExecutionResult по calldata вызова контракта строит ExecutionResult с изменениями storage,
// чтобы ApplyExecutionResult записал слоты в contract_storage при SaveToDB.
// Поддерживаемые селекторы: setGaniToken(address) — слот 0; setOwner(address) — слот 1.
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
