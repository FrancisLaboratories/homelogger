import { useState } from "react";
import { Form, InputGroup } from "react-bootstrap";

interface CurrencyInputProps {
  value: string;
  currency: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

const getUserLocale = () =>
  typeof navigator !== "undefined" && navigator.language
    ? navigator.language
    : "en-US";

const formatInputValue = (value: string, currency: string) => {
  if (value.trim() === "") return "";
  const numericValue = Number(value.replaceAll(",", ""));
  if (!Number.isFinite(numericValue)) return value;

  const currencyOptions = new Intl.NumberFormat(getUserLocale(), {
    style: "currency",
    currency,
  }).resolvedOptions();
  return new Intl.NumberFormat(getUserLocale(), {
    minimumFractionDigits: currencyOptions.minimumFractionDigits,
    maximumFractionDigits: currencyOptions.maximumFractionDigits,
  }).format(numericValue);
};

const parseDisplayValue = (value: string) => {
  const parts = new Intl.NumberFormat(getUserLocale()).formatToParts(12345.6);
  const groupSeparator = parts.find((part) => part.type === "group")?.value;
  const decimalSeparator = parts.find((part) => part.type === "decimal")?.value;

  return value
    .replaceAll(groupSeparator ?? ",", "")
    .replace(decimalSeparator ?? ".", ".");
};

const normalizeValue = (value: string, currency: string) => {
  const parsedValue = Number(parseDisplayValue(value).trim());
  if (!Number.isFinite(parsedValue)) return null;

  const formattedValue = formatInputValue(String(parsedValue), currency);
  return String(Number(parseDisplayValue(formattedValue)));
};

const CurrencyInput: React.FC<CurrencyInputProps> = ({
  value,
  currency,
  onChange,
  placeholder,
}) => {
  const [isFocused, setIsFocused] = useState(false);
  const [displayValue, setDisplayValue] = useState(value);

  const handleBlur = () => {
    const rawValue = displayValue.trim();
    const normalizedValue = normalizeValue(rawValue, currency);

    if (rawValue === "" || normalizedValue === null) {
      const parsedValue = parseDisplayValue(rawValue).trim();
      onChange(parsedValue);
      setDisplayValue(parsedValue);
    } else {
      onChange(normalizedValue);
      setDisplayValue(formatInputValue(normalizedValue, currency));
    }
    setIsFocused(false);
  };

  return (
    <InputGroup>
      <InputGroup.Text>{currency}</InputGroup.Text>
      <Form.Control
        type="text"
        inputMode="decimal"
        value={isFocused ? displayValue : formatInputValue(value, currency)}
        placeholder={placeholder}
        onFocus={() => {
          setIsFocused(true);
          setDisplayValue(formatInputValue(value, currency));
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
