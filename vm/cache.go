package vm

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"sync"
	"time"
)

// CacheConfig определяет настройки кэша
type CacheConfig struct {
	MaxSize         int           // Максимальное количество элементов
	ExpirationTime  time.Duration // Время жизни элемента
	CleanupInterval time.Duration // Интервал очистки
	BatchSize       int           // Размер пакета для пакетной обработки
}

// Cache представляет кэш с автоматическим обновлением
type Cache struct {
	config   CacheConfig
	items    map[string]cacheItem
	mu       sync.RWMutex
	pool     *pgxpool.Pool
	stopChan chan struct{}
}

type cacheItem struct {
	value      interface{}
	expiresAt  time.Time
	lastAccess time.Time
}

// NewCache создает новый кэш
func NewCache(config CacheConfig, pool *pgxpool.Pool) *Cache {
	cache := &Cache{
		config:   config,
		items:    make(map[string]cacheItem),
		pool:     pool,
		stopChan: make(chan struct{}),
	}

	// Запускаем очистку кэша
	go cache.cleanupLoop()

	return cache
}

// Get возвращает значение из кэша
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	// Обновляем время последнего доступа
	c.mu.Lock()
	item.lastAccess = time.Now()
	c.items[key] = item
	c.mu.Unlock()

	return item.value, true
}

// Set сохраняет значение в кэш
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Проверяем размер кэша
	if len(c.items) >= c.config.MaxSize {
		c.evictOldest()
	}

	c.items[key] = cacheItem{
		value:      value,
		expiresAt:  time.Now().Add(c.config.ExpirationTime),
		lastAccess: time.Now(),
	}
}

// evictOldest удаляет самый старый элемент из кэша
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.lastAccess
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// cleanupLoop периодически очищает устаревшие элементы
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

// cleanup удаляет устаревшие элементы
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}

// BatchProcessor обрабатывает транзакции пакетами
type BatchProcessor struct {
	pool      *pgxpool.Pool
	batchSize int
	queue     chan interface{}
	processor func([]interface{}) error
	stopChan  chan struct{}
}

// NewBatchProcessor создает новый процессор пакетов
func NewBatchProcessor(pool *pgxpool.Pool, batchSize int, processor func([]interface{}) error) *BatchProcessor {
	bp := &BatchProcessor{
		pool:      pool,
		batchSize: batchSize,
		queue:     make(chan interface{}, batchSize*2),
		processor: processor,
		stopChan:  make(chan struct{}),
	}

	go bp.processLoop()
	return bp
}

// Add добавляет элемент в очередь обработки
func (bp *BatchProcessor) Add(item interface{}) {
	select {
	case bp.queue <- item:
		// Элемент добавлен в очередь
	default:
		// Очередь переполнена, обрабатываем текущий пакет
		batch := make([]interface{}, 0, bp.batchSize)
		for i := 0; i < bp.batchSize; i++ {
			select {
			case item := <-bp.queue:
				batch = append(batch, item)
			default:
				break
			}
		}
		if len(batch) > 0 {
			bp.processBatch(batch)
		}
		bp.queue <- item
	}
}

// processLoop обрабатывает элементы в фоновом режиме
func (bp *BatchProcessor) processLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	batch := make([]interface{}, 0, bp.batchSize)

	for {
		select {
		case item := <-bp.queue:
			batch = append(batch, item)
			if len(batch) >= bp.batchSize {
				bp.processBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				bp.processBatch(batch)
				batch = batch[:0]
			}
		case <-bp.stopChan:
			if len(batch) > 0 {
				bp.processBatch(batch)
			}
			return
		}
	}
}

// processBatch обрабатывает пакет элементов
func (bp *BatchProcessor) processBatch(batch []interface{}) {
	if err := bp.processor(batch); err != nil {
		log.Printf("Error processing batch: %v", err)
	}
}

// Stop останавливает процессор
func (bp *BatchProcessor) Stop() {
	close(bp.stopChan)
}
