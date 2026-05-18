import { createContext, useEffect, useState } from "react";
import type { ReactNode } from "react";

export const DemoContext = createContext<{ isDemo: boolean }>({
  isDemo: false,
});

if (!import.meta.env.VITE_SERVER_URL) {
  throw new Error(
    "VITE_SERVER_URL environment variable is not set, and is required.",
  );
}

export const SERVER_URL = `${import.meta.env.VITE_SERVER_URL}`;

export const DemoProvider = ({ children }: { children: ReactNode }) => {
  const [isDemo, setIsDemo] = useState(false);

  useEffect(() => {
    let mounted = true;
    const fetchHealth = async () => {
      try {
        const res = await fetch(`${SERVER_URL}/health`);
        if (!res.ok) return;
        const j = await res.json();
        if (mounted) setIsDemo(!!j.demo);
      } catch {
        // ignore network errors
      }
    };
    fetchHealth();
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <DemoContext.Provider value={{ isDemo }}>{children}</DemoContext.Provider>
  );
};

export default DemoContext;
