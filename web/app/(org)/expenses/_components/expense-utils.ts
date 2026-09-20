import type { ExpenseReport } from "@/lib/services/swantara";

export type ExpenseReportState = ExpenseReport["state"];

export function canSubmit(state: ExpenseReportState): boolean {
  return state === "draft";
}

export function canApprove(state: ExpenseReportState): boolean {
  return state === "submitted";
}

export function canRefuse(state: ExpenseReportState): boolean {
  return state === "submitted";
}

export function canPost(state: ExpenseReportState): boolean {
  return state === "approved";
}

export function canReimburse(state: ExpenseReportState): boolean {
  return state === "posted";
}

export function canBill(state: ExpenseReportState): boolean {
  return state === "posted";
}

export function expenseStateTone(
  state: ExpenseReportState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "submitted":
      return "info";
    case "approved":
      return "success";
    case "posted":
      return "warning";
    case "reimbursed":
      return "success";
    case "refused":
      return "danger";
    default:
      return "neutral";
  }
}
