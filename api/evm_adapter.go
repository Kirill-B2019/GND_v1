// api/evm_adapter.go

package api

import (
	"GND/types"
	"GND/vm"
	"math/big"
)

// evmAdapter приводит *vm.EVM к types.EVMInterface (для деплоера токенов).
type evmAdapter struct{ evm *vm.EVM }

func (a *evmAdapter) DeployContract(
	from types.Address,
	bytecode []byte,
	meta types.ContractMeta,
	gasLimit uint64,
	gasPrice *big.Int,
	nonce uint64,
	signature []byte,
	totalSupply *big.Int,
) (string, error) {
	return a.evm.DeployContract(from, bytecode, meta, gasLimit, gasPrice, nonce, signature, totalSupply)
}

func (a *evmAdapter) CallContract(
	from, to types.Address,
	data []byte,
	gasLimit uint64,
	gasPrice, value *big.Int,
	signature []byte,
) ([]byte, error) {
	gp, val := uint64(0), uint64(0)
	if gasPrice != nil {
		gp = gasPrice.Uint64()
	}
	if value != nil {
		val = value.Uint64()
	}
	res, err := a.evm.CallContract(string(from), string(to), data, gasLimit, gp, val)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.ReturnData, nil
}

func (a *evmAdapter) GetBalance(addr types.Address) (*big.Int, error) {
	return a.evm.GetBalance(string(addr))
}

// newEVMAdapter возвращает types.EVMInterface для переданного *vm.EVM.
func newEVMAdapter(evm *vm.EVM) types.EVMInterface {
	if evm == nil {
		return nil
	}
	return &evmAdapter{evm: evm}
}
