# ТЗ агенту: контракты *.sol по варианту А (GND-st1 с самого начала)

## 1. Цель

Основной стандарт — **GND-st1**; токены **GND** и **GANI** реализованы как контракты этого стандарта. Контроллер — единая точка управления (mint GANI, setKyc для GND и GANI).

## 2. Исходные материалы

- Интерфейс: `tokens/standards/gndst1/IGNDst1.sol`
- Базовая реализация: `tokens/standards/gndst1/gndst1Base.sol` (GNDst1Token)
- Контроллер: `tokens/standards/deploy_order/01_NativeTokensController.sol`
- GND: `tokens/standards/deploy_order/02_GNDToken.sol` (наследует GNDst1Token)
- GANI: `tokens/standards/deploy_order/03_GANIToken.sol` (IGNDst1 + mint от контроллера)

## 3. Реализованные файлы (deploy_order)

### 3.1. 01_NativeTokensController.sol

- Конструктор без аргументов; `owner`, `gndToken`, `ganiToken`.
- `setGndToken(address)` — только owner.
- `setGaniToken(address)` — только owner.
- `mintGANI(to, amount)` — вызов `mint(to, amount)` на ganiToken (только owner).
- `setKycGnd(user, status)` — вызов `setKycStatus(user, status)` на gndToken (только owner).
- `setKycGani(user, status)` — вызов `setKycStatus(user, status)` на ganiToken (только owner).

### 3.2. 02_GNDToken.sol

- Наследует `GNDst1Token` из `../gndst1/gndst1Base.sol`.
- Конструктор: `(initialSupply, bridgeAddress, controllerContract)` — передаётся в базовый конструктор.
- Минтинг отключён (в базе); эмиссия только в конструкторе.

### 3.3. 03_GANIToken.sol

- Реализует `IGNDst1`; имя/символ/decimals: Ganimed Governance, GANI, 6.
- `TOTAL_SUPPLY = 100_000_000 * 10**6`.
- Конструктор: `(controllerContract)` — минтит полное предложение на controller.
- `mint(to, amount)` — только controller; проверка `_totalSupply + amount <= TOTAL_SUPPLY`.

## 4. Общие требования по коду

- Solidity ^0.8.16.
- Контроллер проверять как контракт (extcodesize > 0).
- NatSpec (@title, @notice, @param) для контрактов и публичных функций.
- Лицензия SPDX (MIT).

## 5. Критерии приёмки

- Деплой в порядке 01 → 02 → 03 успешен.
- Контроллер: setGndToken, setGaniToken, mintGANI, setKycGnd, setKycGani работают от owner.
- GND: transfer/transferFrom только при KYC; snapshot/дивиденды — только controller.
- GANI: mint только с контроллера; totalSupply не превышает 100M*10^6; KYC и дивиденды как в GND-st1.
