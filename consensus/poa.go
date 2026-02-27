// | KB @CerbeRus - Nexus Invest Team
package consensus

import (
	"GND/core"
	"fmt"
	"time"
)

// Объявление типа PoA
type PoA struct {
	running bool
}

// Конструктор
func NewPoA() *PoA {
	return &PoA{}
}

// Запуск PoA-консенсуса
func (p *PoA) Start(bc *core.Blockchain, mempool *core.Mempool) {
	p.running = true
	go func() {
		for p.running {
			fmt.Println("[PoA] Авторитетный узел создает блок...")
			// Реализация PoA-логики
			time.Sleep(time.Second * 5)
		}
	}()
}

// Остановка
func (p *PoA) Stop() {
	p.running = false
}

// Тип консенсуса
func (p *PoA) Type() string {
	return "poa"
}
