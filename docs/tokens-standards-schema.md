# Итоговая схема стандартов токенов ГАНИМЕД

Документ описывает все стандарты токенов, интерфейсы, вызываемые функции, управляющие контракты и ограничения доступа.

---

## 1. Общая иерархия

```
INativeCoin (базовый ERC-20 для L1)
    ├── IGND   (+ name, symbol, decimals)
    │       → GNDCoinBase   (precompile-обёртка)
    │       → GNDToken      (деплой, с контроллером)
    └── IGANI  (+ name, symbol, decimals)
            → GANICoinBase  (precompile-обёртка)
            → GANIToken     (деплой, с контроллером)

IGNDst1 (ERC-20 + KYC, snapshot, dividends, crossChain, modules)
    └── GNDst1Token (gndst1Base.sol)

IGNDRWA (ERC-20 + mint/burn + пауза, заморозка, cap + KYC, snapshot/дивиденды, модули)
    └── GNDRWAToken (GND-RWA.sol)
```

**Пояснение:** Нативные монеты (GND, GANI) имеют интерфейс L1 и контракты-обёртки; при деплое используются контракты с контроллером. GNDst-1 и GND-RWA — отдельные стандарты с единым контрактом-контроллером на токен.

---

## 2. Нативные монеты (L1): INativeCoin → IGND / IGANI

### 2.1. Назначение

| Элемент | Описание |
|--------|----------|
| **INativeCoin** | Общий интерфейс нативных монет L1. Источник истины — состояние ноды (`native_balances`), а не storage контракта. |
| **IGND** | Утилитарная монета GND (комиссии, стейкинг, награды). Макс. предложение 1e27 (1 млрд, 18 decimals). |
| **IGANI** | Governance-монета GANI. Фиксированное предложение 100M при 6 decimals. |

### 2.2. Реализации

| Реализация | Роль |
|------------|------|
| **GNDCoinBase / GANICoinBase** | Обёртки для precompile: вызовы из контракта делают `revert` с указанием использовать L1/precompile. Реальных балансов в контракте нет. |
| **GNDToken / GANIToken** | Деплоируемые контракты с балансами в storage. Начальная эмиссия в конструкторе; для GANI возможен mint только с адреса контроллера. |

### 2.3. Управляющий контракт: NativeTokensController

- **Файл:** `tokens/standards/deploy_order/01_NativeTokensController.sol`
- **Деплой:** первый (шаг 1). Адрес передаётся в конструкторы GNDToken и GANIToken как `controller`.

| Функция | Кто вызывает | Назначение |
|---------|----------------|------------|
| `setGaniToken(address)` | owner | Один раз задаёт адрес контракта GANI. |
| `mintGANI(address to, uint256 amount)` | owner | Выпуск GANI на адрес `to` (вызов `mint` на GANIToken). |

**Ограничение:** Контракты GND/GANI требуют, чтобы `controller` был контрактом (`extcodesize > 0`). EOA в качестве контроллера не допускается.

### 2.4. Вызываемые функции (интерфейс INativeCoin / IGND / IGANI)

| Функция | Тип | Примечание |
|---------|-----|------------|
| `totalSupply()` | view | В Base — revert (precompile); в Token — из storage. |
| `balanceOf(address)` | view | В Base — revert; в Token — из storage. |
| `transfer(to, amount)` | write | В Base — revert; в Token — обычный перевод. |
| `approve(spender, amount)` | write | Аналогично. |
| `allowance(owner, spender)` | view | Аналогично. |
| `transferFrom(from, to, amount)` | write | Аналогично. |
| `name()`, `symbol()`, `decimals()` | view | Только IGND/IGANI. |

**Ограничения по вызовам:** У GANIToken функцию `mint` может вызывать только `controller` (фактически NativeTokensController после `setGaniToken`). У GNDToken mint отключён — эмиссия только в конструкторе.

---

## 3. GNDst-1: IGNDst1 → GNDst1Token

### 3.1. Назначение

Мультистандартный токен: ERC-20/TRC-20 плюс KYC, снимки балансов, дивиденды, кросс-чейн, модули. Управление привилегированными функциями — только через контракт-контроллер.

### 3.2. Управляющий контракт (контроллер)

