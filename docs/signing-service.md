# Signing Service (кастодиальные ключи)

Встроенный в GND сервис хранения и подписи ключей secp256k1 без выноса приватного ключа в открытом виде в таблицу `wallets`.

## Назначение

- При создании кошелька через API (POST `/api/v1/wallet`) при заданной переменной окружения `GND_MASTER_KEY` ключ генерируется и сохраняется **только** в таблице `signer_wallets` в зашифрованном виде (AES-256-GCM).
- В таблице `wallets` для таких кошельков поле `private_key` остаётся `NULL`, а связь с хранилищем ключей задаётся через `signer_wallet_id`.

## Конфигурация

- **GND_MASTER_KEY** (переменная окружения) — мастер-ключ шифрования в виде hex-строки ровно 64 символа (32 байта). Если не задан, новые кошельки создаются по старой схеме (приватный ключ в `wallets.private_key` в открытом виде).
- Пример генерации ключа: `openssl rand -hex 32`

## Схема БД

- **signer_wallets** (миграция `db/migrations/004_create_signer_wallets.sql`):
  - `id` (UUID, PK)
  - `account_id` (INTEGER, FK → accounts.id, UNIQUE)
  - `public_key` (BYTEA)
  - `encrypted_priv` (BYTEA)
  - `disabled` (BOOLEAN)
  - `created_at`, `updated_at`

- **wallets** (миграция `db/migrations/005_wallets_private_key_nullable.sql`):
  - `private_key` допускает NULL.
  - Добавлено поле `signer_wallet_id` (UUID, FK → signer_wallets.id).

## Очистка БД

- В `db/003_reset_database.sql` добавлена очистка таблицы `signer_wallets` (до очистки `accounts` из-за FK).

## Пакеты

| Пакет | Описание |
|-------|----------|
| **signing_service/crypto** | AES-GCM шифрование, secp256k1 (генерация, подпись, сериализация). |
| **signing_service/storage** | Модели и репозиторий Postgres для `signer_wallets`. |
| **signing_service/service** | SignerService: GenerateKeyForNewWallet, StoreWallet, SignDigest, GetPublicKey, CreateWallet. |

## Интеграция в GND

- В `main.go` при наличии `GND_MASTER_KEY` создаётся `SignerService` и передаётся в `api.StartRESTServer` как `SignerWalletCreator`.
- `core.Blockchain.CreateWallet(ctx)` при наличии `SignerCreator` вызывает `core.NewWalletWithSigner(ctx, pool, creator)`.
- Кошелёк валидатора/майнера по-прежнему загружается через `core.LoadWallet(pool)` — выбирается последний активный кошелёк с непустым `private_key` (без signer).

## Тесты

- `signing_service/crypto`: `encrypt_test.go` (LoadMasterKey, Encrypt/Decrypt, ZeroBytes), `secp256k1_test.go` (генерация ключа, подпись).
