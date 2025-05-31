// ui/wallet/src/utils/crypto.ts

import { ec as EC } from "elliptic";
import sha256 from "crypto-js/sha256";
import ripemd160 from "crypto-js/ripemd160";
import base58 from "base-58";

export function generateWallet() {
    const ec = new EC("secp256k1");
    const key = ec.genKeyPair();
    const pub = key.getPublic(false, "hex");
    const sha = sha256(Buffer.from(pub.slice(2), "hex")).toString();
    const ripemd = ripemd160(Buffer.from(sha, "hex")).toString();
    const prefix = Buffer.from("GND");
    const pubKeyHash = Buffer.from(ripemd, "hex");
    const payload = Buffer.concat([prefix, pubKeyHash]);
    const checksum = sha256(sha256(payload)).toString().slice(0, 8);
    const fullPayload = Buffer.concat([payload, Buffer.from(checksum, "hex")]);
    const address = base58.encode(fullPayload);

    return {
        address,
        privateKey: key.getPrivate("hex"),
        publicKey: pub,
    };
}


/**
 * Импорт кошелька по приватному ключу (hex)
 */
export function importWalletFromPrivateKey(privKeyHex: string) {
    const ec = new EC("secp256k1");
    const key = ec.keyFromPrivate(privKeyHex, "hex");
    const pub = key.getPublic(false, "hex"); // uncompressed hex

    // Хешируем публичный ключ: SHA256 -> RIPEMD160
    const sha = sha256(Buffer.from(pub.slice(2), "hex")).toString();
    const ripemd = ripemd160(Buffer.from(sha, "hex")).toString();

    // Префикс "GND" (можно добавить случайный выбор "GN" для совместимости)
    const prefix = Buffer.from("GND");
    const pubKeyHash = Buffer.from(ripemd, "hex");
    const payload = Buffer.concat([prefix, pubKeyHash]);

    // Контрольная сумма: первые 4 байта двойного SHA256
    const checksum = sha256(sha256(payload)).toString().slice(0, 8);
    const fullPayload = Buffer.concat([payload, Buffer.from(checksum, "hex")]);
    const address = base58.encode(fullPayload);

    return {
        address,
        privateKey: privKeyHex,
        publicKey: pub,
    };
}

/**
 * Подпись сообщения приватным ключом (ECDSA/secp256k1)
 * @param message - строка или Buffer
 * @param privKeyHex - приватный ключ в hex
 * @returns подпись в hex
 */
export function signMessage(message: string | Buffer, privKeyHex: string): string {
    const ec = new EC("secp256k1");
    const key = ec.keyFromPrivate(privKeyHex, "hex");
    const msgHash = sha256(typeof message === "string" ? Buffer.from(message) : message).toString();
    const signature = key.sign(msgHash, "hex");
    return signature.toDER("hex");
}

/**
 * Верификация подписи (ECDSA/secp256k1)
 * @param message - строка или Buffer
 * @param signatureHex - подпись в hex
 * @param pubKeyHex - публичный ключ в hex (uncompressed)
 * @returns true/false
 */
export function verifySignature(
    message: string | Buffer,
    signatureHex: string,
    pubKeyHex: string
): boolean {
    const ec = new EC("secp256k1");
    const key = ec.keyFromPublic(pubKeyHex, "hex");
    const msgHash = sha256(typeof message === "string" ? Buffer.from(message) : message).toString();
    return key.verify(msgHash, signatureHex);
}
