// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"GND/types"
)

// Селектор setGaniToken(address) = первые 4 байта keccak256("setGaniToken(address)")
const selectorSetGaniToken = "\xc7\xd9\x11\x6f"

// buildContractCallExecutionResult по calldata вызова контракта строит ExecutionResult с изменениями storage,
// чтобы ApplyExecutionResult записал слоты в contract_storage при SaveToDB.
// Поддерживается setGaniToken(address): слот 0 контракта = адрес GANI.
func buildContractCallExecutionResult(tx *Transaction) *types.ExecutionResult {
	if tx == nil || len(tx.Data) < 36 {
		return nil
	}
	data := tx.Data
	if len(data) < 4 {
		return nil
	}
	selector := string(data[:4])
	var changes []*types.StateChange

	// setGaniToken(address): запись в storage слот 0 (ganiToken)
	if selector == selectorSetGaniToken {
		// ABI: 4 байта селектор + 32 байта address (правые 20 байт = адрес)
		addrParam := data[4:36]
		addr20 := addrParam[12:] // последние 20 байт
		// Слот 0, значение = 32 байта (12 нулей + 20 байт адреса)
		slotKey := make([]byte, 32) // слот 0
		slotValue := make([]byte, 32)
		copy(slotValue[12:], addr20)
		changes = append(changes, types.NewStorageChange(
			types.Address(tx.Recipient.String()),
			slotKey,
			slotValue,
		))
	}

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
