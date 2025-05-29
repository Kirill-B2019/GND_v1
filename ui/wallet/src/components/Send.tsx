// ui/wallet/src/components/Send.tsx

import React, { useState } from "react";
import GanymedeApi, { SendTxRequest } from "../api/ganymedeApi";

// Для подписания транзакции приватным ключом используйте свою реализацию или библиотеку
// Здесь предполагается, что приватный ключ уже находится в памяти (например, после импорта/разблокировки кошелька)

interface SendProps {
    address: string; // Ваш адрес (отправитель)
    privateKey: string; // Приватный ключ в hex (используйте безопасно!)
}

const api = new GanymedeApi(process.env.REACT_APP_API_URL || "http://localhost:8080");

const Send: React.FC<SendProps> = ({ address, privateKey }) => {
    const [to, setTo] = useState("");
    const [amount, setAmount] = useState("");
    const [gasPrice, setGasPrice] = useState("1");
    const [gasLimit, setGasLimit] = useState("21000");
    const [nonce, setNonce] = useState("");
    const [status, setStatus] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    // Получить nonce (можно реализовать через API)
    const fetchNonce = async () => {
        try {
            const res = await api.api.get<{ nonce: number }>(`/api/wallet/${address}/nonce`);
            setNonce(res.data.nonce.toString());
        } catch {
            setNonce("");
        }
    };

    // Подпись транзакции (заглушка — реализуйте реальную подпись!)
    function signTx(tx: Omit<SendTxRequest, "signature">, privKey: string): string {
        // Здесь должна быть реализация ECDSA/secp256k1 подписи (например, через elliptic)
        // Для MVP можно возвращать "stub-signature"
        return "stub-signature";
    }

    const handleSend = async (e: React.FormEvent) => {
        e.preventDefault();
        setStatus(null);
        setLoading(true);

        try {
            const txData: Omit<SendTxRequest, "signature"> = {
                from: address,
                to: to,
                value: amount,
                gasPrice: gasPrice,
                gasLimit: gasLimit,
                nonce: Number(nonce),
                type: "transfer",
                // data: "" // если нужно
            };

            // Подписываем транзакцию приватным ключом
            const signature = signTx(txData, privateKey);

            const tx: SendTxRequest = { ...txData, signature };

            const res = await api.sendTransaction(tx);
            setStatus(`Транзакция отправлена! Хеш: ${res.hash}`);
            setTo("");
            setAmount("");
        } catch (err: any) {
            setStatus("Ошибка отправки транзакции");
        }
        setLoading(false);
    };

    return (
        <div className="send-card">
            <h3>Отправить токены</h3>
            <form onSubmit={handleSend}>
                <div>
                    <label>С вашего адреса:</label>
                    <div style={{ wordBreak: "break-all" }}>{address}</div>
                </div>
                <div>
                    <label>Адрес получателя:</label>
                    <input
                        type="text"
                        value={to}
                        onChange={e => setTo(e.target.value)}
                        placeholder="GND..."
                        required
                        style={{ width: "100%" }}
                    />
                </div>
                <div>
                    <label>Сумма (GND):</label>
                    <input
                        type="number"
                        value={amount}
                        onChange={e => setAmount(e.target.value)}
                        min="0"
                        required
                        style={{ width: "100%" }}
                    />
                </div>
                <div>
                    <label>Gas Price:</label>
                    <input
                        type="number"
                        value={gasPrice}
                        onChange={e => setGasPrice(e.target.value)}
                        min="1"
                        required
                        style={{ width: "100%" }}
                    />
                </div>
                <div>
                    <label>Gas Limit:</label>
                    <input
                        type="number"
                        value={gasLimit}
                        onChange={e => setGasLimit(e.target.value)}
                        min="21000"
                        required
                        style={{ width: "100%" }}
                    />
                </div>
                <div>
                    <label>Nonce:</label>
                    <input
                        type="number"
                        value={nonce}
                        onChange={e => setNonce(e.target.value)}
                        required
                        style={{ width: "100%" }}
                    />
                    <button type="button" onClick={fetchNonce} style={{ marginLeft: 8 }}>
                        Получить nonce
                    </button>
                </div>
                <button type="submit" disabled={loading || !to || !amount || !nonce}>
                    {loading ? "Отправка..." : "Отправить"}
                </button>
            </form>
            {status && <div style={{ marginTop: 12 }}>{status}</div>}
        </div>
    );
};

export default Send;
