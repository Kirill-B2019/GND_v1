# Подробная инструкция админу: деплой контрактов GND-st1 через админку

Пошаговое развёртывание контроллера, GND и GANI (стандарт GND-st1) через админ-панель и настройка ноды. Опционально — RWA-токен (шаг 4а).

**Связанные документы:** порядок файлов и параметры — `tokens/standards/deploy_order/README.md`; инварианты — `tokens/standards/deploy_order/INVARIANTS.md`; операции и комплаенс — [SOP_TOKEN_OPERATIONS.md](SOP_TOKEN_OPERATIONS.md), [COMPLIANCE_KYC_RWA.md](COMPLIANCE_KYC_RWA.md).

---

## Предусловия

- Доступ в админ-панель (логин/пароль; при включённой 2FA — код).
- Нода GND_v1 запущена, API доступен.
- Кошелёк с балансом GND для газа (или настроен gndself_address).
- **Адрес владельца контроллера** — взять из `config/native_contracts.json` поле **gndself_address** (используется в конструкторе NativeTokensController). Транзакции setGndToken, setGaniToken, mintGANI, setKyc* должны отправляться с этого адреса.
- Исходники в репозитории: `01_NativeTokensController.sol`, `02_GNDToken.sol`, `03_GANIToken.sol`, при необходимости `04_GNDRWAToken.sol` (каталог `tokens/standards/deploy_order/`).

---

## Шаг 1. Вход и раздел контрактов

1. Открыть админ-панель в браузере, войти (при 2FA ввести код).
2. В меню перейти в раздел **«Контракты»** (Contracts).
3. Деплой выполнять через **«Создать контракт»** / **«Деплой»** (загрузка ABI/bytecode или выбор шаблона и отправка транзакции на ноду).

---

## Шаг 2. Деплой контроллера (01_NativeTokensController)

1. Нажать **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт: **NativeTokensController** (файл `01_NativeTokensController.sol`).
3. **Параметр конструктора:** **owner_** (address) — адрес владельца. Указать **gndself_address** из `config/native_contracts.json` (системный кошелёк ГАНИМЕД). С этого адреса в дальнейшем вызываются setGndToken, setGaniToken, mintGANI, setKycGnd, setKycGani.
4. Указать кошелёк для оплаты газа, отправить транзакцию деплоя.
5. После успеха сохранить **адрес контракта** (формат `GNDct` + 32 hex-символа). Обозначить как **ADDR_CONTROLLER**.

---

## Шаг 3. Деплой GND (02_GNDToken, GND-st1)

1. Открыть **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт: **GNDToken** (GND-st1), файл `02_GNDToken.sol`.
3. **Метаданные:** Name = `GND (Ganimed)`, Symbol = `GND`, Description = `Контракт GND (Ganimed) по стандарту GND-st1`, License = `CORP`, Metadata = `{"author":"KB - Nexus Team","version":"1.0"}`.
4. Параметры конструктора (порядок и типы по коду):
   - **initialSupply** (uint256): `1000000000000000000000000000` (1e27).
   - **bridgeAddress** (address): `0x0000000000000000000000000000000000000000` (или адрес моста, если есть). При bridge=0 вызов crossChainTransfer на GND будет ревертиться с "Bridge not set" до установки моста.
   - **controllerContract** (address): **ADDR_CONTROLLER** из шага 2.
5. Указать кошелёк для газа, отправить транзакцию деплоя.
6. Сохранить адрес контракта GND — **ADDR_GND**.

---

## Шаг 4. Деплой GANI (03_GANIToken, GND-st1)

1. Открыть **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт: **GANIToken** (GND-st1), файл `03_GANIToken.sol`.
3. **Метаданные:** Name = `GANI (Ganimed Governance)`, Symbol = `GANI`, Description = `Контракт GANI (Ganimed Governance) по стандарту GND-st1`, License = `CORP`, Metadata = `{"author":"KB - Nexus Team","version":"1.0"}`.
4. Параметры конструктора:
   - **controllerContract** (address): **ADDR_CONTROLLER** (тот же, что в шаге 2).
4. Отправить транзакцию деплоя.
5. Сохранить адрес контракта GANI — **ADDR_GANI**.

---

## Шаг 5. Привязка GND и GANI к контроллеру

1. В разделе **«Контракты»** открыть контракт по адресу **ADDR_CONTROLLER** (NativeTokensController).
2. Транзакции отправлять **от имени gndself_address** (owner).
3. Вызвать метод **setGndToken(address)** с аргументом **ADDR_GND**. Отправить транзакцию.
4. Вызвать метод **setGaniToken(address)** с аргументом **ADDR_GANI**. Отправить транзакцию.
5. Убедиться, что обе транзакции успешны.

**Важно:** setGndToken и setGaniToken вызываются **только один раз**; повторная смена адреса контрактом не допускается (revert TokenAlreadySet). После привязки владелец может вызывать `mintGANI(to, amount)` и `setKycGnd(user, status)`, `setKycGani(user, status)`.

