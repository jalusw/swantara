import type { Tax } from "@/lib/services/swantara";

export type TaxType = Tax["type"];
export type TaxScope = Tax["scope"];

export const taxTypeLabels: Record<TaxType, string> = {
  percent: "Percent",
  fixed: "Tetap",
  group: "Group",
};

export const taxScopeLabels: Record<TaxScope, string> = {
  sale: "Penjualan",
  purchase: "Purchase",
  none: "Tidak ada",
};

export function taxTypeTone(type: TaxType): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (type) {
    case "percent":
      return "info";
    case "fixed":
      return "warning";
    case "group":
      return "neutral";
    default:
      return "neutral";
  }
}

export function formatTaxAmount(tax: Tax): string {
  if (tax.type === "percent") return `${tax.amount ?? 0}%`;
  if (tax.type === "fixed") return `${tax.amount ?? 0}`;
  return "Group";
}
