// ui/wallet/src/pages/Settings.tsx

import React, { useState } from "react";
import Settings from "../components/Settings";

/**
 * Страница настроек кошелька ГАНИМЕД.
 * Позволяет изменить URL API-ноды, включить/отключить шифрование приватного ключа,
 * выйти из кошелька и добавить другие пользовательские настройки.
 */

const DEFAULT_API_URL = "http://localhost:8080";

const SettingsPage: React.FC = () => {
    const [apiUrl, setApiUrl] = useState<string>(DEFAULT_API_URL);
    const [encrypt, setEncrypt] = useState<boolean>(false);

    // Пример обработчика выхода (можно интегрировать с useWallet)
    const handleLogout = () => {
        // Здесь реализуйте логику выхода из кошелька (очистка состояния, редирект и т.д.)
        alert("Выход из кошелька выполнен.");
    };

    return (
        <div style={{ maxWidth: 520, margin: "0 auto", padding: 24 }}>
            <h2>Настройки</h2>
            <Settings
                apiUrl={apiUrl}
                onApiUrlChange={setApiUrl}
                encryptPrivateKey={encrypt}
                onEncryptToggle={setEncrypt}
                onLogout={handleLogout}
            />
            {/* Здесь можно добавить дополнительные настройки, например:
          - выбор языка интерфейса
          - переключение темы (светлая/тёмная)
          - управление адресной книгой
          - автообновление баланса и т.д.
      */}
        </div>
    );
};

export default SettingsPage;
