import { createContext } from "react";

export interface ImportContextValue {
  isImporting: boolean;
  setImporting: (v: boolean) => void;
}

export const ImportContext = createContext<ImportContextValue>({
  isImporting: false,
  setImporting: () => {},
});
