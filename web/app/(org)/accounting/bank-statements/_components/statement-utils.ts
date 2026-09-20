import type { BankStatement } from "@/lib/services/swantara";

export type BankStatementState = BankStatement["state"];

export function bankStatementStateLabel(state: BankStatementState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "open":
      return "Open";
    case "reconciled":
      return "Reconciled";
    case "cancelled":
      return "Cancelled";
    default:
      return state;
  }
}

export function bankStatementStateTone(
  state: BankStatementState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "open":
      return "info";
    case "reconciled":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function statementBalance(balanceStart: number, lines: { amount: number }[]): number {
  return lines.reduce((sum, l) => sum + l.amount, balanceStart);
}
