// ui/wallet/src/components/Receive.tsx

import React from "react";
import {QRCodeSVG} from "qrcode.react";

interface ReceiveProps {
    address: string;
}

const Receive: React.FC<ReceiveProps> = ({ address }) => {
    // Можно добавить копирование адреса в буфер обмена
    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(address);
            alert("Адрес скопирован в буфер обмена!");
        } catch {
            alert("Не удалось скопировать адрес.");
        }
    };

    return (
        <div className="receive-card">
            <h3>Получить токены</h3>
            <div>
                <b>Ваш адрес:</b>
                <div style={{ wordBreak: "break-all", margin: "8px 0" }}>{address}</div>
                <button onClick={handleCopy}>Скопировать адрес</button>
            </div>
            <div style={{ margin: "16px 0" }}>
                <QRCodeSVG  value={address} size={164} />
                <div style={{ fontSize: 12, color: "#666", marginTop: 8 }}>
                    Отсканируйте QR-код для получения токенов GND
                </div>
            </div>
        </div>
    );
};

export default Receive;
