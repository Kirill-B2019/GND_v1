// ui/wallet/src/components/Balance.tsx

import React, { useEffect, useState } from "react";
import GanymedeApi, { BalanceResponse } from "../api/ganymedeApi";

// Пример: адрес пользователя может быть получен из props, контекста или хука useWallet
interface BalanceProps {
    address: string;
}

const api = new GanymedeApi(process.env.REACT_APP_API_URL || "http://localhost:8080");

const Balance: React.FC<BalanceProps> = ({ address }) => {
    const [balance, setBalance] = useState<string | null>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    const fetchBalance = async () => {
        setLoading(true);
        setError(null);
        try {
            const res: BalanceResponse = await api.getBalance(address);
            setBalance(res.balance);
        } catch (e) {
            setError("Ошибка получения баланса");
        }
        setLoading(false);
    };

    useEffect(() => {
        if (address) {
            fetchBalance();
        }
        // Можно добавить автообновление баланса через setInterval
        // return () => clearInterval(timer);
        // eslint-disable-next-line
    }, [address]);

    return (
        <div className="balance-card">
            <h3>Баланс</h3>
            <div>
                <b>Адрес:</b> {address}
            </div>
            {loading && <div>Загрузка...</div>}
            {error && <div style={{ color: "red" }}>{error}</div>}
            {!loading && !error && (
                <div>
                    <b>GND:</b> {balance !== null ? balance : "—"}
                </div>
            )}
            <button onClick={fetchBalance} disabled={loading}>
                Обновить баланс
            </button>
        </div>
    );
};

export default Balance;
