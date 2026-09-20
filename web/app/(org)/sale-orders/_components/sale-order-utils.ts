import type { PriceRule, SaleOrder, SaleOrderLine } from "@/lib/services/swantara";

export function remainingToShip(
  line: Pick<SaleOrderLine, "qtyOrdered" | "qtyDelivered" | "qtyReturns">,
): number {
  const netDelivered = line.qtyDelivered - (line.qtyReturns ?? 0);
  return Math.max(0, line.qtyOrdered - netDelivered);
}

export function remainingToBill(
  line: Pick<SaleOrderLine, "qtyOrdered" | "qtyInvoiced" | "qtyReturns">,
): number {
  const netInvoiced = line.qtyInvoiced - (line.qtyReturns ?? 0);
  return Math.max(0, line.qtyOrdered - netInvoiced);
}

export function orderRemainingToShip(
  _order: Pick<SaleOrder, "amountTotal">,
  lines: SaleOrderLine[],
): number {
  return lines.reduce((sum, line) => sum + remainingToShip(line), 0);
}

export function orderRemainingToBill(
  _order: Pick<SaleOrder, "amountTotal">,
  lines: SaleOrderLine[],
): number {
  return lines.reduce((sum, line) => sum + remainingToBill(line), 0);
}

export type SaleOrderState = SaleOrder["state"];

export function canSend(state: SaleOrderState): boolean {
  return state === "draft";
}

export function canConfirm(state: SaleOrderState): boolean {
  return state === "draft" || state === "sent";
}

export function canCancel(state: SaleOrderState): boolean {
  return state === "draft" || state === "sent" || state === "confirmed";
}

export function canDone(state: SaleOrderState): boolean {
  return state === "confirmed";
}

export function canEdit(state: SaleOrderState): boolean {
  return state === "draft";
}

export function canDeliver(state: SaleOrderState): boolean {
  return state === "confirmed";
}

export function canInvoice(state: SaleOrderState): boolean {
  return state === "confirmed" || state === "done";
}

export function saleOrderStateTone(
  state: SaleOrderState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "sent":
      return "info";
    case "confirmed":
      return "warning";
    case "done":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

const scopePrecedence: Record<PriceRule["appliesTo"], number> = {
  variant: 0,
  item: 1,
  category: 2,
  all: 3,
};

export function resolvePricePreview(
  variantId: number | null,
  itemId: number | null,
  categoryId: number | null,
  qty: number,
  date: Date | null,
  basePrice: number,
  rules: PriceRule[],
): { price: number; ruleId: number | null } {
  if (variantId == null && itemId == null) return { price: basePrice, ruleId: null };
  const candidates = rules.filter((rule) => {
    if (rule.minQty > qty) return false;
    if (rule.dateStart && date && new Date(rule.dateStart) > date) return false;
    if (rule.dateEnd && date && new Date(rule.dateEnd) < date) return false;
    if (rule.appliesTo === "variant" && rule.itemId !== variantId) return false;
    if (rule.appliesTo === "item" && rule.itemId !== itemId) return false;
    if (rule.appliesTo === "category" && rule.categoryId !== categoryId) return false;
    return true;
  });
  if (candidates.length === 0) return { price: basePrice, ruleId: null };
  candidates.sort((a, b) => scopePrecedence[a.appliesTo] - scopePrecedence[b.appliesTo]);
  const best = candidates[0]!;
  if (best.computeType === "fixed" && best.fixedPrice != null)
    return { price: best.fixedPrice, ruleId: best.id };
  if (best.computeType === "percent" && best.discountPct != null)
    return {
      price: Math.round(basePrice * (1 - best.discountPct / 100) * 100) / 100,
      ruleId: best.id,
    };
  return { price: basePrice, ruleId: best.id };
}

export function lineAmount(qty: number, unitPrice: number, discountPct: number): number {
  return Math.round(qty * unitPrice * (1 - discountPct / 100) * 100) / 100;
}
