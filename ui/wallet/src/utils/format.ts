// ui/wallet/src/utils/format.ts

/**
 * Утилиты для форматирования баланса, адресов и чисел в кошельке ГАНИМЕД
 */

/**
 * Форматирует баланс токена с учётом decimals (например, 18 для GND, 6/8 для других токенов)
 * @param balance - строка или число (целое, без разделителя)
 * @param decimals - количество знаков после запятой
 * @param precision - сколько знаков показать после запятой (по умолчанию 6)
 * @returns строка вида "0.123456" или "123.45"
 */
export function formatTokenBalance(
    balance: string | number,
    decimals: number,
    precision = 6
): string {
    if (!balance || isNaN(Number(balance))) return "0";
    const raw = BigInt(balance.toString());
    const divisor = BigInt(10) ** BigInt(decimals);

    const intPart = raw / divisor;
    const fracPart = raw % divisor;

    // Формируем дробную часть с ведущими нулями
    let fracStr = fracPart.toString().padStart(decimals, "0").slice(0, precision);
    // Обрезаем лишние нули справа
    fracStr = fracStr.replace(/0+$/, "");
    return fracStr.length > 0 ? `${intPart.toString()}.${fracStr}` : intPart.toString();
}

/**
 * Форматирует адрес для краткого отображения (например, GND1ab...cdef)
 * @param address - исходный адрес
 * @param prefixLen - сколько символов оставить в начале (по умолчанию 6)
 * @param suffixLen - сколько символов оставить в конце (по умолчанию 4)
 * @returns строка вида "GND1ab...cdef"
 */
export function formatAddress(
    address: string,
    prefixLen = 6,
    suffixLen = 4
): string {
    if (!address || address.length <= prefixLen + suffixLen + 3) return address;
    return `${address.slice(0, prefixLen)}...${address.slice(-suffixLen)}`;
}

/**
 * Форматирует число с разделителем тысяч
 * @param value - число или строка
 * @returns строка с разделителями
 */
export function formatNumber(value: number | string): string {
    const num = Number(value);
    if (isNaN(num)) return "0";
    return num.toLocaleString("ru-RU");
}
