// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title OracleExample - Пример смарт-контракта оракула для блокчейна ГАНИМЕД
/// @notice Данные могут быть записаны только доверенными оракулами (мультисиг)

contract OracleExample {
    // Массив доверенных оракулов
    address[] public oracles;
    // Минимальное количество подтверждений для записи данных
    uint8 public minConfirmations;

    // Структура для хранения заявки на запись данных
    struct OracleRequest {
        string data;                  // Внешние данные (например, курс, событие)
        uint256 confirmations;        // Количество подтверждений
        mapping(address => bool) voted; // Кто уже проголосовал
        bool executed;                // Было ли исполнено
    }

    // Массив всех заявок
    OracleRequest[] public requests;

    // Событие при создании новой заявки
    event OracleRequestCreated(uint256 indexed requestId, string data, address indexed proposer);
    // Событие при подтверждении заявки оракулом
    event OracleRequestConfirmed(uint256 indexed requestId, address indexed oracle);
    // Событие при исполнении заявки
    event OracleRequestExecuted(uint256 indexed requestId, string data);

    // Модификатор: только для оракулов
    modifier onlyOracle() {
        require(isOracle(msg.sender), "Not an oracle");
        _;
    }

    /// @notice Конструктор контракта
    /// @param _oracles Массив адресов доверенных оракулов
    /// @param _minConfirmations Минимальное число подтверждений для записи
    constructor(address[] memory _oracles, uint8 _minConfirmations) {
        require(_oracles.length >= _minConfirmations, "Not enough oracles");
        oracles = _oracles;
        minConfirmations = _minConfirmations;
    }

    /// @notice Проверка, является ли адрес доверенным оракулом
    /// @param who Проверяемый адрес
    /// @return true если адрес - оракул
    function isOracle(address who) public view returns (bool) {
        for (uint i = 0; i < oracles.length; i++) {
            if (oracles[i] == who) return true;
        }
        return false;
    }

    /// @notice Предложить новые внешние данные (создать заявку)
    /// @param data Внешние данные (например, цена, событие)
    /// @return requestId Идентификатор заявки
    function proposeData(string memory data) public onlyOracle returns (uint256 requestId) {
        OracleRequest storage req = requests.push();
        req.data = data;
        req.confirmations = 0;
        req.executed = false;
        emit OracleRequestCreated(requests.length - 1, data, msg.sender);
        return requests.length - 1;
    }

    /// @notice Подтвердить заявку на запись данных
    /// @param requestId Идентификатор заявки
    function confirmData(uint256 requestId) public onlyOracle {
        require(requestId < requests.length, "Invalid requestId");
        OracleRequest storage req = requests[requestId];
        require(!req.executed, "Already executed");
        require(!req.voted[msg.sender], "Already voted");

        req.voted[msg.sender] = true;
        req.confirmations += 1;
        emit OracleRequestConfirmed(requestId, msg.sender);

        if (req.confirmations >= minConfirmations) {
            req.executed = true;
            emit OracleRequestExecuted(requestId, req.data);
        }
    }

    /// @notice Получить данные по заявке
    /// @param requestId Идентификатор заявки
    /// @return data Внешние данные
    /// @return confirmations Количество подтверждений
    /// @return executed Было ли исполнено
    function getRequest(uint256 requestId) public view returns (string memory data, uint256 confirmations, bool executed) {
        require(requestId < requests.length, "Invalid requestId");
        OracleRequest storage req = requests[requestId];
        return (req.data, req.confirmations, req.executed);
    }

    /// @notice Получить количество заявок
    function getRequestsCount() public view returns (uint256) {
        return requests.length;
    }
}
