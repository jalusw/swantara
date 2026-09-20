"use client";

import { MinusIcon, PlusIcon } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Input } from "./input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";

export type QuantityUnit = {
  value: string;
  label: string;
};

export type QuantityFieldProps = {
  value: number | null;
  onValueChange?: (value: number | null) => void;
  min?: number;
  max?: number;
  step?: number;
  units?: QuantityUnit[];
  unit?: string;
  onUnitChange?: (value: string) => void;
  disabled?: boolean;
  placeholder?: string;
  id?: string;
  className?: string;
  "aria-label"?: string;
};

function parseNumeric(raw: string): number | null {
  if (!raw.trim()) {
    return null;
  }
  const value = Number(raw);
  return Number.isNaN(value) ? null : value;
}

export function QuantityField({
  value,
  onValueChange,
  min,
  max,
  step = 1,
  units,
  unit,
  onUnitChange,
  disabled = false,
  placeholder = "0",
  id,
  className,
  "aria-label": ariaLabel,
}: QuantityFieldProps) {
  const [draft, setDraft] = useState<string | null>(null);
  const [notice, setNotice] = useState("");
  const resolvedAriaLabel = ariaLabel ?? "Quantity";

  const handleUnit = (value: string | null) => onUnitChange?.(value ?? "");
  const clamp = (candidate: number): number => {
    let next = candidate;
    if (min != null) {
      next = Math.max(min, next);
    }
    if (max != null) {
      next = Math.min(max, next);
    }
    return next;
  };

  const dir = (direction: 1 | -1) => {
    if (disabled) {
      return;
    }
    const base = value ?? 0;
    onValueChange?.(clamp(Number((base + direction * step).toFixed(8))));
    setDraft(null);
  };

  const onInputBlur = (raw: string) => {
    setDraft(null);
    const next = parseNumeric(raw);
    const clamped = next == null ? null : clamp(next);
    onValueChange?.(clamped);
    if (next != null && clamped != null && next !== clamped) {
      setNotice(`Value adjusted to ${clamped}.`);
    } else {
      setNotice("");
    }
  };

  const display = draft ?? (value == null ? "" : String(value));
  const atMin = min != null && value != null && value <= min;
  const atMax = max != null && value != null && value >= max;

  return (
    <div
      data-slot="quantity-field"
      className={cn(
        "flex items-stretch overflow-hidden rounded-md border border-input bg-transparent transition-colors focus-within:border-ring focus-within:ring-3 ring-offset-2 ring-offset-background focus-within:ring-ring",
        disabled && "cursor-not-allowed bg-input/50 opacity-50",
        className,
      )}
    >
      <output className="sr-only">{notice}</output>
      <Button
        variant="ghost"
        size="icon"
        className="min-h-11 w-11 self-stretch rounded-none border-r border-input"
        aria-label={"Decrease quantity"}
        disabled={disabled || atMin}
        onClick={() => dir(-1)}
      >
        <MinusIcon aria-hidden />
      </Button>
      <Input
        id={id}
        type="text"
        inputMode="decimal"
        autoComplete="off"
        value={display}
        aria-label={resolvedAriaLabel}
        placeholder={placeholder}
        disabled={disabled}
        onChange={(event) => {
          setDraft(event.target.value);
          const next = parseNumeric(event.target.value);
          onValueChange?.(next == null ? null : clamp(next));
        }}
        onBlur={(event) => onInputBlur(event.target.value)}
        className="min-h-11 w-16 min-w-0 flex-1 rounded-none border-none p-0 text-center tabular-nums shadow-none focus-visible:ring-0"
      />
      <Button
        variant="ghost"
        size="icon"
        className="min-h-11 w-11 self-stretch rounded-none border-l border-input"
        aria-label={"Increase quantity"}
        disabled={disabled || atMax}
        onClick={() => dir(1)}
      >
        <PlusIcon aria-hidden />
      </Button>
      {units && units.length > 0 ? (
        <Select value={unit} onValueChange={handleUnit} disabled={disabled}>
          <SelectTrigger
            size="sm"
            className="min-h-11 w-16 justify-center rounded-none border-0 border-l border-input bg-transparent"
          >
            <SelectValue placeholder="—" />
          </SelectTrigger>
          <SelectContent align="end">
            {units.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ) : null}
    </div>
  );
}
