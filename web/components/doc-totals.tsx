"use client";

import { cn } from "@/lib/utils";
import { formatCurrency } from "./currency-field";

export type DocTotalLine = {
  label: string;
  amount: number;
};

export type DocTotalsProps = {
  subtotal: number;
  discount?: number;
  taxes?: DocTotalLine[];
  shipping?: number;
  currency?: string;
  locale?: string;
  className?: string;
};

export function DocTotals({
  subtotal,
  discount,
  taxes = [],
  shipping,
  currency,
  locale,
  className,
}: DocTotalsProps) {
  const discountValue = discount ?? 0;
  const shippingValue = shipping ?? 0;
  const taxTotal = taxes.reduce((sum, line) => sum + line.amount, 0);
  const total = subtotal - discountValue + taxTotal + shippingValue;

  const money = (value: number) => formatCurrency(value, { locale, currency });

  return (
    <dl data-slot="doc-totals" className={cn("w-full space-y-2", className)}>
      <Row label={"Subtotal"} value={money(subtotal)} />
      {discountValue !== 0 ? (
        <Row label={"Discount"} value={`-${money(Math.abs(discountValue))}`} muted />
      ) : null}
      {taxes.map((tax) => (
        <Row key={tax.label} label={tax.label} value={money(tax.amount)} muted />
      ))}
      {shippingValue !== 0 ? <Row label={"Shipping"} value={money(shippingValue)} muted /> : null}
      <div className="flex items-baseline justify-between gap-4 border-t pt-3">
        <dt className="text-sm">{"Total"}</dt>
        <dd className="font-heading text-lg tracking-tight tabular-nums">{money(total)}</dd>
      </div>
    </dl>
  );
}

function Row({ label, value, muted = false }: { label: string; value: string; muted?: boolean }) {
  return (
    <div className="flex items-baseline justify-between gap-4">
      <dt className={cn("text-sm", muted ? "text-muted-foreground" : "text-foreground")}>
        {label}
      </dt>
      <dd className={cn("text-sm tabular-nums", muted ? "text-muted-foreground" : "")}>{value}</dd>
    </div>
  );
}
