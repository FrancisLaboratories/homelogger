export type RegionalSettings = {
  locale: string;
  language: string;
  currency: string;
  timeZone: string;
  measurementSystem: string;
  weekStart: number;
  dateFormat: string;
  numberingSystem: string;
};

export const defaultSettings: RegionalSettings = {
  locale: "en-US",
  language: "en",
  currency: "USD",
  timeZone: "UTC",
  measurementSystem: "imperial",
  weekStart: 0,
  dateFormat: "",
  numberingSystem: "latn",
};

const buildNumberFormatOptions = (settings: RegionalSettings) => {
  const options: Intl.NumberFormatOptions = {};
  if (settings.numberingSystem) {
    options.numberingSystem = settings.numberingSystem;
  }
  return options;
};

export const formatNumber = (value: number, settings: RegionalSettings) => {
  if (!Number.isFinite(value)) return String(value);
  return new Intl.NumberFormat(settings.locale, buildNumberFormatOptions(settings)).format(
    value
  );
};

export const formatCurrency = (value: number, settings: RegionalSettings) => {
  if (!Number.isFinite(value)) return String(value);
  const options: Intl.NumberFormatOptions = {
    style: "currency",
    currency: settings.currency || "USD",
    ...buildNumberFormatOptions(settings),
  };
  return new Intl.NumberFormat(settings.locale, options).format(value);
};

export const formatDate = (value: string, settings: RegionalSettings) => {
  if (!value) return "";
  const isDateOnly = /^\d{4}-\d{2}-\d{2}$/.test(value);
  const date = isDateOnly ? new Date(`${value}T00:00:00Z`) : new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  const options: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: settings.timeZone || "UTC",
  };
  return new Intl.DateTimeFormat(settings.locale, options).format(date);
};
