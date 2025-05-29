// ui/wallet/src/components/TokenList.tsx

import React, { useState } from "react";

// Тип токена (может быть расширен)
export interface Token {
    address: string;
    name: string;
    symbol: string;
    decimals: number;
    balance: string; // строка для поддержки больших чисел
}

// Пропсы компонента: адрес кошелька и список токенов
interface TokenListProps {
    walletAddress: string;
    tokens: Token[];
}

const TokenList: React.FC<TokenListProps> = ({ walletAddress, tokens }) => {
    const [notification, setNotification] = useState<string>("");

    // Копирование адреса токена в буфер обмена
    const handleCopyAddress = (address: string) => {
        navigator.clipboard.writeText(address).then(() => {
            setNotification("Адрес токена скопирован!");
            setTimeout(() => setNotification(""), 1000);
        });
    };

    // Форматирование баланса с учётом decimals
    const formatBalance = (balance: string, decimals: number) => {
        if (!balance) return "0";
        const intPart = balance.slice(0, balance.length - decimals) || "0";
        const fracPart = balance.slice(-decimals).padStart(decimals, "0");
        return `${parseInt(intPart, 10)}.${fracPart.replace(/0+$/, "")}`;
    };

    return (
        <div className="token-list">
            {notification && (
                <div className="notification" style={{ color: "green", marginBottom: 8 }}>
                    {notification}
                </div>
            )}
            <h3>Токены на адресе {walletAddress}</h3>
            {tokens.length > 0 ? (
                tokens.map((token) => (
                    <div key={token.address} className="token-item" style={{ border: "1px solid #eee", marginBottom: 12, padding: 8 }}>
                        <h4>
                            {token.name} ({token.symbol})
                        </h4>
                        <p>
                            Баланс: <b>{formatBalance(token.balance, token.decimals)}</b>
                        </p>
                        <p>
                            Адрес токена:{" "}
                            <span style={{ fontFamily: "monospace" }}>{token.address}</span>
                            <button
                                style={{ marginLeft: 8 }}
                                onClick={() => handleCopyAddress(token.address)}
                            >
                                Копировать
                            </button>
                        </p>
                    </div>
                ))
            ) : (
                <p>Нет токенов на этом адресе.</p>
            )}
        </div>
    );
};

export default TokenList;
