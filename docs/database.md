# Базы данных в блокчейне ГАНИМЕД

## Обзор

Блокчейн ГАНИМЕД использует несколько типов баз данных для хранения различных данных:
- LevelDB для блоков и транзакций
- Redis для кэширования
- PostgreSQL для индексов и метаданных

## LevelDB

### Структура
```go
type LevelDB struct {
    db     *leveldb.DB
    path   string
    cache  *cache.Cache
    logger *zap.Logger
}
```

### Хранение данных
- Блоки
  - Номер блока -> Блок
  - Хеш блока -> Номер блока
  - Последний блок

- Транзакции
  - Хеш транзакции -> Транзакция
  - Адрес -> Список транзакций
  - Статус транзакции

- Состояние
  - Адрес -> Баланс
  - Адрес -> Код контракта
  - Адрес -> Хранилище

### Оптимизация
- Кэширование
- Сжатие
- Индексация
- Предварительная загрузка

## Redis

### Структура
```go
type Redis struct {
    client *redis.Client
    config *Config
    logger *zap.Logger
}
```

### Кэширование
- Блоки
  - Последние блоки
  - Часто используемые блоки
  - Метаданные блоков

- Транзакции
  - Ожидающие транзакции
  - Часто используемые транзакции
  - Статусы транзакций

- Состояние
  - Активные аккаунты
  - Часто используемые контракты
  - Временные данные

### Оптимизация
- TTL
- LRU
- Сжатие
- Кластеризация

## PostgreSQL

### Структура
```go
type PostgreSQL struct {
    db     *sql.DB
    config *Config
    logger *zap.Logger
}
```

### Таблицы
- Блоки
  ```sql
  CREATE TABLE blocks (
      number BIGINT PRIMARY KEY,
      hash VARCHAR(66) UNIQUE,
      parent_hash VARCHAR(66),
      timestamp BIGINT,
      validator VARCHAR(42),
      transactions_count INTEGER,
      gas_used BIGINT,
      gas_limit BIGINT,
      size INTEGER,
      created_at TIMESTAMP
  );
  ```

- Транзакции
  ```sql
  CREATE TABLE transactions (
      hash VARCHAR(66) PRIMARY KEY,
      block_number BIGINT,
      block_hash VARCHAR(66),
      from_address VARCHAR(42),
      to_address VARCHAR(42),
      value NUMERIC,
      gas_price BIGINT,
      gas_limit BIGINT,
      gas_used BIGINT,
      nonce BIGINT,
      status INTEGER,
      created_at TIMESTAMP,
      FOREIGN KEY (block_number) REFERENCES blocks(number)
  );
  ```

- Аккаунты
  ```sql
  CREATE TABLE accounts (
      address VARCHAR(42) PRIMARY KEY,
      balance NUMERIC,
      nonce BIGINT,
      code_hash VARCHAR(66),
      created_at TIMESTAMP,
      updated_at TIMESTAMP
  );
  ```

- Контракты
  ```sql
  CREATE TABLE contracts (
      address VARCHAR(42) PRIMARY KEY,
      creator VARCHAR(42),
      creation_tx VARCHAR(66),
      creation_block BIGINT,
      code TEXT,
      abi TEXT,
      created_at TIMESTAMP,
      FOREIGN KEY (creation_block) REFERENCES blocks(number)
  );
  ```

### Индексы
- Блоки
  - number
  - hash
  - timestamp
  - validator

- Транзакции
  - hash
  - block_number
  - from_address
  - to_address
  - status

- Аккаунты
  - address
  - balance
  - nonce

- Контракты
  - address
  - creator
  - creation_block

## Оптимизация

### LevelDB
- Кэширование
- Сжатие
- Индексация
- Предварительная загрузка

### Redis
- TTL
- LRU
- Сжатие
- Кластеризация

### PostgreSQL
- Индексы
- Партиционирование
- Ваккум
- Анализ

## Безопасность

### Защита данных
- Шифрование
- Бэкапы
- Аудит
- Контроль доступа

### Мониторинг
- Метрики
- Логи
- Алерты
- Отчеты

## Масштабирование

### Горизонтальное
- Шардирование
- Репликация
- Балансировка
- Синхронизация

### Вертикальное
- Оптимизация
- Кэширование
- Индексация
- Партиционирование

## Примеры

### JavaScript
```javascript
const db = new GND.Database({
    leveldb: {
        path: './data/leveldb'
    },
    redis: {
        host: 'localhost',
        port: 6379
    },
    postgres: {
        host: 'localhost',
        port: 5432,
        database: 'gnd',
        user: 'gnd',
        password: 'password'
    }
});

// Сохранение блока
await db.saveBlock(block);

// Получение блока
const block = await db.getBlock(number);

// Сохранение транзакции
await db.saveTransaction(tx);

// Получение транзакции
const tx = await db.getTransaction(hash);
```

### Python
```python
from gnd import Database

db = Database(
    leveldb={
        'path': './data/leveldb'
    },
    redis={
        'host': 'localhost',
        'port': 6379
    },
    postgres={
        'host': 'localhost',
        'port': 5432,
        'database': 'gnd',
        'user': 'gnd',
        'password': 'password'
    }
)

# Сохранение блока
db.save_block(block)

# Получение блока
block = db.get_block(number)

# Сохранение транзакции
db.save_transaction(tx)

# Получение транзакции
tx = db.get_transaction(hash)
```

### Go
```go
import "github.com/gnd/database"

config := database.Config{
    LevelDB: database.LevelDBConfig{
        Path: "./data/leveldb",
    },
    Redis: database.RedisConfig{
        Host: "localhost",
        Port: 6379,
    },
    PostgreSQL: database.PostgreSQLConfig{
        Host:     "localhost",
        Port:     5432,
        Database: "gnd",
        User:     "gnd",
        Password: "password",
    },
}

db := database.New(config)

// Сохранение блока
err := db.SaveBlock(block)

// Получение блока
block, err := db.GetBlock(number)

// Сохранение транзакции
err := db.SaveTransaction(tx)

// Получение транзакции
tx, err := db.GetTransaction(hash)
```

## Интеграция

### SDK
- JavaScript/TypeScript
- Python
- Go
- Java

### Инструменты
- CLI
- GUI
- Мониторинг
- Аналитика

## Обновления

### Версионирование
- Семантическое версионирование
- Обратная совместимость
- Миграции
- Обновления

### Миграции
- Планирование
- Тестирование
- Резервное копирование
- Откат

## Мониторинг

### Метрики
- Размер данных
- Производительность
- Использование ресурсов
- Ошибки

### Алерты
- Переполнение
- Ошибки
- Замедление
- Аномалии

## Безопасность

### Аудит
- Код
- Конфигурация
- Доступ
- Данные

### Мониторинг
- Активность
- Аномалии
- Угрозы
- Инциденты

### Реагирование
- Обнаружение
- Анализ
- Устранение
- Профилактика 