---

## Шаг 6. Обновление конфига ноды

1. На сервере ноды открыть файл `GND_v1/config/native_contracts.json`.
2. Заполнить:
   - **gnd_contract_address**: ADDR_GND
   - **gani_contract_address**: ADDR_GANI
   - **gndself_address** — не менять (адрес владельца контроллера, задаётся при деплое контроллера)
   - при необходимости **fee_collector_address**
3. Пример:

```json
{
  "gnd_contract_address": "GNDct...",
  "gani_contract_address": "GNDct...",
  "fee_collector_address": "",
  "gndself_address": "GN_BSP..."
}
```

4. Сохранить файл, перезапустить ноду (или дождаться перечитывания конфига).

---

## Шаг 7. Регистрация в БД ноды

1. В админке (или напрямую в БД) зарегистрировать контракты и токены.
2. Для каждого контракта (Controller, GND, GANI) — запись в таблице **contracts**: адрес, владелец (адрес деплоера), тип (`controller` / `token`).
3. Для **GND** и **GANI** — запись в таблице **tokens**: привязка к contract_id, symbol `GND` / `GANI`, name, decimals (18 / 6), total_supply, **standard = `GND-st1`**.
4. При наличии в админке раздела «Синхронизация с нодой» — использовать его, указав ADDR_GND и ADDR_GANI и заполнив symbol/name/decimals/standard.

---

## Шаг 8. Включение KYC для адресов

Чтобы переводы GND/GANI работали, у отправителя должен быть KYC.

1. Открыть контракт **ADDR_CONTROLLER** в админке → вызов методов.
2. Вызвать **setKycGnd(user, status)**: user = адрес кошелька, status = true. Отправить транзакцию (от owner).
3. Вызвать **setKycGani(user, status)** для тех же или других адресов при необходимости.
4. Проверить на контрактах GND/GANI метод **isKycPassed(user)** — должен вернуть true.

---

## Шаг 9. Проверка работы

1. **Балансы:** Запросить балансы кошелька контроллера по ADDR_GND и ADDR_GANI — должны быть: GND = 1e27 (initialSupply), GANI = 20M (первая эмиссия FIRST_EMISSION; всего лимит 100M).
2. **Эмиссия GANI:** На контроллере вызвать **mintGANI(to, amount)** (to — тестовый адрес). Проверить рост баланса GANI у `to`.
3. **Перевод GND:** Для кошелька с KYC вызвать на контракте GND **transfer(to, amount)**. Проверить изменение балансов. Перевод на адрес `0x0` запрещён (revert).
4. **Дивиденды (опционально):** На GND от controller вызвать snapshot(), затем setSnapshotBalance(snapshotId, user, amount) и setDividendsPerShare(snapshotId, amount). От имени user вызвать claimDividends(snapshotId) и проверить перевод с controller на user.

---

## Шаг 4а. (Опционально) Деплой RWA-токена (04_GNDRWAToken)

Если нужен токен по стандарту RWA (GND-st1 + RWA):

1. Открыть **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт **GNDRWAToken** (файл `04_GNDRWAToken.sol`). Убедиться, что в каталоге deploy_order есть `IGNDRWA.sol`.
3. Параметры конструктора (порядок по коду): **controllerContract** (ADDR_CONTROLLER), **bridgeAddress**, **name**, **symbol**, **decimals**, **maxSupply** — по регламенту проекта.
4. Сохранить адрес RWA-контракта при необходимости; зарегистрировать в БД как токен с нужным standard.

---

## Краткая памятка порядка деплоя

| Шаг | Действие | Контракт | Параметры конструктора |
|-----|----------|----------|-------------------------|
| 1 | Деплой | NativeTokensController | **owner_** = gndself_address из config |
| 2 | Сохранить адрес | — | ADDR_CONTROLLER |
| 3 | Деплой | GNDToken (GND-st1) | Name: GND (Ganimed), Desc: Контракт GND (Ganimed) по стандарту GND-st1, License: CORP; initialSupply=1e27, bridge=0…0, controller=ADDR_CONTROLLER |
| 4 | Сохранить адрес | — | ADDR_GND |
| 5 | Деплой | GANIToken (GND-st1) | Name: GANI (Ganimed Governance), Desc: Контракт GANI (Ganimed Governance) по стандарту GND-st1, License: CORP; controller=ADDR_CONTROLLER |
| 6 | Сохранить адрес | — | ADDR_GANI |
| 7 | Вызовы на контроллере (от gndself) | setGndToken(ADDR_GND), setGaniToken(ADDR_GANI) — один раз | — |
| 8 | Конфиг ноды | native_contracts.json | gnd_contract_address, gani_contract_address |
| 9 | БД/админка | contracts + tokens | standard = GND-st1 |
| 10 | По необходимости | setKycGnd(user, true), setKycGani(user, true) | от owner |
