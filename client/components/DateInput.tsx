import React, {useMemo, useRef} from "react";
import {Button, Form} from "react-bootstrap";
import {RegionalSettings, formatDate, getDatePattern, parseDateInput} from "@/lib/formatters";

interface DateInputProps {
    id?: string;
    value: string;
    onChange: (value: string) => void;
    settings: RegionalSettings;
}

const DateInput: React.FC<DateInputProps> = ({id, value, onChange, settings}) => {
    const datePattern = getDatePattern(settings);
    const nativeRef = useRef<HTMLInputElement>(null);

    const isoValue = useMemo(() => parseDateInput(value, settings) ?? "", [value, settings]);

    const openPicker = () => {
        const el = nativeRef.current;
        if (!el) return;
        const picker = (el as HTMLInputElement & { showPicker?: () => void }).showPicker;
        if (picker) {
            picker.call(el);
            return;
        }
        el.click();
    };

    return (
        <div style={{display: "flex", gap: "0.5rem", alignItems: "center"}}>
            <Form.Control
                id={id}
                type="text"
                value={value}
                placeholder={datePattern}
                title={`Format: ${datePattern}`}
                onChange={(e) => onChange(e.target.value)}
            />
            <Button variant="outline-secondary" type="button" onClick={openPicker}>
                <i className="bi bi-calendar3" />
            </Button>
            <Form.Control
                ref={nativeRef}
                type="date"
                value={isoValue}
                onChange={(e) => {
                    const iso = e.target.value;
                    onChange(iso ? formatDate(iso, settings) : "");
                }}
                tabIndex={-1}
                aria-hidden="true"
                style={{position: "absolute", opacity: 0, width: 1, height: 1, pointerEvents: "none"}}
            />
        </div>
    );
};

export default DateInput;
