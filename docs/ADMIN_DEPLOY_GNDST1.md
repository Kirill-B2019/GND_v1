# Подробная инструкция админу: деплой контрактов GND-st1 через админку

Пошаговое развёртывание контроллера, GND и GANI (стандарт GND-st1) через админ-панель и настройка ноды.

---

## Предусловия

- Доступ в админ-панель (логин/пароль; при включённой 2FA — код).
- Нода GND_v1 запущена, API доступен.
- Кошелёк с балансом GND для газа (или настроен gndself_address).
- Исходники в репозитории: `01_NativeTokensController.sol`, `02_GNDToken.sol`, `03_GANIToken.sol` (каталог `tokens/standards/deploy_order/`).

---

## Шаг 1. Вход и раздел контрактов

1. Открыть админ-панель в браузере, войти (при 2FA ввести код).
2. В меню перейти в раздел **«Контракты»** (Contracts).
3. Деплой выполнять через **«Создать контракт»** / **«Деплой»** (загрузка ABI/bytecode или выбор шаблона и отправка транзакции на ноду).

---

## Шаг 2. Деплой контроллера (01_NativeTokensController)

1. Нажать **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт: **NativeTokensController** (файл `01_NativeTokensController.sol`).
3. Параметры конструктора: **нет**.
4. Указать кошелёк для оплаты газа, отправить транзакцию деплоя.
5. После успеха сохранить **адрес контракта** (формат `GNDct` + 32 hex-символа). Обозначить как **ADDR_CONTROLLER**.

---

## Шаг 3. Деплой GND (02_GNDToken, GND-st1)

1. Открыть **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт: **GNDToken** (GND-st1), файл `02_GNDToken.sol`.
3. Параметры конструктора (порядок и типы по коду):
   - **initialSupply** (uint256): `1000000000000000000000000000` (1e27).
   - **bridgeAddress** (address): `0x0000000000000000000000000000000000000000` (или адрес моста, если есть).
   - **controllerContract** (address): **ADDR_CONTROLLER** из шага 2.
4. Указать кошелёк для газа, отправить транзакцию деплоя.
5. Сохранить адрес контракта GND — **ADDR_GND**.

---

## Шаг 4. Деплой GANI (03_GANIToken, GND-st1)

1. Открыть **«Создать контракт»** / **«Деплой»**.
2. Выбрать контракт: **GANIToken** (GND-st1), файл `03_GANIToken.sol`.
3. Параметры конструктора:
   - **controllerContract** (address): **ADDR_CONTROLLER** (тот же, что в шаге 3).
4. Отправить транзакцию деплоя.
5. Сохранить адрес контракта GANI — **ADDR_GANI**.

---

## Шаг 5. Привязка GND и GANI к контроллеру

1. В разделе **«Контракты»** открыть контракт по адресу **ADDR_CONTROLLER** (NativeTokensController).
2. Вызвать метод **setGndToken(address)** с аргументом **ADDR_GND**. Отправить транзакцию.
3. Вызвать метод **setGaniToken(address)** с аргументом **ADDR_GANI**. Отправить транзакцию.
4. Убедиться, что обе транзакции успешны.

После этого владелец может вызывать `mintGANI(to, amount)` и `setKycGnd(user, status)`, `setKycGani(user, status)`.

---

## Шаг 6. Обновление конфига ноды

1. На сервере ноды открыть файл `GND_v1/config/native_contracts.json`.
2. Заполнить:
   - **gnd_contract_address**: ADDR_GND
   - **gani_contract_address**: ADDR_GANI
   - при необходимости **fee_collector_address**, **gndself_address**
3. Пример:

```json
{
  "gnd_contract_address": "GNDct...",
  "gani_contract_address": "GNDct...",
  "fee_collector_address": "",
  "gndself_address": ""
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

1. **Балансы:** Запросить балансы кошелька контроллера по ADDR_GND и ADDR_GANI — должны быть начальные supply (1e27 GND и 100M*10^6 GANI).
2. **Эмиссия GANI:** На контроллере вызвать **mintGANI(to, amount)** (to — тестовый адрес). Проверить рост баланса GANI у `to`.
3. **Перевод GND:** Для кошелька с KYC вызвать на контракте GND **transfer(to, amount)**. Проверить изменение балансов.
4. **Дивиденды (опционально):** На GND от controller вызвать snapshot(), затем setSnapshotBalance(snapshotId, user, amount) и setDividendsPerShare(snapshotId, amount). От имени user вызвать claimDividends(snapshotId) и проверить перевод с controller на user.

---

## Краткая памятка порядка деплоя

| Шаг | Действие | Контракт | Параметры конструктора |
|-----|----------|----------|-------------------------|
| 1 | Деплой | NativeTokensController | — |
| 2 | Сохранить адрес | — | ADDR_CONTROLLER |
| 3 | Деплой | GNDToken (GND-st1) | initialSupply=1e27, bridge=0…0, controller=ADDR_CONTROLLER |
| 4 | Сохранить адрес | — | ADDR_GND |
| 5 | Деплой | GANIToken (GND-st1) | controller=ADDR_CONTROLLER |
| 6 | Сохранить адрес | — | ADDR_GANI |
| 7 | Вызовы на контроллере | setGndToken(ADDR_GND), setGaniToken(ADDR_GANI) | — |
| 8 | Конфиг ноды | native_contracts.json | gnd_contract_address, gani_contract_address |
| 9 | БД/админка | contracts + tokens | standard = GND-st1 |
| 10 | По необходимости | setKycGnd(user, true), setKycGani(user, true) | от owner |
