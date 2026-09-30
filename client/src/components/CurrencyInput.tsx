import { useState } from "react";
import { Form, InputGroup } from "react-bootstrap";

interface CurrencyInputProps {
  value: string;
  currency: string;
  onChange: (value: string) => void;
  placeholder?: string;
  min?: string;
  step?: string;
}

const getCurrencySymbol = (currency: string) => {
  try {
    return (
      new Intl.NumberFormat(undefined, {
        style: "currency",
        currency,
        currencyDisplay: "symbol",
      })
        .formatToParts(0)
        .find((part) => part.type === "currency")?.value || currency
    );
  } catch {
    return currency;
  }
};

const formatInputValue = (value: string, currency: string) => {
  if (value.trim() === "") return "";
  const numericValue = Number(value.replaceAll(",", ""));
  if (!Number.isFinite(numericValue)) return value;

  const currencyOptions = new Intl.NumberFormat(undefined, {
    style: "currency",
    currency,
  }).resolvedOptions();
  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: currencyOptions.minimumFractionDigits,
    maximumFractionDigits: currencyOptions.maximumFractionDigits,
  }).format(numericValue);
};

const parseDisplayValue = (value: string) => {
  const parts = new Intl.NumberFormat(undefined).formatToParts(12345.6);
  const groupSeparator = parts.find((part) => part.type === "group")?.value;
  const decimalSeparator = parts.find((part) => part.type === "decimal")?.value;

  return value
    .replaceAll(groupSeparator ?? ",", "")
    .replace(decimalSeparator ?? ".", ".");
};

const CurrencyInput: React.FC<CurrencyInputProps> = ({
  value,
  currency,
  onChange,
  placeholder,
  min,
  step,
}) => {
  const [isFocused, setIsFocused] = useState(false);
  const [displayValue, setDisplayValue] = useState(value);

  const handleBlur = () => {
    const rawValue = parseDisplayValue(displayValue).trim();
    const numericValue = Number(rawValue);

    if (rawValue === "" || !Number.isFinite(numericValue)) {
      onChange(rawValue);
      setDisplayValue(rawValue);
    } else {
      const normalizedValue = String(numericValue);
      onChange(normalizedValue);
      setDisplayValue(formatInputValue(normalizedValue, currency));
    }
    setIsFocused(false);
  };

  return (
    <InputGroup>
      <InputGroup.Text>{getCurrencySymbol(currency)}</InputGroup.Text>
      <Form.Control
        type="text"
        inputMode="decimal"
        min={min}
        step={step}
        value={isFocused ? displayValue : formatInputValue(value, currency)}
        placeholder={placeholder}
        onFocus={() => {
          setIsFocused(true);
          setDisplayValue(value);
        }}
        onChange={(event) => {
          setDisplayValue(event.target.value);
          onChange(event.target.value);
        }}
        onBlur={handleBlur}
      />
    </InputGroup>
  );
};

export default CurrencyInput;
