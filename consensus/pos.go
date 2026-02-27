// | KB @CerbeRus - Nexus Invest Team
package consensus

import (
	"GND/core"
	"fmt"
)

type PoS struct {
	running bool
}

func NewPoS() *PoS {
	return &PoS{}
}

func (p *PoS) Start(bc *core.Blockchain, mempool *core.Mempool) {
	p.running = true
	go func() {
		for p.running {
			// Логика выбора валидатора, создание и добавление блока
			fmt.Println("[PoS] Валидатор выбирается, блок создается...")
			// ...реализация PoS...
			// time.Sleep(time.Second * 5)
			break // demo, уберите break для реального цикла
		}
	}()
}

func (p *PoS) Stop() {
	p.running = false
}

func (p *PoS) Type() string {
	return "pos"
}
