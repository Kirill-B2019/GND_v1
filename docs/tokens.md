# Токены GANYMED

## Обзор

GANYMED поддерживает различные стандарты токенов, включая ERC20, ERC721 и ERC1155. Каждый стандарт имеет свои особенности и области применения.

## Стандарты

### ERC20
Стандарт для взаимозаменяемых токенов (fungible tokens).

#### Основные функции
```solidity
interface IERC20 {
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address recipient, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function approve(address spender, uint256 amount) external returns (bool);
    function transferFrom(address sender, address recipient, uint256 amount) external returns (bool);
    
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
}
```

#### Пример использования
```solidity
contract MyToken is ERC20 {
    constructor() ERC20("My Token", "MTK") {
        _mint(msg.sender, 1000000 * 10 ** decimals());
    }
}
```

### ERC721
Стандарт для невзаимозаменяемых токенов (non-fungible tokens).

#### Основные функции
```solidity
interface IERC721 {
    function balanceOf(address owner) external view returns (uint256);
    function ownerOf(uint256 tokenId) external view returns (address);
    function safeTransferFrom(address from, address to, uint256 tokenId) external;
    function transferFrom(address from, address to, uint256 tokenId) external;
    function approve(address to, uint256 tokenId) external;
    function getApproved(uint256 tokenId) external view returns (address);
    function setApprovalForAll(address operator, bool approved) external;
    function isApprovedForAll(address owner, address operator) external view returns (bool);
    
    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);
    event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId);
    event ApprovalForAll(address indexed owner, address indexed operator, bool approved);
}
```

#### Пример использования
```solidity
contract MyNFT is ERC721 {
    using Counters for Counters.Counter;
    Counters.Counter private _tokenIds;
    
    constructor() ERC721("My NFT", "MNFT") {}
    
    function mint(address to) public returns (uint256) {
        _tokenIds.increment();
        uint256 newTokenId = _tokenIds.current();
        _mint(to, newTokenId);
        return newTokenId;
    }
}
```

### ERC1155
Стандарт для мультитокенов, поддерживающий как взаимозаменяемые, так и невзаимозаменяемые токены.

#### Основные функции
```solidity
interface IERC1155 {
    function balanceOf(address account, uint256 id) external view returns (uint256);
    function balanceOfBatch(address[] calldata accounts, uint256[] calldata ids) external view returns (uint256[] memory);
    function setApprovalForAll(address operator, bool approved) external;
    function isApprovedForAll(address account, address operator) external view returns (bool);
    function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes calldata data) external;
    function safeBatchTransferFrom(address from, address to, uint256[] calldata ids, uint256[] calldata amounts, bytes calldata data) external;
    
    event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value);
    event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values);
    event ApprovalForAll(address indexed account, address indexed operator, bool approved);
    event URI(string value, uint256 indexed id);
}
```

#### Пример использования
```solidity
contract MyMultiToken is ERC1155 {
    constructor() ERC1155("") {}
    
    function mint(address to, uint256 id, uint256 amount, bytes memory data) public {
        _mint(to, id, amount, data);
    }
    
    function mintBatch(address to, uint256[] memory ids, uint256[] memory amounts, bytes memory data) public {
        _mintBatch(to, ids, amounts, data);
    }
}
```

## Расширения

### ERC20
- ERC20Burnable
- ERC20Capped
- ERC20Pausable
- ERC20Snapshot
- ERC20Votes

### ERC721
- ERC721Enumerable
- ERC721URIStorage
- ERC721Burnable
- ERC721Pausable
- ERC721Votes

### ERC1155
- ERC1155Burnable
- ERC1155Pausable
- ERC1155Supply

## Безопасность

### Рекомендации
1. Используйте проверенные библиотеки
2. Проводите аудит кода
3. Тестируйте контракты
4. Следите за газами
5. Используйте безопасные паттерны

### Паттерны
1. Checks-Effects-Interactions
2. Pull over Push
3. Emergency Stop
4. Rate Limiting
5. Access Control

## Интеграция

### Web3.js
```javascript
const Web3 = require('web3');
const web3 = new Web3('https://api.gnd-net.com:8181');

const contract = new web3.eth.Contract(ABI, address);

async function getBalance(address) {
    const balance = await contract.methods.balanceOf(address).call();
    return balance;
}

async function transfer(to, amount) {
    const accounts = await web3.eth.getAccounts();
    await contract.methods.transfer(to, amount).send({from: accounts[0]});
}
```

### Ethers.js
```javascript
const { ethers } = require('ethers');
const provider = new ethers.providers.JsonRpcProvider('https://api.gnd-net.com:8181');

const contract = new ethers.Contract(address, ABI, provider);

async function getBalance(address) {
    const balance = await contract.balanceOf(address);
    return balance;
}

async function transfer(signer, to, amount) {
    const contractWithSigner = contract.connect(signer);
    await contractWithSigner.transfer(to, amount);
}
```

## Мониторинг

### События
```javascript
contract.on("Transfer", (from, to, amount, event) => {
    console.log(`Transfer: ${from} -> ${to}: ${amount}`);
});
```

### Метрики
1. Количество транзакций
2. Использование газа
3. Активность контракта
4. Балансы токенов
5. События

## Обновление

### Прокси паттерн
```solidity
contract Proxy {
    address public implementation;
    
    function upgrade(address newImplementation) external {
        implementation = newImplementation;
    }
    
    fallback() external payable {
        address _impl = implementation;
        assembly {
            calldatacopy(0, 0, calldatasize())
            let result := delegatecall(gas(), _impl, 0, calldatasize(), 0, 0)
            returndatacopy(0, 0, returndatasize())
            switch result
            case 0 { revert(0, returndatasize()) }
            default { return(0, returndatasize()) }
        }
    }
}
```

### Миграция данных
1. Создание нового контракта
2. Копирование состояния
3. Обновление ссылок
4. Тестирование
5. Переключение

## Документация

### NatSpec
```solidity
/// @title Simple Token
/// @author John Doe
/// @notice This is a simple ERC20 token
/// @dev All function calls are currently implemented without side effects
contract SimpleToken {
    /// @notice Returns the balance of the specified address
    /// @param account The address to query the balance of
    /// @return The amount of tokens owned by the specified address
    function balanceOf(address account) external view returns (uint256) {
        return _balances[account];
    }
}
```

### README
1. Описание контракта
2. Установка
3. Использование
4. Тестирование
5. Деплой
6. Безопасность
7. Лицензия

## Лицензии

### MIT
```solidity
// SPDX-License-Identifier: MIT
```

### GPL-3.0
```solidity
// SPDX-License-Identifier: GPL-3.0
```

### Apache-2.0
```solidity
// SPDX-License-Identifier: Apache-2.0
```