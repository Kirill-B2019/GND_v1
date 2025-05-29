// ui/wallet/src/hooks/useWallet.ts

import { useState, useCallback } from "react";

/**
 * Типы для кошелька ГАНИМЕД
 */
export interface WalletData {
    address: string;
    privateKey: string;
    publicKey: string;
}

/**
 * Хук для управления состоянием кошелька пользователя.
 * Позволяет создавать, импортировать, экспортировать и удалять кошелек.
 * Приватный ключ хранится только в памяти (НЕ отправляйте его на сервер!).
 */
export function useWallet(initialWallet?: WalletData) {
    const [wallet, setWallet] = useState<WalletData | null>(initialWallet || null);

    // Создание нового кошелька (вызывайте вашу функцию генерации)
    const createWallet = useCallback((generator: () => WalletData) => {
        const newWallet = generator();
        setWallet(newWallet);
        return newWallet;
    }, []);

    // Импорт кошелька по приватному ключу (вызывайте вашу функцию импорта)
    const importWallet = useCallback((importer: (privKey: string) => WalletData, privKey: string) => {
        const imported = importer(privKey);
        setWallet(imported);
        return imported;
    }, []);

    // Удаление кошелька из памяти (выход)
    const removeWallet = useCallback(() => {
        setWallet(null);
    }, []);

    // Экспорт приватного ключа
    const exportPrivateKey = useCallback(() => {
        return wallet?.privateKey || "";
    }, [wallet]);

    // Экспорт публичного ключа
    const exportPublicKey = useCallback(() => {
        return wallet?.publicKey || "";
    }, [wallet]);

    // Экспорт адреса
    const exportAddress = useCallback(() => {
        return wallet?.address || "";
    }, [wallet]);

    return {
        wallet,
        setWallet,
        createWallet,
        importWallet,
        removeWallet,
        exportPrivateKey,
        exportPublicKey,
        exportAddress,
        isLoaded: !!wallet,
    };
}

export default useWallet;
