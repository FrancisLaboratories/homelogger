import { useState } from "react";
import type { ReactNode } from "react";
import { ImportContext } from "./ImportContext";

export const ImportProvider = ({ children }: { children: ReactNode }) => {
  const [isImporting, setImporting] = useState(false);

  return (
    <ImportContext.Provider value={{ isImporting, setImporting }}>
      {children}
    </ImportContext.Provider>
  );
};
