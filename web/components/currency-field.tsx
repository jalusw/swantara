"use client";

import { useCallback, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { InputGroup, InputGroupAddon, InputGroupInput, InputGroupText } from "./input-group";

export type CurrencyFormatOptions = {
  locale?: string;
  currency?: string;
  maximumFractionDigits?: number;
};

export function formatCurrency(value: number, options: CurrencyFormatOptions = {}) {
  const { locale = "en-US", currency = "USD", maximumFractionDigits = 2 } = options;
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    maximumFractionDigits,
  }).format(value);
}

function formatDecimal(value: number, locale?: string) {
  return new Intl.NumberFormat(locale, {
    style: "decimal",
    maximumFractionDigits: 0,
  }).format(value);
}

function parseCurrencyInput(raw: string): number | null {
  const trimmed = raw.trim();
  if (!trimmed || trimmed === "-" || trimmed === ".") {
    return null;
  }
  const negative = trimmed.startsWith("-") ? -1 : 1;
  const digits = trimmed.replace(/[^\d.,-]/g, "").replace(/,/g, "");
  if (digits === "" || digits === "-") {
    return null;
  }
  const value = Number(digits) * negative;
  return Number.isNaN(value) ? null : value;
}

export type CurrencyViewProps = {
  value: number | null;
  locale?: string;
  currency?: string;
  placeholder?: string;
  className?: string;
};

export function CurrencyView({
  value,
  locale,
  currency,
  placeholder = "—",
  className,
}: CurrencyViewProps) {
  return (
    <span
      data-slot="currency-view"
      className={cn("tabular-nums", value == null && "text-muted-foreground", className)}
    >
      {value == null ? placeholder : formatCurrency(value, { locale, currency })}
    </span>
  );
}

export type CurrencyFieldProps = Omit<
  React.ComponentProps<typeof InputGroupInput>,
  "value" | "onChange" | "defaultValue"
> & {
  value: number | null;
  onValueChange?: (value: number | null) => void;
  locale?: string;
  currency?: string;
};

export function CurrencyField({
  value,
  onValueChange,
  locale,
  currency,
  onFocus,
  onBlur,
  className,
  ...props
}: CurrencyFieldProps) {
  const [editing, setEditing] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const formatEditing = useCallback(
    (raw: number) => {
      return raw === 0 ? "" : formatDecimal(raw, locale);
    },
    [locale],
  );

  const display = editing !== null ? editing : value == null ? "" : formatDecimal(value, locale);

  const handleChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const input = event.target;
      const cursorPos = input.selectionStart ?? input.value.length;
      const prevLength = input.value.length;

      const raw = parseCurrencyInput(event.target.value);
      const formatted = raw != null ? formatEditing(raw) : "";

      setEditing(formatted);
      onValueChange?.(raw);

      requestAnimationFrame(() => {
        const diff = formatted.length - prevLength;
        const newCursorPos = Math.max(0, Math.min(cursorPos + diff, formatted.length));
        input.setSelectionRange(newCursorPos, newCursorPos);
      });
    },
    [formatEditing, onValueChange],
  );

  const handleFocus = useCallback(
    (event: React.FocusEvent<HTMLInputElement>) => {
      const formatted = value == null ? "" : formatEditing(value);
      setEditing(formatted);
      onFocus?.(event);
    },
    [value, formatEditing, onFocus],
  );

  const handleBlur = useCallback(
    (event: React.FocusEvent<HTMLInputElement>) => {
      setEditing(null);
      onBlur?.(event);
    },
    [onBlur],
  );

  return (
    <InputGroup className={cn("group/input-group", className)}>
      <InputGroupAddon align="inline-start">
        <InputGroupText className="font-mono text-xs uppercase">{currency}</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput
        ref={inputRef}
        data-slot="currency-field"
        inputMode="decimal"
        autoComplete="off"
        placeholder="0"
        value={display}
        aria-invalid={props["aria-invalid"]}
        onChange={handleChange}
        onFocus={handleFocus}
        onBlur={handleBlur}
        className="tabular-nums"
        {...props}
      />
    </InputGroup>
  );
}

export { parseCurrencyInput as parseCurrency };
