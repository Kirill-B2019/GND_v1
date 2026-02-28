# Админское API GND_v1

**KB @CerbeRus - Nexus Invest Team**

Инструкция по настройке и использованию админских маршрутов: выдача и отзыв API-ключей, управление именами и ролями кошельков.

---

## 1. Назначение

- **API-ключи**: создание ключей через админский маршрут (ключ возвращается один раз), хранение в БД только хеша (SHA-256) и префикса; отзыв по id.
- **Кошельки**: просмотр списка кошельков с именами/ролями, обновление имени и роли по адресу (для системных кошельков: validator, treasury и т.д.).

Все админские маршруты защищены **отдельным секретом** `GND_ADMIN_SECRET`. Обычный `X-API-Key` для них не используется.

---

## 2. Переменные окружения

| Переменная           | Описание |
|----------------------|----------|
| `GND_ADMIN_SECRET`   | Секрет для заголовка `X-Admin-Token`. Если не задан, все запросы к `/api/v1/admin/*` возвращают 401. **Не коммитить в репозиторий.** |
| `GND_MASTER_KEY`     | Мастер-ключ для signing_service (шифрование приватных ключей). К админскому API не относится. |

Рекомендация: задавать `GND_ADMIN_SECRET` только на сервере (например в `.env` или systemd unit), не хранить в коде и не добавлять файлы с секретами в git (см. раздел «Безопасность и .gitignore»).

---

## 3. Миграции БД

Перед использованием админского API нужно применить миграции.

### 3.1. Расширение таблицы `api_keys`

Файл: `db/migrations/006_api_keys_admin.sql`

- Добавляются колонки: `name`, `key_prefix`, `key_hash`, `disabled`.
- Для новых ключей хранится только `key_hash` (SHA-256 в hex) и `key_prefix` (первые 8 символов для отображения); поле `key` может быть NULL.
- `disabled = true` — ключ отозван и не принимается при проверке.

Выполнение (от имени пользователя БД, например `gnduser`):

```bash
psql -U gnduser -d gnddb -f db/migrations/006_api_keys_admin.sql
```

### 3.2. Имена и роли кошельков

Файл: `db/migrations/007_wallets_name_role.sql`

- В таблицу `wallets` добавляются колонки: `name`, `role`.
- `role` — системная роль, например: `validator`, `treasury`, `fee_collector` или NULL.

Выполнение:

```bash
psql -U gnduser -d gnddb -f db/migrations/007_wallets_name_role.sql
```

После сброса БД скриптом `003_reset_database.sql` миграции 006 и 007 нужно применить заново (они идемпотентны за счёт `IF NOT EXISTS` / осторожного `DO $$ ... $$`).

---

## 4. Базовый URL и заголовок

- Базовый префикс: **`/api/v1/admin`**
- Все запросы требуют заголовок: **`X-Admin-Token: <значение GND_ADMIN_SECRET>`**
- При отсутствии или неверном токене возвращается **401 Unauthorized**.

Пример базового URL (при работе с нодой напрямую):

```text
http://localhost:8182/api/v1/admin
```

---

## 5. Маршруты

### 5.1. Создание API-ключа

**POST** `/api/v1/admin/keys`

Тело запроса (JSON):

| Поле          | Тип     | Обязательное | Описание |
|---------------|---------|--------------|----------|
| `name`        | string  | нет          | Человекочитаемое имя (например: "Laravel Backend"). По умолчанию: "API Key". |
| `permissions` | []string| нет          | Список прав/ролей (на будущее). По умолчанию: пустой массив. |
| `expires_at`  | string  | нет          | Срок действия в формате RFC3339 (например `2026-12-31T23:59:59Z`). Если не указан — без срока. |

Ответ **200** (ключ возвращается **один раз**):

```json
{
  "success": true,
  "data": {
    "id": 1,
    "key": "gnd_a1b2c3d4e5f6...",
    "name": "Laravel Backend",
    "key_prefix": "gnd_a1b2",
    "permissions": [],
    "expires_at": "",
    "created_at": "2026-02-28T12:00:00Z"
  }
}
```

Клиент должен сохранить `data.key` (полный ключ) и использовать его в заголовке `X-API-Key` для обычных эндпоинтов (создание кошелька, деплой токена и т.д.). В БД сохраняются только хеш и префикс; повторно получить полный ключ нельзя.

Пример (curl):

```bash
curl -X POST http://localhost:8182/api/v1/admin/keys \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: YOUR_ADMIN_SECRET" \
  -d '{"name":"My App","permissions":[]}'
```

---

### 5.2. Список API-ключей

**GET** `/api/v1/admin/keys`

Возвращает все ключи **без поля `key`** (только id, name, key_prefix, permissions, created_at, expires_at, disabled).

Ответ **200**:

```json
{
  "success": true,
  "data": {
    "keys": [
      {
        "id": 1,
        "name": "My App",
        "key_prefix": "gnd_a1b2",
        "permissions": [],
        "created_at": "2026-02-28T12:00:00Z",
        "expires_at": "",
        "disabled": false
      }
    ]
  }
}
```

---

### 5.3. Отзыв API-ключа

**POST** `/api/v1/admin/keys/:id/revoke`  
или  
**DELETE** `/api/v1/admin/keys/:id`

