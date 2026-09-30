import { createContext } from "react";

export interface CurrencyContextValue {
  currency: string;
}

export const CurrencyContext = createContext<CurrencyContextValue | null>(null);
