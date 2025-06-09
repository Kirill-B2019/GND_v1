# Система событий и кэширования

## События

Система событий позволяет отслеживать различные действия в блокчейне, такие как переводы токенов, деплой контрактов и ошибки.

### Типы событий

- `Transfer` - перевод токенов
- `Approval` - одобрение расходования токенов
- `ContractDeployment` - деплой контракта
- `Error` - ошибка выполнения операции

### Использование

```go
// Создание менеджера событий
em := NewEventManager(pool)

// Подписка на события
em.Subscribe(EventTransfer, func(e Event) {
    // Обработка события
})

// Отправка события
em.Emit(Event{
    Type:     EventTransfer,
    Contract: contract,
    From:     from,
    To:       to,
    Amount:   amount,
})

// Получение событий из БД
events, err := em.GetEvents(ctx, contract, EventTransfer, 10)
```

## Кэширование

Система кэширования позволяет оптимизировать производительность за счет хранения часто используемых данных в памяти.

### Настройка кэша

```go
config := CacheConfig{
    MaxSize:         100,        // Максимальное количество элементов
    ExpirationTime:  time.Hour,  // Время жизни элемента
    CleanupInterval: time.Minute,// Интервал очистки
    BatchSize:       10,         // Размер пакета
}
cache := NewCache(config, pool)
```

### Использование кэша

```go
// Сохранение значения
cache.Set("key", value)

// Получение значения
if value, ok := cache.Get("key"); ok {
    // Использование значения
}
```

## Пакетная обработка

Система пакетной обработки позволяет оптимизировать работу с базой данных за счет группировки операций.

### Использование

```go
// Создание процессора
processor := func(batch []interface{}) error {
    // Обработка пакета
    return nil
}
bp := NewBatchProcessor(pool, 10, processor)

// Добавление элементов
bp.Add(item)

// Остановка процессора
bp.Stop()
```

## Структура базы данных

### Таблица events

```sql
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    contract VARCHAR(42) NOT NULL,
    from_address VARCHAR(42),
    to_address VARCHAR(42),
    amount NUMERIC,
    timestamp TIMESTAMP NOT NULL,
    tx_hash VARCHAR(66),
    error TEXT,
    metadata JSONB
);

CREATE INDEX idx_events_contract ON events(contract);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
```

## Примеры использования

### Отслеживание переводов

```go
// Подписка на события перевода
em.Subscribe(EventTransfer, func(e Event) {
    log.Printf("Transfer: %s -> %s: %s tokens",
        e.From, e.To, e.Amount.String())
})

// Получение истории переводов
events, err := em.GetEvents(ctx, contract, EventTransfer, 100)
if err != nil {
    log.Printf("Error getting transfer history: %v", err)
    return
}

for _, event := range events {
    log.Printf("Transfer at %s: %s -> %s: %s tokens",
        event.Timestamp, event.From, event.To, event.Amount.String())
}
```

### Кэширование балансов

```go
// Создание кэша для балансов
balanceCache := NewCache(CacheConfig{
    MaxSize:         1000,
    ExpirationTime:  time.Minute * 5,
    CleanupInterval: time.Minute,
}, pool)

// Получение баланса с кэшированием
func GetBalance(address string) (*big.Int, error) {
    if balance, ok := balanceCache.Get(address); ok {
        return balance.(*big.Int), nil
    }

    // Получение баланса из БД
    balance, err := getBalanceFromDB(address)
    if err != nil {
        return nil, err
    }

    // Сохранение в кэш
    balanceCache.Set(address, balance)
    return balance, nil
}
```

### Пакетная обработка транзакций

```go
// Создание процессора для транзакций
txProcessor := NewBatchProcessor(pool, 100, func(batch []interface{}) error {
    // Обработка пакета транзакций
    return processTransactions(batch)
})

// Добавление транзакций
for _, tx := range transactions {
    txProcessor.Add(tx)
}
``` 