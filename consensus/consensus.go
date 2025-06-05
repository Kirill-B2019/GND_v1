package consensus

import (
	"GND/core"
	"log"
	"sync"
)

// Consensus — интерфейс для всех алгоритмов консенсуса
type Consensus interface {
	Start(bc *core.Blockchain, mempool *core.Mempool)
	Stop()
	Type() string
}
type ConsensusProvider interface {
	Start()
	ValidateBlock(block *core.Block) bool
}

var posConfig *core.ConsensusPosConfig
var poaConfig *core.ConsensusPoaConfig

var CurrentProvider ConsensusProvider

func InitPosConsensus(cfg *core.ConsensusPosConfig) {
	if cfg == nil {
		log.Fatal("Конфиг PoS не может быть nil")
	}
	posConfig = cfg
	log.Printf("Инициализирован PoS-консенсус: %+v", cfg)
}

// GetPosConfig возвращает текущие настройки PoS
func GetPosConfig() *core.ConsensusPosConfig {
	return posConfig
}

func InitPoaConsensus(cfg *core.ConsensusPoaConfig) {
	if cfg == nil {
		log.Fatal("Конфиг PoA не может быть nil")
	}
	poaConfig = cfg
	log.Printf("Инициализирован PoA-консенсус: %+v", cfg)
}

func GetPoaConfig() *core.ConsensusPoaConfig {
	return poaConfig
}

// BaseConsensus — базовая структура с общими полями и методами
type BaseConsensus struct {
	bc      *core.Blockchain
	mempool *core.Mempool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// Start — запускает консенсус
func (c *BaseConsensus) Start(bc *core.Blockchain, mempool *core.Mempool) {
	c.bc = bc
	c.mempool = mempool
	c.stopCh = make(chan struct{})
	c.wg.Add(1)
	go c.run()
}

// Stop — останавливает консенсус
func (c *BaseConsensus) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

// run — заглушка, конкретные консенсусы должны переопределять этот метод
func (c *BaseConsensus) run() {
	// Заглушка
}

// Type — возвращает тип консенсуса (по умолчанию "base")
func (c *BaseConsensus) Type() string {
	return "base"
}

// Пример реализации PoS
type PoSConsensus struct {
	BaseConsensus
	// дополнительные поля для PoS
}

func (p *PoSConsensus) run() {
	for {
		select {
		case <-p.stopCh:
			p.wg.Done()
			return
		default:
			// Здесь логика PoS
		}
	}
}

func (p *PoSConsensus) Type() string {
	return "pos"
}

// Пример реализации PoA
type PoAConsensus struct {
	BaseConsensus
	// дополнительные поля для PoA
}

func (a *PoAConsensus) run() {
	for {
		select {
		case <-a.stopCh:
			a.wg.Done()
			return
		default:
			// Здесь логика PoA
		}
	}
}

func (a *PoAConsensus) Type() string {
	return "poa"
}
