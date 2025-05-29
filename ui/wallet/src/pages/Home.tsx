// ui/wallet/src/pages/Home.tsx

import React, { useState } from "react";
import WalletCreate from "../components/WalletCreate";
import WalletImport from "../components/WalletImport";
import Balance from "../components/Balance";
import Send from "../components/Send";
import Receive from "../components/Receive";
import Transactions from "../components/Transactions";
import TokenList from "../components/TokenList";
import Settings from "../components/Settings";
import WalletExport from "../components/WalletExport";
import useWallet, { WalletData } from "../hooks/useWallet";

/**
 * Главная страница кошелька ГАНИМЕД.
 * Позволяет создать или импортировать кошелек, а после входа — управлять активами, отправлять и получать токены, просматривать историю, экспортировать приватный ключ и настраивать параметры.
 */

const DEFAULT_API_URL = "http://localhost:8080";

const Home: React.FC = () => {
    const {
        wallet,
        createWallet,
        importWallet,
        removeWallet,
        isLoaded,
    } = useWallet();

    const [apiUrl, setApiUrl] = useState(DEFAULT_API_URL);
    const [tokens, setTokens] = useState<any[]>([]); // Для примера, реализуйте загрузку токенов через API/hook
    const [encrypt, setEncrypt] = useState(false);

    // Обработчики создания/импорта кошелька
    const handleCreate = (data: WalletData) => {
        // Можно добавить сохранение в localStorage или context
    };
    const handleImport = (data: WalletData) => {
        // Можно добавить сохранение в localStorage или context
    };
    const handleLogout = () => {
        removeWallet();
    };

    // Пример загрузки токенов по адресу (реализуйте через API/hook)
    // useEffect(() => {
    //   if (wallet?.address) {
    //     api.getTokens(wallet.address).then(setTokens);
    //   }
    // }, [wallet, apiUrl]);

    return (
        <div style={{ maxWidth: 600, margin: "0 auto", padding: 24 }}>
            <h2>Кошелек ГАНИМЕД</h2>
            {!isLoaded && (
                <div style={{ marginBottom: 24 }}>
                    <WalletCreate onCreate={handleCreate} />
                    <div style={{ margin: "16px 0", textAlign: "center" }}>или</div>
                    <WalletImport onImport={handleImport} />
                </div>
            )}
            {isLoaded && wallet && (
                <>
                    <div style={{ marginBottom: 24 }}>
                        <WalletExport
                            privateKey={wallet.privateKey}
                            publicKey={wallet.publicKey}
                            address={wallet.address}
                        />
                    </div>
                    <Balance address={wallet.address} />
                    <TokenList walletAddress={wallet.address} tokens={tokens} />
                    <Send address={wallet.address} privateKey={wallet.privateKey} />
                    <Receive address={wallet.address} />
                    <Transactions address={wallet.address} />
                    <Settings
                        apiUrl={apiUrl}
                        onApiUrlChange={setApiUrl}
                        onLogout={handleLogout}
                        encryptPrivateKey={encrypt}
                        onEncryptToggle={setEncrypt}
                    />
                </>
            )}
        </div>
    );
};

export default Home;
