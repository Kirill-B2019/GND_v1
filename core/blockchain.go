package core

import (
	"errors"
	"fmt"
	"sync"
)

// Blockchain - основная структура цепочки блоков
type Blockchain struct {
	blocks []*Block
	State  *State
	mutex  sync.RWMutex
}

// NewBlockchain создает новую цепочку с генезис-блоком
func NewBlockchain(genesis *Block) *Blockchain {
	bc := &Blockchain{
		blocks: []*Block{genesis},
		State:  NewState(),
	}
	// Применяем генезис-блок к состоянию
	bc.applyBlock(genesis)
	return bc
}

// LatestBlock возвращает последний блок в цепочке
func (bc *Blockchain) LatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return bc.blocks[len(bc.blocks)-1]
}

// AddBlock добавляет новый блок в цепочку после проверки
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Проверка предыдущего хеша
	if block.PrevHash != bc.LatestBlock().Hash {
		return errors.New("invalid previous hash")
	}
	// Проверка корректности блока
	if !bc.validateBlock(block) {
		return errors.New("block validation failed")
	}
	// Добавление блока
	bc.blocks = append(bc.blocks, block)
	// Применение транзакций блока к состоянию
	bc.applyBlock(block)
	return nil
}

// validateBlock выполняет базовую валидацию блока
func (bc *Blockchain) validateBlock(block *Block) bool {
	// Проверка хеша блока
	if block.Hash != block.CalculateHash() {
		fmt.Println("Hash mismatch")
		return false
	}
	// TODO: Проверка подписи валидатора/авторитета
	// TODO: Проверка комиссий, времени, консенсусных правил
	// TODO: Проверка уникальности транзакций, double-spend
	return true
}

// applyBlock применяет все транзакции блока к состоянию
func (bc *Blockchain) applyBlock(block *Block) {
	for _, tx := range block.Transactions {
		ok := bc.State.ApplyTransaction(tx)
		if !ok {
			fmt.Printf("TX %s failed, skipped\n", tx.Hash)
		}
		// TODO: обработка вызова смарт-контрактов через VM
		// TODO: начисление комиссии майнеру/валидатору
	}
}

// GetBlockByHash ищет блок по хешу
func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	for _, b := range bc.blocks {
		if b.Hash == hash {
			return b
		}
	}
	return nil
}

// GetBlockByIndex возвращает блок по номеру
func (bc *Blockchain) GetBlockByIndex(idx uint64) *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	if idx < uint64(len(bc.blocks)) {
		return bc.blocks[idx]
	}
	return nil
}

// Height возвращает высоту цепочки
func (bc *Blockchain) Height() uint64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return uint64(len(bc.blocks) - 1)
}

// AllBlocks возвращает копию всех блоков (для API/обозревателя)
func (bc *Blockchain) AllBlocks() []*Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	blocksCopy := make([]*Block, len(bc.blocks))
	copy(blocksCopy, bc.blocks)
	return blocksCopy
}
