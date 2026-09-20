import type { CommissionEntry, CommissionPlan } from "@/lib/services/swantara";

export type CommissionEntryState = CommissionEntry["state"];

export function canConfirm(state: CommissionEntryState): boolean {
  return state === "draft";
}

export function canPay(state: CommissionEntryState): boolean {
  return state === "confirmed";
}

export function canCancelEntry(state: CommissionEntryState): boolean {
  return state === "draft" || state === "confirmed";
}

export function commissionEntryStateTone(
  state: CommissionEntryState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "confirmed":
      return "info";
    case "paid":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function formatBasis(basis: CommissionPlan["basis"]): string {
  switch (basis) {
    case "revenue":
      return "Revenue";
    case "margin":
      return "Margin";
    case "collected":
      return "Collected";
    default:
      return basis;
  }
}
