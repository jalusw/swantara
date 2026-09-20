"use client";

import { PlusIcon, Trash2Icon } from "lucide-react";
import { cn, createId } from "@/lib/utils";
import { Button } from "./button";
import { CurrencyView, formatCurrency } from "./currency-field";
import { EmptyState } from "./empty-state";
import { Input } from "./input";
import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "./table";

export type LineItem = {
  id: string;
  description?: string;
  quantity: number;
  unitPrice: number;
};

export type LineItemsTableProps = {
  items: LineItem[];
  onChange?: (items: LineItem[]) => void;
  currency?: string;
  locale?: string;
  disabled?: boolean;
  addLabel?: string;
  productPlaceholder?: string;
  className?: string;
};

export function LineItemsTable({
  items,
  onChange,
  currency,
  locale,
  disabled = false,
  addLabel,
  productPlaceholder,
  className,
}: LineItemsTableProps) {
  const editable = !disabled && Boolean(onChange);

  function update(id: string, patch: Partial<LineItem>) {
    onChange?.(items.map((item) => (item.id === id ? { ...item, ...patch } : item)));
  }

  function remove(id: string) {
    onChange?.(items.filter((item) => item.id !== id));
  }

  function addLine() {
    onChange?.([...items, { id: createId(), description: "", quantity: 1, unitPrice: 0 }]);
  }

  const quantity = (value: string) => (value.trim() ? Math.max(0, Number(value) || 0) : 0);
  const amount = (item: LineItem) => item.quantity * item.unitPrice;
  const totalAmount = items.reduce((sum, item) => sum + amount(item), 0);

  return (
    <div data-slot="line-items-table" className={cn("space-y-2", className)}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-full">{"Item"}</TableHead>
            <TableHead className="text-center">{"Qty"}</TableHead>
            <TableHead className="text-right">{"Unit price"}</TableHead>
            <TableHead className="text-right">{"Amount"}</TableHead>
            {editable ? <TableHead className="w-10" /> : null}
          </TableRow>
        </TableHeader>
        {items.length > 0 ? (
          <TableBody>
            {items.map((item, index) => (
              <TableRow key={item.id}>
                <TableCell className="whitespace-normal">
                  {editable ? (
                    <Input
                      value={item.description}
                      onChange={(event) => update(item.id, { description: event.target.value })}
                      placeholder={productPlaceholder}
                      aria-label={`Description for line item ${index + 1}`}
                      className="min-h-11 w-full rounded-sm"
                    />
                  ) : (
                    <span className="">{item.description || "—"}</span>
                  )}
                </TableCell>
                <TableCell className="w-16 text-center">
                  {editable ? (
                    <Input
                      type="number"
                      inputMode="decimal"
                      min={0}
                      value={item.quantity}
                      onChange={(event) =>
                        update(item.id, {
                          quantity: quantity(event.target.value),
                        })
                      }
                      aria-label={`Quantity for line item ${index + 1}`}
                      className="min-h-11 w-16 rounded-sm px-1 text-center tabular-nums"
                    />
                  ) : (
                    <span className="tabular-nums">{item.quantity}</span>
                  )}
                </TableCell>
                <TableCell className="w-28 text-right">
                  {editable ? (
                    <Input
                      type="number"
                      inputMode="decimal"
                      min={0}
                      step={0.01}
                      value={item.unitPrice}
                      onChange={(event) =>
                        update(item.id, {
                          unitPrice: quantity(event.target.value),
                        })
                      }
                      aria-label={`Unit price for line item ${index + 1}`}
                      className="min-h-11 w-28 rounded-sm px-2 text-right tabular-nums"
                    />
                  ) : (
                    <CurrencyView value={item.unitPrice} currency={currency} locale={locale} />
                  )}
                </TableCell>
                <TableCell className="w-28 text-right">
                  <span className=" tabular-nums">
                    {formatCurrency(amount(item), { locale, currency })}
                  </span>
                </TableCell>
                {editable ? (
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="min-h-11 w-11 text-destructive"
                      aria-label={`Remove line item ${index + 1}${
                        item.description ? `: ${item.description}` : ""
                      }`}
                      onClick={() => remove(item.id)}
                    >
                      <Trash2Icon aria-hidden />
                    </Button>
                  </TableCell>
                ) : null}
              </TableRow>
            ))}
          </TableBody>
        ) : null}
        {items.length === 0 ? (
          <TableBody>
            <TableRow>
              <TableCell colSpan={editable ? 5 : 4} className="p-0">
                <EmptyState
                  title={"No line items"}
                  description={"Add products or services to this document."}
                />
              </TableCell>
            </TableRow>
          </TableBody>
        ) : null}
        <TableFooter>
          <TableRow>
            <TableCell colSpan={3} className="text-right">
              {"Total"}
            </TableCell>
            <TableCell colSpan={editable ? 2 : 1} className="text-right">
              <span className=" tabular-nums">
                {formatCurrency(totalAmount, { locale, currency })}
              </span>
            </TableCell>
          </TableRow>
        </TableFooter>
      </Table>
      {editable ? (
        <Button variant="outline" size="sm" onClick={addLine}>
          <PlusIcon aria-hidden />
          {addLabel ?? "Add line"}
        </Button>
      ) : null}
    </div>
  );
}
