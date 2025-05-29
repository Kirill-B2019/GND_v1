// ui/wallet/src/components/WalletImport.tsx

import React, { useState } from "react";
import { importWalletFromPrivateKey } from "../utils/crypto"; // Ваша функция импорта (пример ниже)
import { ValidateAddress } from "../utils/address";

interface WalletData {
    address: string;
    privateKey: string;
    publicKey: string;
}

const WalletImport: React.FC<{ onImport: (wallet: WalletData) => void }> = ({ onImport }) => {
    const [privateKey, setPrivateKey] = useState("");
    const [wallet, setWallet] = useState<WalletData | null>(null);
    const [error, setError] = useState<string | null>(null);

    const handleImport = () => {
        try {
            const importedWallet = importWalletFromPrivateKey(privateKey.trim());
            if (!ValidateAddress(importedWallet.address)) {
                setError("Ошибка: адрес некорректен.");
                setWallet(null);
                return;
            }
            setWallet(importedWallet);
            setError(null);
            onImport(importedWallet);
        } catch (e) {
            setError("Не удалось импортировать кошелек. Проверьте приватный ключ.");
            setWallet(null);
        }
    };

    return (
        <div className="wallet-import-card">
            <h3>Импортировать кошелек</h3>
            <div>
                <label>Приватный ключ:</label>
                <input
                    type="text"
                    value={privateKey}
                    onChange={e => setPrivateKey(e.target.value)}
                    placeholder="Введите приватный ключ (hex)"
                    style={{ width: "100%" }}
                />
            </div>
            <button onClick={handleImport} style={{ marginTop: 8 }}>
                Импортировать
            </button>
            {error && <div style={{ color: "red", marginTop: 8 }}>{error}</div>}
            {wallet && (
                <div style={{ marginTop: 16 }}>
                    <div>
                        <b>Адрес:</b>
                        <div style={{ wordBreak: "break-all" }}>{wallet.address}</div>
                    </div>
                    <div style={{ marginTop: 12 }}>
                        <b>Публичный ключ:</b>
                        <div style={{ wordBreak: "break-all", fontSize: 12 }}>{wallet.publicKey}</div>
                    </div>
                    <div style={{ marginTop: 12 }}>
                        <b>Приватный ключ:</b>
                        <div style={{ wordBreak: "break-all", fontSize: 12, color: "crimson" }}>{wallet.privateKey}</div>
                        <div style={{ color: "red", fontSize: 12, marginTop: 4 }}>
                            ВНИМАНИЕ: Никогда не делитесь приватным ключом! Храните его в надёжном месте.
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default WalletImport;