Отдельный контракт в репозитории не задан. В конструктор GNDst1Token передаётся произвольный адрес контракта — это и есть **контроллер**. Контроллер обязан быть контрактом. Модули регистрируются контроллером через `registerModule` и вызываются через `moduleCall` (каталог модулей: `tokens/standards/gndst1/modules/`).

### 3.3. Вызываемые функции и ограничения

| Функция | Кто может | Пояснение |
|---------|-----------|-----------|
| `totalSupply()`, `balanceOf()`, `allowance()` | любой | Чтение. |
| `approve()` | любой | Без ограничений. |
| `transfer()`, `transferFrom()` | любой с KYC | Модификатор `onlyKyc`. |
| `crossChainTransfer(targetChain, to, amount)` | любой с KYC | Перевод на bridge + событие. |
| `setKycStatus(user, status)` | только контроллер | Включение/выключение KYC. |
| `isKycPassed(user)` | любой | view. |
| `snapshot()` | только контроллер | Создание снимка. |
| `getSnapshotBalance(user, snapshotId)` | любой | view. |
| `setSnapshotBalance(snapshotId, user, amount)` | только контроллер | В контракте, не в интерфейсе. |
| `setDividendsPerShare(snapshotId, amount)` | только контроллер | В контракте, не в интерфейсе. |
| `claimDividends(snapshotId)` | любой | Получение дивидендов по снимку. |
| `registerModule(moduleId, moduleAddress, name)` | только контроллер | Регистрация модуля. |
| `moduleCall(moduleId, data)` | любой | Вызов зарегистрированного модуля. |
| `mint()` | никто | Всегда revert «MintingDisabled». |

**События:** Transfer, Approval, CrossChainTransfer, KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered.

---

## 4. GND-RWA: IGNDRWA → GNDRWAToken

### 4.1. Назначение

Токен реальных активов (RWA). ERC-20 с mint/burn только от контроллера. Расширения: пауза (переводы, mint, burn), заморозка адресов, лимит эмиссии (cap); **дополнительно** как в GNDst-1: **KYC** (KycStatusChanged), **снимки и дивиденды** (SnapshotCreated, DividendClaimed), **модули** (ModuleRegistered, ModuleCall).

### 4.2. Управляющий контракт (контроллер)

В конструктор передаётся адрес контракта управления (`controllerContract`) — это **контроллер**. Он должен быть контрактом. Реализация контроллера в репозитории может быть дополнена позже.

### 4.3. Вызываемые функции и ограничения

| Функция | Кто может | Пояснение |
|---------|-----------|-----------|
| **ERC-20** | | |
| `totalSupply()`, `balanceOf()`, `allowance()` | любой | view. |
| `transfer()`, `transferFrom()` | любой с KYC | При паузе и заморозке проверки как в п. ниже; модификатор onlyKyc. |
| `approve()` | любой | Без KYC. |
| **Эмиссия** | | |
| `mint(to, amount)` | только контроллер | При `mintPaused == false`; при `maxSupply > 0` — не выше лимита. |
| `burn(from, amount)` | только контроллер | При `burnPaused == false`. |
| **Пауза** | | |
| `setTransfersPaused(bool)` | только контроллер | Вкл/выкл паузу переводов. |
| `setMintPaused(bool)`, `setBurnPaused(bool)` | только контроллер | Аналогично. |
| `transfersPaused()`, `mintPaused()`, `burnPaused()` | любой | view. |
| **Заморозка** | | |
| `setFrozen(account, frozen)` | только контроллер | Заморозка/разморозка адреса. |
| `isFrozen(account)` | любой | view. |
| **KYC** | | |
| `setKycStatus(user, status)` | только контроллер | Включение/выключение KYC. Событие KycStatusChanged. |
| `isKycPassed(user)` | любой | view. |
| **Снимки и дивиденды** | | |
| `snapshot()` | только контроллер | Создание снимка. Событие SnapshotCreated. |
| `getSnapshotBalance(user, snapshotId)` | любой | view. |
| `setSnapshotBalance(snapshotId, user, amount)` | только контроллер | В контракте (не в интерфейсе). |
| `setDividendsPerShare(snapshotId, amount)` | только контроллер | В контракте (не в интерфейсе). |
| `claimDividends(snapshotId)` | любой | Получение дивидендов. Событие DividendClaimed. |
| **Модули** | | |
| `registerModule(moduleId, moduleAddress, name)` | только контроллер | Регистрация модуля. Событие ModuleRegistered. |
| `moduleCall(moduleId, data)` | любой | Вызов зарегистрированного модуля. Событие ModuleCall. |

