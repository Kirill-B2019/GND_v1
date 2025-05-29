// ui/wallet/src/components/WalletExport.tsx

import React, { useState } from "react";

interface WalletExportProps {
    privateKey: string;        // Приватный ключ в hex-формате
    publicKey?: string;        // Публичный ключ (опционально)
    address: string;           // Адрес кошелька
    onExport?: () => void;     // Колбэк при экспорте (опционально)
}

const WalletExport: React.FC<WalletExportProps> = ({
                                                       privateKey,
                                                       publicKey,
                                                       address,
                                                       onExport,
                                                   }) => {
    const [copied, setCopied] = useState<string | null>(null);

    const handleCopy = async (text: string, label: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setCopied(label);
            setTimeout(() => setCopied(null), 1200);
        } catch {
            alert(`Не удалось скопировать ${label}.`);
        }
    };

    const handleExport = () => {
        if (onExport) onExport();
        // Можно добавить скачивание файла с приватным ключом
        const blob = new Blob(
            [
                `GANYMEDE WALLET EXPORT\n\nADDRESS: ${address}\nPUBLIC KEY: ${publicKey || "-"}\nPRIVATE KEY: ${privateKey}\n\nВНИМАНИЕ: Никогда не делитесь приватным ключом! Храните его в надёжном месте.`,
            ],
            { type: "text/plain" }
        );
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `ganymede_wallet_${address}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    return (
        <div className="wallet-export-card">
            <h3>Экспорт кошелька</h3>
            <div>
                <b>Адрес:</b>
                <div style={{ wordBreak: "break-all" }}>{address}</div>
                <button onClick={() => handleCopy(address, "Адрес")}>
                    Скопировать адрес
                </button>
            </div>
            {publicKey && (
                <div style={{ marginTop: 12 }}>
                    <b>Публичный ключ:</b>
                    <div style={{ wordBreak: "break-all", fontSize: 12 }}>{publicKey}</div>
                    <button onClick={() => handleCopy(publicKey, "Публичный ключ")}>
                        Скопировать публичный ключ
                    </button>
                </div>
            )}
            <div style={{ marginTop: 12 }}>
                <b>Приватный ключ:</b>
                <div style={{ wordBreak: "break-all", fontSize: 12, color: "crimson" }}>
                    {privateKey}
                </div>
                <button onClick={() => handleCopy(privateKey, "Приватный ключ")}>
                    Скопировать приватный ключ
                </button>
                <div style={{ color: "red", fontSize: 12, marginTop: 4 }}>
                    ВНИМАНИЕ: Никогда не делитесь приватным ключом! Храните его в надёжном месте.
                </div>
            </div>
            <div style={{ marginTop: 18 }}>
                <button onClick={handleExport}>Скачать файл экспорта</button>
            </div>
            {copied && (
                <div style={{ color: "green", marginTop: 8 }}>
                    {copied} скопирован в буфер обмена!
                </div>
            )}
        </div>
    );
};

export default WalletExport;
