// ui/wallet/src/components/Transactions.tsx

import React, { useEffect, useState } from "react";
import GanymedeApi, { Transaction } from "../api/ganymedeApi";

interface TransactionsProps {
    address: string;
}

const api = new GanymedeApi(process.env.REACT_APP_API_URL || "http://localhost:8080");

const formatDate = (timestamp?: number) => {
    if (!timestamp) return "";
    const date = new Date(timestamp * 1000);
    return date.toLocaleString();
};

const formatValue = (value: string | number) => {
    if (!value) return "0";
    return typeof value === "string" ? value : value.toString();
};

const Transactions: React.FC<TransactionsProps> = ({ address }) => {
    const [txs, setTxs] = useState<Transaction[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchTxs = async () => {
        setLoading(true);
        setError(null);
        try {
            const res = await api.getTransactions(address);
            setTxs(res);
        } catch {
            setError("Ошибка загрузки истории транзакций");
        }
        setLoading(false);
    };

    useEffect(() => {
        if (address) fetchTxs();
        // eslint-disable-next-line
    }, [address]);

    return (
        <div className="transactions-card">
            <h3>История транзакций</h3>
            {loading && <div>Загрузка...</div>}
            {error && <div style={{ color: "red" }}>{error}</div>}
            {!loading && !error && txs.length === 0 && (
                <div>Нет транзакций для этого адреса.</div>
            )}
            {!loading && !error && txs.length > 0 && (
                <table style={{ width: "100%", fontSize: 14, marginTop: 8 }}>
                    <thead>
                    <tr>
                        <th>Хеш</th>
                        <th>Тип</th>
                        <th>Отправитель</th>
                        <th>Получатель</th>
                        <th>Сумма</th>
                        <th>Nonce</th>
                        <th>Время</th>
                        <th>Статус</th>
                    </tr>
                    </thead>
                    <tbody>
                    {txs.map((tx) => (
                        <tr key={tx.hash}>
                            <td style={{ fontFamily: "monospace", maxWidth: 120, overflow: "hidden", textOverflow: "ellipsis" }}>
                                {tx.hash.slice(0, 10)}...
                            </td>
                            <td>{tx.type}</td>
                            <td style={{ fontFamily: "monospace" }}>
                                {tx.from.slice(0, 8)}...
                            </td>
                            <td style={{ fontFamily: "monospace" }}>
                                {tx.to.slice(0, 8)}...
                            </td>
                            <td>{formatValue(tx.value)}</td>
                            <td>{tx.nonce}</td>
                            <td>{formatDate(tx.timestamp)}</td>
                            <td>{tx.status || "—"}</td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            )}
            <button onClick={fetchTxs} disabled={loading} style={{ marginTop: 12 }}>
                Обновить
            </button>
        </div>
    );
};

export default Transactions;
