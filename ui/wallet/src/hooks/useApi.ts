// ui/wallet/src/hooks/useApi.ts

import React, { useState, useCallback } from "react";
import GanymedeApi from "../api/ganymedeApi";

/**
 * Универсальный React-хук для работы с REST API блокчейна ГАНИМЕД.
 * Позволяет делать запросы, получать данные, статус загрузки и ошибки.
 */

export interface ApiState<T> {
    data: T | null;
    error: string | null;
    loading: boolean;
    refetch: () => Promise<void>;
}

/**
 * useApi — универсальный хук для GET-запросов к API.
 * @param fetcher — функция, возвращающая Promise с данными (например, api.getBalance)
 * @param deps — зависимости для useCallback/useEffect (например, [address])
 */
export function useApi<T>(fetcher: () => Promise<T>, deps: any[] = []): ApiState<T> {
    const [data, setData] = useState<T | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState<boolean>(false);

    const fetchData = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const res = await fetcher();
            setData(res);
        } catch (e: any) {
            setError(e?.message || "Ошибка запроса к API");
            setData(null);
        }
        setLoading(false);
    }, deps); // eslint-disable-line react-hooks/exhaustive-deps

    // Автоматический вызов при изменении deps
    React.useEffect(() => {
        fetchData();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, deps);

    return { data, error, loading, refetch: fetchData };
}

/**
 * Пример использования:
 *
 * import { useApi } from "../hooks/useApi";
 * import GanymedeApi from "../api/ganymedeApi";
 *
 * const api = new GanymedeApi("http://localhost:8080");
 * const { data: balance, loading, error, refetch } = useApi(
 *   () => api.getBalance(address),
 *   [address]
 * );
 */

export default useApi;
