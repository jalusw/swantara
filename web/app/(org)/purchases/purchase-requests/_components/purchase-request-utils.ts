import type { PurchaseRequest } from "@/lib/services/swantara";

export type RequisitionState = PurchaseRequest["state"];

export function canConfirm(state: RequisitionState): boolean {
  return state === "draft";
}

export function canApprove(state: RequisitionState): boolean {
  return state === "confirmed";
}

export function canCancel(state: RequisitionState): boolean {
  return state === "draft" || state === "confirmed";
}

export function canCreateQuoteRequest(state: RequisitionState): boolean {
  return state === "approved";
}

export function canEdit(state: RequisitionState): boolean {
  return state === "draft";
}

export function requisitionStateTone(
  state: RequisitionState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "confirmed":
      return "info";
    case "approved":
      return "success";
    case "done":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function requisitionTotalQty(lines: { qty: number }[]): number {
  return lines.reduce((sum, line) => sum + line.qty, 0);
}
