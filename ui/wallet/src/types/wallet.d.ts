// ui/wallet/src/types/wallet.d.ts

/**
 * Типы для кошелька сети ГАНИМЕД (TypeScript declaration file)
 * Используется для типизации данных кошелька и связанных сущностей.
 */

export interface WalletData {
    /** Адрес кошелька (строка, base58, с префиксом GND/GN) */
    address: string;
    /** Приватный ключ в hex-формате (secp256k1) */
    privateKey: string;
    /** Публичный ключ в hex-формате (uncompressed) */
    publicKey: string;
}

export interface Token {
    /** Адрес токена (контракта) */
    address: string;
    /** Название токена */
    name: string;
    /** Символ токена */
    symbol: string;
    /** Количество знаков после запятой */
    decimals: number;
    /** Баланс токена (строка для поддержки больших чисел) */
    balance: string;
}

export interface Transaction {
    /** Хеш транзакции */
    hash: string;
    /** Адрес отправителя */
    from: string;
    /** Адрес получателя */
    to: string;
    /** Сумма перевода (GND или токены) */
    value: string;
    /** Цена газа */
    gasPrice: string;
    /** Лимит газа */
    gasLimit: string;
    /** Нонc */
    nonce: number;
    /** Тип транзакции (например, "transfer") */
    type: string;
    /** Подпись отправителя (hex) */
    signature?: string;
    /** Время (timestamp, unix) */
    timestamp?: number;
    /** Статус (например, "success", "pending", "failed") */
    status?: string;
    /** Данные для вызова контракта (опционально) */
    data?: string;
}
