"use client";

import { Minus, Plus, Trash2 } from "lucide-react";
import { formatMoney, formatNumber } from "@/lib/utils";
import type { CartLine } from "./pos-register-section";

export function PosCart({
  lines,
  subtotal,
  onUpdateLine,
  onRemoveLine,
  onClear,
}: {
  lines: CartLine[];
  subtotal: number;
  onUpdateLine: (key: string, updates: Partial<CartLine>) => void;
  onRemoveLine: (key: string) => void;
  onClear: () => void;
}) {
  if (lines.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center p-4">
        <p className="text-sm text-muted-foreground">{"Cart is empty."}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <div className="flex items-center justify-between border-b px-3 py-2">
        <span className="text-xs text-muted-foreground">{`${lines.length} item(s)`}</span>
        <button
          type="button"
          onClick={onClear}
          className="text-xs text-destructive hover:underline"
        >
          {"Clear"}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {lines.map((line) => (
          <div key={line.key} className="flex items-start gap-2 border-b px-3 py-2">
            <div className="flex-1 min-w-0">
              <p className="truncate text-sm">{line.productName}</p>
              <p className="text-xs text-muted-foreground tabular-nums">
                {formatMoney(line.unitPrice)}
                {line.discountPct > 0 ? ` (-${line.discountPct}%)` : ""}
              </p>
            </div>

            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={() =>
                  onUpdateLine(line.key, {
                    qty: Math.max(1, line.qty - 1),
                  })
                }
                className="flex h-7 w-7 items-center justify-center rounded border text-muted-foreground hover:bg-accent"
              >
                <Minus className="h-3 w-3" />
              </button>
              <span className="w-8 text-center text-sm tabular-nums">{formatNumber(line.qty)}</span>
              <button
                type="button"
                onClick={() => onUpdateLine(line.key, { qty: line.qty + 1 })}
                className="flex h-7 w-7 items-center justify-center rounded border text-muted-foreground hover:bg-accent"
              >
                <Plus className="h-3 w-3" />
              </button>
            </div>

            <div className="flex items-center gap-2">
              <span className="text-sm tabular-nums">
                {formatMoney(line.unitPrice * line.qty * (1 - line.discountPct / 100))}
              </span>
              <button
                type="button"
                onClick={() => onRemoveLine(line.key)}
                className="text-destructive hover:text-destructive/80"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="flex items-center justify-between border-t px-3 py-2">
        <span className="text-sm">{"Subtotal"}</span>
        <span className="text-base font-bold tabular-nums">{formatMoney(subtotal)}</span>
      </div>
    </div>
  );
}