**Конструктор:** `(controllerContract, name_, symbol_, decimals_, maxSupply_)`. Параметр `maxSupply_ = 0` — без лимита эмиссии.

**События:** Transfer, Approval, Mint, Burn, TransfersPausedChanged, MintPausedChanged, BurnPausedChanged, FrozenChanged, **KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered, ModuleCall**.

---

## 5. Сводные таблицы

### 5.1. Управляющие контракты

| Стандарт / токен | Управляющий контракт | Задаётся при деплое | Кто может mint |
|------------------|----------------------|----------------------|-----------------|
| GND (GNDToken) | NativeTokensController | controller = адрес 01_NativeTokensController | Никто (только конструктор). |
| GANI (GANIToken) | NativeTokensController | controller = тот же | Только controller (через mintGANI в контроллере). |
| GNDst1Token | Любой контракт (внешний) | controller в конструкторе | Никто. |
| GNDRWAToken | Контракт (дописывается при необходимости) | controller в конструкторе | Только controller. |

### 5.2. Ограничения по ролям

| Роль | Нативные (GND/GANI) | GNDst-1 | GND-RWA |
|------|--------------------|---------|---------|
| Нода / precompile | Источник истины для балансов (при использовании Base как интерфейса к L1) | — | — |
| Owner контроллера (NativeTokensController) | setGaniToken, mintGANI | — | — |
| Controller (контракт) | Для GND/GANI — адрес с начальным supply; для GANI — единственный источник mint | setKycStatus, snapshot, setSnapshotBalance, setDividendsPerShare, registerModule | mint, burn, set*Paused, setFrozen, setKycStatus, snapshot, setSnapshotBalance, setDividendsPerShare, registerModule |
| Любой с KYC (GNDst-1 и GND-RWA) | — | transfer, transferFrom, crossChainTransfer | transfer, transferFrom (при !paused и !frozen) |
| Любой | Переводы по балансам (Token) | approve, claimDividends, moduleCall, чтение | approve, claimDividends, moduleCall, чтение (с учётом паузы и заморозки) |

### 5.3. События по стандартам

| Стандарт | События |
|----------|---------|
| INativeCoin / IGND / IGANI | Transfer, Approval |
| IGNDst1 | Transfer, Approval, CrossChainTransfer, KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered |
| IGNDRWA | Transfer, Approval, Mint, Burn, TransfersPausedChanged, MintPausedChanged, BurnPausedChanged, FrozenChanged, KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered, ModuleCall |

### 5.4. Расположение файлов

| Стандарт | Интерфейс | Реализация | Управляющий контракт |
|----------|-----------|------------|----------------------|
| Нативные | native/INativeCoin.sol, IGND.sol, IGANI.sol | native/GNDCoinBase.sol, GANICoinBase.sol; native/GNDToken.sol, GANIToken.sol | deploy_order/01_NativeTokensController.sol |
| GNDst-1 | gndst1/IGNDst1.sol | gndst1/gndst1Base.sol | Внешний (адрес в конструкторе); модули — gndst1/modules/ |
| GND-RWA | gndrwa/IGNDRWA.sol | gndrwa/GND-RWA.sol | Внешний (адрес в конструкторе) |

---

## 6. Краткие пояснения

- **Нативные монеты:** Реальные балансы и переводы обрабатываются протоколом L1 (native_balances). Контракты Base — только интерфейс к precompile; контракты Token — для сценариев деплоя с контроллером и единой точкой эмиссии GANI.
- **GNDst-1:** Один контроллер управляет KYC, снимками, дивидендами и модулями. Переводы и crossChain доступны только адресам с пройденным KYC. Минт отключён.
- **GND-RWA:** Один контроллер управляет mint, burn, паузой, заморозкой, KYC, снимками/дивидендами и регистрацией модулей. Переводы и transferFrom требуют KYC (как в GNDst-1), а также проверок паузы и заморозки. События: KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered, ModuleCall.

Итоговая схема и пояснения приведены выше; при изменении стандартов или добавлении новых контрактов документ следует обновлять.
