import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { SERVER_URL } from "./DemoContext";
import { CurrencyContext } from "./CurrencyContext";

export const CurrencyProvider = ({ children }: { children: ReactNode }) => {
  const [currency, setCurrency] = useState("USD");

  useEffect(() => {
    let mounted = true;
    const loadCurrency = async () => {
      try {
        const response = await fetch(`${SERVER_URL}/health`);
        if (!response.ok) return;
        const body: { currency?: string } = await response.json();
        if (mounted && body.currency) {
          setCurrency(body.currency);
        }
      } catch {
        // Keep USD fallback when server configuration cannot be loaded.
      }
    };

    loadCurrency();
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <CurrencyContext.Provider value={{ currency }}>
      {children}
    </CurrencyContext.Provider>
  );
};
