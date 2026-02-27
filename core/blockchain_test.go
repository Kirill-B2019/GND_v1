package core

import (
	"context"
	"testing"
	"time"

	"GND/types"
)

func TestGenesisTimestamp_InFirstPartitionRange(t *testing.T) {
	// Время генезиса должно попадать в партицию transactions_2025_06 (2025-06-01 .. 2025-07-01)
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if genesisTimestamp.Before(start) || !genesisTimestamp.Before(end) {
		t.Errorf("genesisTimestamp %v должен быть в [%v, %v)", genesisTimestamp, start, end)
	}
	if genesisTimestamp != start {
		t.Errorf("genesisTimestamp ожидался %v, получен %v", start, genesisTimestamp)
	}
}

func TestNewSystemTransaction_SetsBlockIDAndHash(t *testing.T) {
	ts := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	tx := newSystemTransaction(42, ts, "genesis", types.Address("GND_GENESIS"), types.Address("GNDaddr"), nil, "")
	if tx.BlockID != 42 {
		t.Errorf("BlockID ожидался 42, получен %d", tx.BlockID)
	}
	if tx.Hash == "" {
		t.Error("Hash не должен быть пустым")
	}
	if tx.Type != "genesis" || tx.Status != "confirmed" {
		t.Errorf("Type=%q Status=%q", tx.Type, tx.Status)
	}
}

func TestLoadTransactionsForBlock_WithNilPool_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	txs, err := LoadTransactionsForBlock(ctx, nil, 1)
	if err != nil {
		t.Fatalf("ожидался nil error при pool=nil: %v", err)
	}
	if txs != nil {
		t.Errorf("ожидался nil slice при pool=nil, получен %v", txs)
	}
}

func TestAddBlock_WithNilPool_AppendsBlock(t *testing.T) {
	genesis := &Block{
		Index:     0,
		Timestamp: time.Now(),
		Miner:     "miner",
		GasUsed:   0,
		GasLimit:  10_000_000,
		Consensus: "poa",
		Nonce:     0,
		Status:    "finalized",
	}
	genesis.Hash = genesis.CalculateHash()
	bc := NewBlockchain(genesis, nil)

	block := &Block{
		Index:        1,
		PrevHash:     genesis.Hash,
		Timestamp:    time.Now(),
		Miner:        "miner",
		GasUsed:      0,
		GasLimit:     10_000_000,
		Consensus:    "poa",
		Nonce:        0,
		Status:       "finalized",
		Transactions: []*Transaction{},
	}
	block.Hash = block.CalculateHash()

	err := bc.AddBlock(block)
	if err != nil {
		t.Fatalf("AddBlock с pool=nil не должен возвращать ошибку: %v", err)
	}
	if len(bc.Blocks) != 2 {
		t.Errorf("ожидалось 2 блока в цепи, получено %d", len(bc.Blocks))
	}
	if bc.Blocks[1] != block {
		t.Error("второй блок в цепи должен быть тот же указатель")
	}
	if block.TxCount != 0 {
		t.Errorf("TxCount без БД остаётся 0, получен %d", block.TxCount)
	}
}

func TestAddBlock_InvalidHash_ReturnsError(t *testing.T) {
	genesis := &Block{
		Index:     0,
		Timestamp: time.Now(),
		Miner:     "miner",
		GasUsed:   0,
		GasLimit:  10_000_000,
		Consensus: "poa",
		Nonce:     0,
		Status:    "finalized",
	}
	genesis.Hash = genesis.CalculateHash()
	bc := NewBlockchain(genesis, nil)

	block := &Block{
		Index:     1,
		PrevHash:  genesis.Hash,
		Timestamp: time.Now(),
		Miner:     "miner",
		GasUsed:   0,
		GasLimit:  10_000_000,
		Consensus: "poa",
		Nonce:     0,
		Status:    "finalized",
	}
	block.Hash = "invalid_hash_not_matching"

	err := bc.AddBlock(block)
	if err == nil {
		t.Fatal("AddBlock с неверным хешем должен вернуть ошибку")
	}
	if len(bc.Blocks) != 1 {
		t.Errorf("блок с неверным хешем не должен добавляться, в цепи %d блоков", len(bc.Blocks))
	}
}

func TestStoreBlock_WithNilPool_DoesNotPanic(t *testing.T) {
	bc := NewBlockchain(&Block{Index: 0, Hash: "genesis", Miner: "m", GasUsed: 0, GasLimit: 10_000_000, Consensus: "poa", Nonce: 0, Status: "finalized"}, nil)
	block := &Block{
		Index:        1,
		Hash:         "h1",
		PrevHash:     "genesis",
		Timestamp:    time.Now(),
		Miner:        "m",
		GasUsed:      0,
		GasLimit:     10_000_000,
		Consensus:    "poa",
		Nonce:        0,
		Status:       "finalized",
		Transactions: []*Transaction{},
	}
	block.Hash = block.CalculateHash()

	err := bc.storeBlock(block)
	if err != nil {
		t.Fatalf("storeBlock с pool=nil возвращает nil: %v", err)
	}
	// block.ID не заполняется при pool=nil
	if block.ID != 0 {
		t.Errorf("при pool=nil block.ID должен оставаться 0, получен %d", block.ID)
	}
}
