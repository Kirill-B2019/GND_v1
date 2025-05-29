// api/ganymedeApi.ts

import axios, { AxiosInstance } from "axios";

// Типы для баланса, транзакций и токенов
export interface BalanceResponse {
    address: string;
    balance: string; // GND
}

export interface Transaction {
    hash: string;
    from: string;
    to: string;
    value: string;
    gasPrice: string;
    gasLimit: string;
    nonce: number;
    data?: string;
    type: string;
    signature?: string;
    timestamp?: number;
    status?: string;
}

export interface SendTxRequest {
    from: string;
    to: string;
    value: string;
    gasPrice: string;
    gasLimit: string;
    nonce: number;
    data?: string;
    type: string;
    signature: string;
}

export interface Token {
    address: string;
    name: string;
    symbol: string;
    decimals: number;
    balance: string;
}

export class GanymedeApi {
    api: AxiosInstance;

    constructor(baseURL: string = "http://localhost:8080") {
        this.api = axios.create({
            baseURL,
            timeout: 10_000,
        });
    }

    // Получить баланс по адресу
    async getBalance(address: string): Promise<BalanceResponse> {
        const res = await this.api.get<BalanceResponse>(`/api/wallet/${address}/balance`);
        return res.data;
    }

    // Получить историю транзакций по адресу
    async getTransactions(address: string): Promise<Transaction[]> {
        const res = await this.api.get<Transaction[]>(`/api/wallet/${address}/transactions`);
        return res.data;
    }

    // Отправить транзакцию (подписанную на клиенте)
    async sendTransaction(tx: SendTxRequest): Promise<{ hash: string }> {
        const res = await this.api.post<{ hash: string }>(`/api/tx/send`, tx);
        return res.data;
    }

    // Получить информацию о транзакции по хешу
    async getTransaction(hash: string): Promise<Transaction> {
        const res = await this.api.get<Transaction>(`/api/tx/${hash}`);
        return res.data;
    }

    // Получить список токенов (GND, ERC-20, TRC-20)
    async getTokens(address: string): Promise<Token[]> {
        const res = await this.api.get<Token[]>(`/api/wallet/${address}/tokens`);
        return res.data;
    }

    // Получить информацию о токене по адресу контракта
    async getTokenInfo(tokenAddress: string): Promise<Token> {
        const res = await this.api.get<Token>(`/api/token/${tokenAddress}`);
        return res.data;
    }

    // Новый публичный метод для получения nonce
    async getNonce(address: string): Promise<number> {
        const res = await this.api.get<{ nonce: number }>(`/api/wallet/${address}/nonce`);
        return res.data.nonce;
    }
}

export default GanymedeApi;
