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
  measurementSystem: "metric",
  weekStart: 0,
  dateFormat: "YYYY-MM-DD",
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

const formatWithPattern = (parts: { year: string; month: string; day: string }, pattern: string) => {
  const tokens: Record<string, string> = {
    YYYY: parts.year,
    MM: parts.month,
    DD: parts.day,
  };
  return pattern.replace(/YYYY|MM|DD/g, (token) => tokens[token] || token);
};

const getDateParts = (value: string, settings: RegionalSettings) => {
  if (!value) return null;
  const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (dateOnly) {
    return { year: dateOnly[1], month: dateOnly[2], day: dateOnly[3] };
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;

  const formatter = new Intl.DateTimeFormat(settings.locale, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    timeZone: settings.timeZone || "UTC",
  });
  const parts = formatter.formatToParts(date);
  const lookup = Object.fromEntries(parts.map((p) => [p.type, p.value]));
  return {
    year: lookup.year,
    month: lookup.month,
    day: lookup.day,
  };
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
  const parts = getDateParts(value, settings);
  if (!parts) return value;

  const pattern = settings.dateFormat?.trim();
  if (!pattern || pattern.toLowerCase() === "auto") {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    const options: Intl.DateTimeFormatOptions = {
      year: "numeric",
      month: "short",
      day: "numeric",
      timeZone: settings.timeZone || "UTC",
    };
    return new Intl.DateTimeFormat(settings.locale, options).format(date);
  }

  return formatWithPattern(parts, pattern);
};

export const formatDateTime = (value: string, settings: RegionalSettings) => {
  if (!value) return "";
  const datePart = formatDate(value, settings);
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return datePart;
  const time = new Intl.DateTimeFormat(settings.locale, {
    hour: "2-digit",
    minute: "2-digit",
    timeZone: settings.timeZone || "UTC",
  }).format(date);
  return time ? `${datePart} ${time}` : datePart;
};
