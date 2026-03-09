# Формальная спецификация инвариантов (deploy_order)

Канонический стандарт: **IGNDst1** (GND-st1); **IGNDRWA** расширяет IGNDst1 для RWA. Каталог `deploy_order` содержит копии контрактов для автономной компиляции; при изменениях синхронизировать с `tokens/standards/gndst1` и `tokens/standards/gndrwa`.

---

## NativeTokensController (01)

- **INVARIANT:** `owner == constructor(owner_)` и `owner != address(0)`; `owner` immutable, не меняется.
- **INVARIANT:** `setGndToken` вызывается не более одного раза: до вызова `gndToken == address(0)`, после — запись запрещена (TokenAlreadySet).
- **INVARIANT:** Аналогично для `setGaniToken` и `ganiToken`.
- **INVARIANT:** Единственные функции, меняющие состояние: `setGndToken`, `setGaniToken`, `mintGANI`, `setKycGnd`, `setKycGani`; все защищены `onlyOwner`.
- **REENTRANCY:** Нет пользовательских callback'ов; только низкоуровневые `call` в известные адреса токенов (gndToken, ganiToken). После вызова — только emit; опасного reentrancy нет.

---

## GNDst1Token / GNDToken (gndst1Base, 02)

- **INVARIANT:** Всегда `0 < _totalSupply <= TOTAL_SUPPLY`. `_totalSupply` увеличивается только в конструкторе через `_mint(controller, initialSupply)`; `mint()` всегда revert (MintingDisabled).
- **INVARIANT:** Сумма `_balances[*]` равна `_totalSupply`; иного способа изменить supply нет.
- **INVARIANT:** `controller` immutable; `onlyController` требует `msg.sender == controller` и `extcodesize(msg.sender) > 0`.
- **INVARIANT:** `transfer`, `transferFrom`, `crossChainTransfer` защищены `onlyKyc` (отправитель должен быть в `_kycPassed`); `setKycStatus` — только `onlyController`.
- **INVARIANT:** В `claimDividends` выплата идёт с баланса `controller`; контроллер должен иметь достаточный баланс (обеспечивается при деплое и эмиссии).

---

## GANIToken (03)

- **INVARIANT:** Всегда `0 <= _totalSupply <= TOTAL_SUPPLY`. В конструкторе `_totalSupply = FIRST_EMISSION`; далее рост только через `mint()` с проверкой `_totalSupply + amount <= TOTAL_SUPPLY`.
- **INVARIANT:** `mint()` доступен только через `onlyController`; `onlyController` требует `msg.sender == controller` и контракт-контроллер.
- **INVARIANT:** Сумма балансов равна `_totalSupply`; толькоKyc / onlyController — как у GND.
- **INVARIANT:** При `bridge == address(0)` вызов `crossChainTransfer` запрещён (require); иначе токены уходили бы на нулевой адрес (сжигание).

---

## GNDRWAToken (04, RWA)

- **INVARIANT:** При `_maxSupply > 0` всегда `_totalSupply <= _maxSupply`. В `mint()` проверка `_totalSupply + amount <= _maxSupply`; в `burn()` уменьшается `_totalSupply`.
- **INVARIANT:** `transfer`, `transferFrom`, `crossChainTransfer` требуют `onlyKyc` и `_requireTransferAllowed` (нет паузы переводов, нет заморозки отправителя и получателя).
- **INVARIANT:** Функции mint, burn, setTransfersPaused, setMintPaused, setBurnPaused, setFrozen, setKycStatus, snapshot, setSnapshotBalance, setDividendsPerShare, registerModule доступны только через `onlyController`.
- **INVARIANT:** В `claimDividends` выплата с баланса контроллера; ответственность за наличие баланса у контроллера — off-chain / деплой.
- **REENTRANCY:** Нет вызовов во внешние ненадёжные контракты; `moduleCall` — заглушка (делегирование в модуль не реализовано).
