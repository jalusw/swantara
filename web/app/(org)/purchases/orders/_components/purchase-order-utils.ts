import type { PurchaseOrder, PurchaseOrderLine } from "@/lib/services/swantara";

export type PoState = PurchaseOrder["state"];

export function canConfirm(state: PoState): boolean {
  return state === "draft";
}

export function canCancel(state: PoState): boolean {
  return state === "draft" || state === "sent";
}

export function canReceive(state: PoState): boolean {
  return state === "confirmed" || state === "done";
}

export function canBill(state: PoState): boolean {
  return state === "confirmed" || state === "done";
}

export function canPay(state: PoState): boolean {
  return state === "confirmed" || state === "done";
}

export function canEdit(state: PoState): boolean {
  return state === "draft";
}

export function poStateTone(state: PoState): "neutral" | "success" | "warning" | "danger" | "info" {
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

export function remainingToReceive(
  line: Pick<PurchaseOrderLine, "qtyOrdered" | "qtyReceived" | "qtyReturns">,
): number {
  const netReceived = line.qtyReceived - (line.qtyReturns ?? 0);
  return Math.max(0, line.qtyOrdered - netReceived);
}

export function remainingToBill(
  line: Pick<PurchaseOrderLine, "qtyOrdered" | "qtyBilled" | "qtyReturns">,
): number {
  const netBilled = line.qtyBilled - (line.qtyReturns ?? 0);
  return Math.max(0, line.qtyOrdered - netBilled);
}

export function orderRemainingToReceive(lines: PurchaseOrderLine[]): number {
  return lines.reduce((sum, line) => sum + remainingToReceive(line), 0);
}

export function orderRemainingToBill(lines: PurchaseOrderLine[]): number {
  return lines.reduce((sum, line) => sum + remainingToBill(line), 0);
}

export function threeWayMatchStatus(
  line: PurchaseOrderLine,
): "matched" | "over_received" | "over_billed" | "partial" {
  const received = remainingToReceive(line);
  const billed = remainingToBill(line);
  if (line.qtyBilled > line.qtyOrdered) return "over_billed";
  if (line.qtyReceived > line.qtyOrdered) return "over_received";
  if (received === 0 && billed === 0) return "matched";
  return "partial";
}
