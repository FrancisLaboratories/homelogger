export const getFractionDigits = (value: number | string) => {
  const fraction = String(value).match(/\.(\d+)$/)?.[1];
  return fraction?.length ?? 0;
};

export const formatCurrency = (value: number, currency: string) => {
  const locale =
    typeof navigator !== "undefined" && navigator.language
      ? navigator.language
      : "en-US";
  const currencyOptions = new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
  }).resolvedOptions();

  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    currencyDisplay: "code",
    minimumFractionDigits: currencyOptions.minimumFractionDigits,
    maximumFractionDigits: Math.min(
      20,
      Math.max(
        currencyOptions.maximumFractionDigits ?? 0,
        getFractionDigits(value),
      ),
    ),
  }).format(value);
};
