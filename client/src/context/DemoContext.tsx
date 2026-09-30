import { createContext, useEffect, useState } from "react";
import type { ReactNode } from "react";

const configuredDemoMode =
  import.meta.env.VITE_DEMO_MODE === "true" ||
  import.meta.env.VITE_DEMO_MODE === "1";

export const DemoContext = createContext<{
  isDemo: boolean;
  isDemoResolved: boolean;
}>({
  isDemo: false,
  isDemoResolved: false,
});

if (!import.meta.env.VITE_SERVER_URL) {
  throw new Error(
    "VITE_SERVER_URL environment variable is not set, and is required.",
  );
}

export const SERVER_URL = `${import.meta.env.VITE_SERVER_URL}`;

export const DemoProvider = ({ children }: { children: ReactNode }) => {
  const [isDemo, setIsDemo] = useState(configuredDemoMode);
  const [isDemoResolved, setIsDemoResolved] = useState(configuredDemoMode);

  useEffect(() => {
    let mounted = true;
    const fetchHealth = async () => {
      try {
        const res = await fetch(`${SERVER_URL}/health`);
        if (!res.ok) return;
        const j = await res.json();
        if (mounted) {
          setIsDemo(!!j.demo);
          setIsDemoResolved(true);
        }
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
    <DemoContext.Provider value={{ isDemo, isDemoResolved }}>
      {children}
    </DemoContext.Provider>
  );
};

export default DemoContext;
