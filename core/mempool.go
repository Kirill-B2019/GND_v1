// | KB @CerberRus00 - Nexus Invest Team
package core

import (
	"errors"
	"log"
	"sync"
	"time"
)

type Mempool struct {
	txChan chan *Transaction
	mu     sync.RWMutex
	txMap  map[string]*Transaction
	logger *log.Logger
}

func NewMempool() *Mempool {
	return &Mempool{
		txChan: make(chan *Transaction, 10000), // буфер на 10 000 транзакций
		txMap:  make(map[string]*Transaction),
		logger: log.New(log.Writer(), "[Mempool] ", log.LstdFlags),
	}
}

func (m *Mempool) Add(tx *Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Проверяем, не существует ли уже такая транзакция
	if _, exists := m.txMap[tx.Hash]; exists {
		return errors.New("transaction already exists")
	}

	select {
	case m.txChan <- tx:
		m.txMap[tx.Hash] = tx
		m.logger.Printf("Transaction %s added to mempool", tx.Hash)
		return nil
	default:
		// Канал полон — всё равно кладём в txMap, чтобы TakePending (блок-продюсер) мог забрать транзакцию
		m.txMap[tx.Hash] = tx
		m.logger.Printf("Transaction %s added to mempool (map only, channel full)", tx.Hash)
		return nil
	}
}

// ErrSkip возвращается из Pop(), когда транзакция оставлена в мемпуле (contract_call забирает блок-продюсер).
var ErrSkip = errors.New("skip")

// PutBack возвращает транзакцию в мемпул после Pop() (например, для повторной постановки).
func (m *Mempool) PutBack(tx *Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txMap[tx.Hash] = tx
	select {
	case m.txChan <- tx:
	default:
		// канал полон — транзакция только в txMap
	}
}

func (m *Mempool) Pop() (*Transaction, error) {
	select {
	case tx := <-m.txChan:
		// contract_call не забираем — оставляем в мемпуле для блок-продюсера
		if tx.IsContractCall() {
			m.mu.Lock()
			_, stillInMap := m.txMap[tx.Hash]
			if stillInMap {
				select {
				case m.txChan <- tx:
				default:
				}
			}
			m.mu.Unlock()
			return nil, ErrSkip
		}
		m.mu.Lock()
		delete(m.txMap, tx.Hash)
		m.mu.Unlock()
		m.logger.Printf("Transaction %s popped from mempool", tx.Hash)
		return tx, nil
	case <-time.After(100 * time.Millisecond):
		return nil, errors.New("timeout")
	}
}

func (m *Mempool) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.txMap)
}

func (m *Mempool) Exists(txHash string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.txMap[txHash]
	return exists
}

// GetTransaction возвращает транзакцию по хешу
func (m *Mempool) GetTransaction(txHash string) (*Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if tx, exists := m.txMap[txHash]; exists {
		return tx, nil
	}
	return nil, errors.New("transaction not found")
}

// Clear очищает мемпул
func (m *Mempool) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Очищаем канал
	for len(m.txChan) > 0 {
		<-m.txChan
	}

	// Очищаем карту
	m.txMap = make(map[string]*Transaction)
	m.logger.Println("Mempool cleared")
}

// GetPendingTransactions возвращает список всех ожидающих транзакций
func (m *Mempool) GetPendingTransactions() []*Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txs := make([]*Transaction, 0, len(m.txMap))
	for _, tx := range m.txMap {
		txs = append(txs, tx)
	}
	return txs
}

// TakePending забирает до max транзакций из мемпула и удаляет их из очереди (для включения в блок).
// Возвращённые транзакции больше не будут в мемпуле и не попадут в Pop().
func (m *Mempool) TakePending(max int) []*Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	if max <= 0 || len(m.txMap) == 0 {
		return nil
	}
	taken := make([]*Transaction, 0, max)
	takenSet := make(map[string]bool)
	for hash, tx := range m.txMap {
		if len(taken) >= max {
			break
		}
		taken = append(taken, tx)
		takenSet[hash] = true
	}
	for h := range takenSet {
		delete(m.txMap, h)
	}
	// Очищаем канал от взятых: вычитываем всё, обратно кладём только те, что не взяты
	var toKeep []*Transaction
	for {
		select {
		case tx := <-m.txChan:
			if !takenSet[tx.Hash] {
				toKeep = append(toKeep, tx)
			}
		default:
			goto done
		}
	}
done:
	for _, tx := range toKeep {
		select {
		case m.txChan <- tx:
		default:
			m.txMap[tx.Hash] = tx
		}
	}
	if len(taken) > 0 {
		m.logger.Printf("TakePending: взято %d транзакций в блок", len(taken))
	}
	return taken
}
