package core

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// Block - структура блока в блокчейне ГАНИМЕД
type Block struct {
	Index        uint64         // Порядковый номер блока в цепочке
	Timestamp    int64          // Время создания блока (unix)
	PrevHash     string         // Хеш предыдущего блока
	Hash         string         // Хеш текущего блока
	Nonce        uint64         // Nonce для PoS/PoA/PoW (расширяемо)
	Miner        string         // Адрес валидатора/авторитета, создавшего блок
	GasUsed      uint64         // Использованный gas (комиссия)
	GasLimit     uint64         // Лимит gas на блок
	Transactions []*Transaction // Список транзакций в блоке
	Consensus    string         // Тип консенсуса ("pos", "poa")
}

// NewBlock - конструктор блока
func NewBlock(index uint64, prevHash, miner string, txs []*Transaction, gasLimit uint64, consensus string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevHash,
		Miner:        miner,
		Transactions: txs,
		GasLimit:     gasLimit,
		Consensus:    consensus,
	}
	block.Hash = block.CalculateHash()
	return block
}

// CalculateHash - вычисляет хеш блока
func (b *Block) CalculateHash() string {
	var sb strings.Builder
	sb.WriteString(string(b.Index))
	sb.WriteString(b.PrevHash)
	sb.WriteString(string(b.Timestamp))
	sb.WriteString(b.Miner)
	sb.WriteString(string(b.Nonce))
	sb.WriteString(b.Consensus)
	for _, tx := range b.Transactions {
		sb.WriteString(tx.Hash)
	}
	sb.WriteString(string(b.GasLimit))
	sb.WriteString(string(b.GasUsed))
	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}
