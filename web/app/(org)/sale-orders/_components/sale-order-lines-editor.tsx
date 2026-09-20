"use client";

import { Button } from "@/components/button";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import type { Item, ItemVariant } from "@/lib/services/swantara";
import { createId, formatMoney } from "@/lib/utils";
import { resolvePricePreview } from "./sale-order-utils";

export type LineRow = {
  id: string;
  itemId: string;
  qtyOrdered: string;
  discountPct: string;
  description: string;
};

export function createEmptyLine(): LineRow {
  return {
    id: createId(),
    itemId: "",
    qtyOrdered: "1",
    discountPct: "0",
    description: "",
  };
}

type SaleOrderLinesEditorProps = {
  lines: LineRow[];
  products: Item[];
  variantsByProduct: Record<string, ItemVariant[]>;
  priceBookId: string;
  orderDate: string;
  onAdd: () => void;
  onUpdate: (id: string, patch: Partial<LineRow>) => void;
  onRemove: (id: string) => void;
  onSelectItem: (itemId: string) => void;
};

export function SaleOrderLinesEditor({
  lines,
  products,
  variantsByProduct,
  priceBookId,
  orderDate,
  onAdd,
  onUpdate,
  onRemove,
  onSelectItem,
}: SaleOrderLinesEditorProps) {
  return (
    <div className="mt-6 flex flex-col gap-3 rounded-md border p-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm">{"Order lines"}</h3>
        <Button type="button" variant="outline" size="sm" onClick={onAdd}>
          {"Add line"}
        </Button>
      </div>
      {lines.length === 0 ? (
        <p className="text-sm text-muted-foreground">{"Add at least one line."}</p>
      ) : null}
      <div className="flex flex-col gap-3">
        {lines.map((line, index) => {
          const template = products.find((product) => String(product.id) === line.itemId);
          const variants = variantsByProduct[line.itemId] ?? [];
          const preview =
            priceBookId && line.itemId
              ? resolvePricePreview(
                  null,
                  Number(line.itemId),
                  template?.categoryId ? Number(template.categoryId) : null,
                  Number(line.qtyOrdered) || 0,
                  orderDate ? new Date(orderDate) : null,
                  template?.listPrice ?? 0,
                  [],
                )
              : null;
          return (
            <div key={line.id} className="grid gap-2 rounded-md border p-3 sm:grid-cols-12">
              <div className="sm:col-span-5">
                <span className="text-xs text-muted-foreground">{"Item variant"}</span>
                <Select
                  value={line.itemId}
                  onValueChange={(value) => {
                    onUpdate(line.id, { itemId: value ?? "" });
                    if (value) onSelectItem(value);
                  }}
                >
                  <SelectTrigger aria-label={`${"Item variant"} ${index + 1}`}>
                    <SelectValue placeholder={"Select item"} />
                  </SelectTrigger>
                  <SelectContent>
                    {products.map((product) => (
                      <SelectItem key={product.id} value={String(product.id)}>
                        {product.name} —{" "}
                        {formatMoney(product.listPrice, { currency: DEFAULT_CURRENCY })}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {variants.length > 0 ? (
                  <p className="mt-1 text-xs text-muted-foreground">
                    {variants.length} variants — using template price
                  </p>
                ) : null}
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{"Qty"}</span>
                <Input
                  value={line.qtyOrdered}
                  onChange={(event) => onUpdate(line.id, { qtyOrdered: event.target.value })}
                  type="number"
                  min="0"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{"Discount %"}</span>
                <Input
                  value={line.discountPct}
                  onChange={(event) => onUpdate(line.id, { discountPct: event.target.value })}
                  type="number"
                  min="0"
                  max="100"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2 flex flex-col justify-end">
                <span className="text-xs text-muted-foreground">{"Price preview"}</span>
                <span className="text-sm tabular-nums">
                  {preview
                    ? formatMoney(preview.price, { currency: DEFAULT_CURRENCY })
                    : "Select item and price_book".slice(0, 20)}
                </span>
              </div>
              <div className="sm:col-span-1 flex items-end">
                <Button type="button" variant="ghost" size="sm" onClick={() => onRemove(line.id)}>
                  {"Remove"}
                </Button>
              </div>
              <div className="sm:col-span-12">
                <Input
                  value={line.description}
                  onChange={(event) => onUpdate(line.id, { description: event.target.value })}
                  placeholder={"Description"}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
