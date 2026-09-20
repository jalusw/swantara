import type { DeferralSchedule } from "@/lib/services/swantara";

export type DeferralState = DeferralSchedule["state"];

export function canRecognize(state: DeferralState): boolean {
  return state === "running";
}

export function canCancel(state: DeferralState): boolean {
  return state === "draft" || state === "running";
}

export function deferralStateTone(
  state: DeferralState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "running":
      return "success";
    case "done":
      return "info";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function formatDeferralType(type: DeferralSchedule["type"]): string {
  switch (type) {
    case "deferred_revenue":
      return "Deferred Revenue";
    case "deferred_expense":
      return "Deferred Expense";
    case "prepaid":
      return "Prepaid";
    default:
      return type;
  }
}
