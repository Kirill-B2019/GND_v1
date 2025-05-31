// ui/wallet/src/components/WalletCreate.tsx

import React, { useState } from "react";
import { generateWallet } from "../utils/crypto"; // Ваша функция генерации кошелька (пример ниже)
import { ValidateAddress } from "../utils/address"; // Ваша функция валидации адреса (пример ниже)

interface WalletData {
    address: string;
    privateKey: string;
    publicKey: string;
}

const WalletCreate: React.FC<{ onCreate: (wallet: WalletData) => void }> = ({ onCreate }) => {
    const [wallet, setWallet] = useState<WalletData | null>(null);
    const [error, setError] = useState<string | null>(null);

    const handleCreate = () => {
        try {
            const newWallet = generateWallet();
            if (!ValidateAddress(newWallet.address)) {
                setError("Ошибка: адрес некорректен.");
                return;
            }
            setWallet(newWallet);
            setError(null);
            onCreate(newWallet);
        } catch (e) {
            setError("Не удалось создать кошелек");
            console.error(e);
        }
    };

    const handleCopy = async (text: string, label: string) => {
        try {
            await navigator.clipboard.writeText(text);
            alert(`${label} скопирован в буфер обмена!`);
        } catch {
            alert(`Не удалось скопировать ${label}.`);
        }
    };

    return (
        <div className="wallet-create-card">
            <h3>Создать новый кошелек</h3>
            <button onClick={handleCreate}>Создать кошелек</button>
            {error && <div style={{ color: "red", marginTop: 8 }}>{error}</div>}
            {wallet && (
                <div style={{ marginTop: 16 }}>
                    <div>
                        <b>Адрес:</b>
                        <div style={{ wordBreak: "break-all" }}>{wallet.address}</div>
                        <button onClick={() => handleCopy(wallet.address, "Адрес")}>Скопировать адрес</button>
                    </div>
                    <div style={{ marginTop: 12 }}>
                        <b>Публичный ключ:</b>
                        <div style={{ wordBreak: "break-all", fontSize: 12 }}>{wallet.publicKey}</div>
                        <button onClick={() => handleCopy(wallet.publicKey, "Публичный ключ")}>Скопировать публичный ключ</button>
                    </div>
                    <div style={{ marginTop: 12 }}>
                        <b>Приватный ключ:</b>
                        <div style={{ wordBreak: "break-all", fontSize: 12, color: "crimson" }}>{wallet.privateKey}</div>
                        <button onClick={() => handleCopy(wallet.privateKey, "Приватный ключ")}>Скопировать приватный ключ</button>
                        <div style={{ color: "red", fontSize: 12, marginTop: 4 }}>
                            ВНИМАНИЕ: Никогда не делитесь приватным ключом! Храните его в надёжном месте.
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default WalletCreate;
