"use client";

import React, {createContext, useContext, useEffect, useMemo, useState} from "react";
import {SERVER_URL} from "@/lib/config";
import {defaultSettings, RegionalSettings} from "@/lib/formatters";

type SettingsContextValue = {
  settings: RegionalSettings;
  loading: boolean;
  refresh: () => Promise<void>;
  update: (updates: Partial<RegionalSettings>) => Promise<RegionalSettings | null>;
};

const SettingsContext = createContext<SettingsContextValue | undefined>(undefined);

export const SettingsProvider: React.FC<{ children: React.ReactNode }> = ({children}) => {
  const [settings, setSettings] = useState<RegionalSettings>(defaultSettings);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    try {
      setLoading(true);
      const resp = await fetch(`${SERVER_URL}/settings`);
      if (!resp.ok) throw new Error("Failed to fetch settings");
      const data: RegionalSettings = await resp.json();
      setSettings({...defaultSettings, ...data});
    } catch (err) {
      console.error("Error fetching settings:", err);
    } finally {
      setLoading(false);
    }
  };

  const update = async (updates: Partial<RegionalSettings>) => {
    try {
      const resp = await fetch(`${SERVER_URL}/settings`, {
        method: "PUT",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(updates),
      });
      if (!resp.ok) throw new Error("Failed to update settings");
      const data: RegionalSettings = await resp.json();
      setSettings({...defaultSettings, ...data});
      return data;
    } catch (err) {
      console.error("Error updating settings:", err);
      return null;
    }
  };

  useEffect(() => {
    void refresh();
  }, []);

  const value = useMemo(
    () => ({
      settings,
      loading,
      refresh,
      update,
    }),
    [settings, loading]
  );

  return <SettingsContext.Provider value={value}>{children}</SettingsContext.Provider>;
};

export const useSettings = () => {
  const ctx = useContext(SettingsContext);
  if (!ctx) {
    throw new Error("useSettings must be used within SettingsProvider");
  }
  return ctx;
};
