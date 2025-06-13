# ГАНИМЕД (GND) - Блокчейн платформа

## Обзор

ГАНИМЕД - это высокопроизводительная блокчейн платформа, поддерживающая смарт-контракты и токены. Платформа использует гибридный консенсус (PoA/PoS) и предоставляет полный набор API для интеграции.

## Основные возможности

- Гибридный консенсус (PoA/PoS)
- Поддержка смарт-контрактов (EVM)
- Токены стандарта GNDST-1
- REST API, WebSocket API и RPC API
- Мониторинг и аналитика
- Автоматический деплой

## Требования

- Go 1.21+
- PostgreSQL 15+
- Node.js 18+ (для веб-интерфейса)
- Git

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/your-org/gnd.git
cd gnd
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте базу данных:
```bash
psql -U postgres -f db/console_21.sql
```

4. Соберите проект:
```bash
go build -o GND.exe
```

## Конфигурация

Основные настройки находятся в файле `config/servers.json`:

```json
{
  "server": {
    "rpc": {
      "rpc_addr": "0.0.0.0:8181",
      "name": "GND RPC"
    },
    "rest": {
      "host": "0.0.0.0",
      "port": 8182
    },
    "ws": {
      "ws_addr": "0.0.0.0:8183",
      "name": "GND WebSocket"
    }
  }
}
```

## Запуск

1. Запустите ноду:
```bash
./GND.exe
```

2. Проверьте статус:
```bash
curl http://localhost:8182/api/health
```

## API

### REST API
- Базовый URL: `https://api.gnd-net.com:8182/api/`
- Документация: [docs/api.md](docs/api.md)

### WebSocket API
- URL: `wss://api.gnd-net.com:8183/ws`
- Документация: [docs/websocket_api.md](docs/websocket_api.md)

### RPC API
- URL: `https://api.gnd-net.com:8181`
- Документация: [docs/api.md](docs/api.md)

## Деплой

### Автоматический деплой

1. Настройте сервер:
```bash
sudo mkdir -p /var/www/gnd-api
sudo mkdir -p /var/www/backups/gnd-api
sudo chown -R www-data:www-data /var/www/gnd-api
sudo chown -R www-data:www-data /var/www/backups/gnd-api
```

2. Клонируйте репозиторий:
```bash
cd /var/www/gnd-api
sudo -u www-data git clone <URL_РЕПОЗИТОРИЯ> .
```

3. Установите сервис:
```bash
sudo cp scripts/gnd-api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable gnd-api
```

4. Настройте Git hooks:
```bash
sudo -u www-data mkdir -p /var/www/gnd-api/.git/hooks
sudo -u www-data nano /var/www/gnd-api/.git/hooks/post-receive
```

5. Сделайте скрипт деплоя исполняемым:
```bash
sudo chmod +x /var/www/gnd-api/scripts/deploy.sh
sudo chown www-data:www-data /var/www/gnd-api/scripts/deploy.sh
```

Подробная документация по деплою: [docs/deployment.md](docs/deployment.md)

## Структура проекта

```
GND/
├── api/            # API серверы (REST, WebSocket, RPC)
├── audit/          # Аудит безопасности
├── cmd/            # CLI команды
├── config/         # Конфигурационные файлы
├── consensus/      # Реализация консенсуса
├── core/           # Ядро блокчейна
├── db/             # Миграции и схемы БД
├── docs/           # Документация
├── integration/    # Интеграционные тесты
├── monitoring/     # Мониторинг
├── scripts/        # Скрипты деплоя и утилиты
├── static/         # Статические файлы
├── tests/          # Тесты
├── tokens/         # Реализация токенов
├── types/          # Общие типы
├── ui/             # Веб-интерфейс
├── utils/          # Утилиты
└── vm/             # Виртуальная машина (EVM)
```

## Документация

- [Архитектура](docs/architecture.md)
- [API](docs/api.md)
- [Смарт-контракты](docs/smart_contracts.md)
- [Токены](docs/tokens.md)
- [Безопасность](docs/security.md)
- [Тестирование](docs/testing.md)
- [Интеграция](docs/integration.md)
- [Деплой](docs/deployment.md)

## Мониторинг

### Метрики
- Количество запросов
- Время ответа
- Ошибки
- Использование ресурсов
- Размер блокчейна
- Количество транзакций
- Газ

### Алерты
- Превышение лимитов
- Ошибки
- Замедление
- Аномалии
- Недоступность
- Атаки

## Безопасность

### Аудит
- Код
- Конфигурация
- Доступ
- Данные
- Сеть
- Инфраструктура

### Мониторинг
- Активность
- Аномалии
- Угрозы
- Инциденты
- Доступ
- Изменения

## Лицензия

MIT License

## Контакты

- Email: support@gnd-net.com
- Telegram: @gnd_support
- Discord: [GND Community](https://discord.gg/gnd)