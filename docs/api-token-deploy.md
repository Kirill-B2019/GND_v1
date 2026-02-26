# Создание и регистрация токена через API (внешняя система, api-keys)

Документ входит в техническую документацию ГАНИМЕД. Примеры запросов: [api-requests.md](api-requests.md). Обзор токенов: [tokens.md](tokens.md).

## Общая логика

1. **Внешняя система** подключается к ноде по REST API (**main-node.gnd-net.com**, порт 8182) и идентифицируется заголовком **X-API-Key**.
2. Запрос **POST /api/v1/token/deploy** создаёт токен и регистрирует его в системе:
   - проверка API-ключа (константа из конфига или таблица `api_keys` в БД);
   - генерация байткода токена (имя, символ, decimals, total_supply);
   - деплой контракта через EVM (адрес контракта, списание комиссии в GND);
   - регистрация в in-memory реестре токенов и запись в БД (`contracts`, `tokens`);
   - начальный баланс владельца = total_supply.
3. Дальнейшие операции с токеном (transfer, approve, balance) выполняются по адресу контракта без обязательного API-ключа (в текущей реализации).

## Цепочка вызовов

```
Внешняя система
  → POST /api/v1/token/deploy + X-API-Key
  → api.ValidateAPIKey (константа или public.api_keys)
  → api.Server.DeployToken
  → deployer.DeployToken(ctx, TokenParams)
       → generateBytecode(name, symbol, decimals, totalSupply)
       → evm.DeployContract(...)  → адрес контракта
       → eventManager.Emit(Deploy)
       → registerToken(ctx, TokenInfo)
            → gndst1.NewGNDst1 + SetInitialBalance(owner, totalSupply)
            → registry.RegisterToken(addr, token)   // in-memory
            → INSERT contracts, INSERT tokens       // БД
  → ответ: { "success": true, "data": { "address", "name", "symbol", ... } }
```

## API-ключи

- **Заголовок:** `X-API-Key: <ключ>`.
- **Проверка:**
  - если ключ совпадает с константой из конфига (например для тестов) — доступ разрешён;
  - иначе запрос к таблице `public.api_keys` (поле `key`, учёт `expires_at`).
- Эндпоинты, требующие ключа в текущей реализации: **POST /api/v1/token/deploy**.

## Тело запроса POST /api/v1/token/deploy

| Поле          | Тип    | Обязательное | Описание                          |
|---------------|--------|--------------|-----------------------------------|
| name          | string | да           | Имя токена                        |
| symbol        | string | да           | Символ (например USDT)            |
| decimals      | number | да           | Знаков после запятой (обычно 18)  |
| total_supply  | string | да           | Общее количество (число как строка)|
| owner         | string | да           | Адрес владельца (получает total_supply) |
| standard      | string | нет          | Стандарт (по умолчанию GND-st1)   |

Пример:

```json
{
  "name": "Test Token",
  "symbol": "TST",
  "decimals": 18,
  "total_supply": "1000000000000000000000000",
  "owner": "GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz",
  "standard": "GND-st1"
}
```

## Коды ответов

| Код | Ситуация |
|-----|----------|
| 200 | Токен создан и зарегистрирован |
| 400 | Неверное тело запроса или ошибка деплоя (например, не реализована генерация байткода) |
| 401 | Нет заголовка X-API-Key или ключ неверный |
| 503 | Сервис деплоя токенов недоступен (EVM/deployer не передан при старте REST) |

## Связанные компоненты

- **api/auth.go** — `ValidateAPIKey`.
- **api/rest.go** — `DeployToken`, маршрут `POST /token/deploy`.
- **tokens/deployer** — `DeployToken`, `registerToken` (реестр + БД).
- **tokens/registry** — in-memory реестр; **db**: `contracts`, `tokens`.