Устанавливает для ключа `disabled = true`. После этого ключ не принимается при проверке `X-API-Key`.

Ответ **200**:

```json
{ "success": true, "data": { "revoked": true } }
```

Ответ **404**: ключ с таким `id` не найден.

Пример:

```bash
curl -X POST http://localhost:8182/api/v1/admin/keys/1/revoke \
  -H "X-Admin-Token: YOUR_ADMIN_SECRET"
```

---

### 5.4. Список кошельков

**GET** `/api/v1/admin/wallets`

Опциональный query-параметр: `role` — фильтр по роли (например `validator`, `treasury`).

Ответ **200**:

```json
{
  "success": true,
  "data": {
    "wallets": [
      {
        "id": 1,
        "account_id": 1,
        "address": "GND...",
        "name": "Validator",
        "role": "validator",
        "signer_wallet_id": "",
        "created_at": "2026-02-28T10:00:00Z"
      }
    ]
  }
}
```

---

### 5.5. Обновление имени и/или роли кошелька

**PATCH** `/api/v1/admin/wallets/:address`

Тело (JSON): хотя бы одно из полей обязательно.

| Поле  | Тип    | Описание |
|-------|--------|----------|
| `name`| string | Человекочитаемое имя кошелька. |
| `role`| string | Системная роль: `validator`, `treasury`, `fee_collector` или другая; для сброса можно передать пустую строку (в зависимости от реализации). |

Ответ **200**:

```json
{ "success": true, "data": { "updated": true } }
```

Ответ **404**: кошелёк с указанным `address` не найден.

Пример:

```bash
curl -X PATCH "http://localhost:8182/api/v1/admin/wallets/GND..." \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: YOUR_ADMIN_SECRET" \
  -d '{"name":"Treasury","role":"treasury"}'
```

---

## 6. Проверка обычных API-ключей

При запросах к **неадминским** эндпоинтам (создание кошелька, деплой токена и т.д.) ключ передаётся в заголовке **X-API-Key**. Сервер проверяет его так:

1. Совпадение с константой (тестовый ключ), если задана.
2. По таблице `api_keys`: либо по полю `key` (старые ключи в открытом виде), либо по полю `key_hash` (SHA-256 от переданного ключа).
3. Условия: `disabled = false` (или NULL для старых записей), срок действия не истёк (`expires_at IS NULL OR expires_at > NOW()`).

Отозванные ключи перестают работать без перезапуска ноды.

---

## 7. Интеграция с Laravel (GND_admin)

В Laravel-приложении (панель администратора) рекомендуется:

1. В `.env` задать:
   - `GND_NODE_URL` — базовый URL ноды (например `http://localhost:8182`),
   - `GND_API_KEY` — обычный API-ключ для операций создания кошельков и т.д.,
   - `GND_ADMIN_SECRET` — секрет для вызова админских маршрутов (создание/отзыв ключей, обновление кошельков).

2. Для админских запросов отправлять заголовок:
   - `X-Admin-Token: config('services.gnd.admin_secret')`  
   или из `env('GND_ADMIN_SECRET')`.

3. Пример создания ключа из Laravel (HTTP-клиент):

```php
$response = Http::withHeaders([
    'X-Admin-Token' => config('services.gnd.admin_secret'),
    'Content-Type'  => 'application/json',
])->post(config('services.gnd.node_url') . '/api/v1/admin/keys', [
    'name'        => 'Laravel Backend',
    'permissions' => [],
]);
$data = $response->json('data');
// Один раз сохранить $data['key'] и показать пользователю; в БД Laravel можно хранить только key_prefix и id.
```

4. Список ключей и отзыв — GET и POST/DELETE к `/api/v1/admin/keys` и `/api/v1/admin/keys/:id/revoke` с тем же `X-Admin-Token`.

5. Список кошельков и обновление имени/роли — GET и PATCH к `/api/v1/admin/wallets` и `/api/v1/admin/wallets/:address` с `X-Admin-Token`.

Не храните `GND_ADMIN_SECRET` в репозитории и не логируйте его.

---

## 8. Безопасность и .gitignore

- Файлы с секретами не должны попадать в репозиторий. В корне проекта GND_v1 в `.gitignore` добавлено игнорирование файлов с секретами:
  - `*.secret`
  - `.env.local`
  - `.env.*.local`
  - `*.pem`
- Переменную `GND_ADMIN_SECRET` задавайте только через окружение или защищённый конфиг на сервере, без коммита в git.
- Админские маршруты не логируют тело запросов с ключами и не возвращают полный ключ повторно — только при создании в ответе `data.key`.

---

## 9. Краткая сводка маршрутов

| Метод   | Путь                          | Описание |
|---------|-------------------------------|----------|
| POST    | /api/v1/admin/keys            | Создать API-ключ (ключ в ответе один раз). |
| GET     | /api/v1/admin/keys            | Список ключей (без поля key). |
| POST    | /api/v1/admin/keys/:id/revoke | Отозвать ключ. |
| DELETE  | /api/v1/admin/keys/:id        | То же (отзыв). |
| GET     | /api/v1/admin/wallets         | Список кошельков (опционально ?role=). |
| PATCH   | /api/v1/admin/wallets/:address| Обновить name и/или role кошелька. |

Все маршруты требуют заголовок **X-Admin-Token** равный **GND_ADMIN_SECRET**.
