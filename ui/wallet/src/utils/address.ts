// ui/wallet/src/utils/address.ts

import base58 from "base-58";
import sha256 from "crypto-js/sha256";

/**
 * Проверяет корректность адреса сети ГАНИМЕД (GND/GN + base58 + checksum)
 * - Префикс: "GND" (3 байта) или "GN" (2 байта), только верхний регистр
 * - Длина: 27 (GND) или 26 (GN) байт после декодирования
 * - Контрольная сумма: первые 4 байта двойного SHA256 от payload (префикс + pubKeyHash)
 */
export function ValidateAddress(address: string): boolean {
    try {
        const decoded = Buffer.from(base58.decode(address));
        if (decoded.length === 27 && decoded.slice(0, 3).toString() === "GND") {
            // ok
        } else if (decoded.length === 26 && decoded.slice(0, 2).toString() === "GN") {
            // ok
        } else {
            return false;
        }
        // Проверяем контрольную сумму
        const payload = decoded.slice(0, decoded.length - 4);
        const checksum = decoded.slice(decoded.length - 4);
        const calcChecksum = Buffer.from(
            sha256(sha256(payload)).toString(),
            "hex"
        ).slice(0, 4);
        return checksum.equals(calcChecksum);
    } catch {
        return false;
    }
}

/**
 * Добавляет префикс "GND" к адресу без префикса (для совместимости)
 */
export function AddPrefix(address: string): string {
    if (address.startsWith("GND") || address.startsWith("GN")) {
        return address;
    }
    return "GND" + address;
}

/**
 * Удаляет префикс "GND" или "GN" из адреса (для внутреннего использования)
 */
export function RemovePrefix(address: string): string {
    if (address.startsWith("GND")) {
        return address.slice(3);
    }
    if (address.startsWith("GN")) {
        return address.slice(2);
    }
    return address;
}
