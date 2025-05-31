package core

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv" // используйте стандартную функцию
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
	Miner        string         // Адрес валидатора/авторитета, создавшего блок (всегда с префиксом GND, GND_, GN, GN_)
	GasUsed      uint64         // Использованный gas (комиссия)
	GasLimit     uint64         // Лимит gas на блок
	Transactions []*Transaction // Список транзакций в блоке
	Consensus    string         // Тип консенсуса ("pos", "poa")
}

// NewBlock - конструктор блока
func NewBlock(index uint64, prevHash, miner string, txs []*Transaction, gasLimit uint64, consensus string) *Block {
	// ВАЖНО: miner должен быть адресом с допустимым префиксом!
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevHash,
		Miner:        miner, // не меняйте, не удаляйте префикс!
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
	sb.WriteString(strconv.FormatUint(b.Index, 10))
	sb.WriteString(b.PrevHash)
	sb.WriteString(strconv.FormatInt(b.Timestamp, 10))
	sb.WriteString(b.Miner)
	sb.WriteString(strconv.FormatUint(b.Nonce, 10))
	sb.WriteString(b.Consensus)
	for _, tx := range b.Transactions {
		sb.WriteString(tx.Hash)
	}
	sb.WriteString(strconv.FormatUint(b.GasLimit, 10))
	sb.WriteString(strconv.FormatUint(b.GasUsed, 10))
	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}
