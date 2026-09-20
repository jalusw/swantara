import type { JournalEntry } from "@/lib/services/swantara";

export type EntryState = JournalEntry["state"];

export function canPost(state: EntryState): boolean {
  return state === "draft";
}

export function canReverse(state: EntryState): boolean {
  return state === "posted";
}

export function canCancel(state: EntryState): boolean {
  return state === "draft";
}

export function moveStateTone(
  state: EntryState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "posted":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function moveBalance(lines: { debit: number; credit: number }[]): {
  totalDebit: number;
  totalCredit: number;
  isBalanced: boolean;
} {
  const totalDebit = lines.reduce((sum, l) => sum + l.debit, 0);
  const totalCredit = lines.reduce((sum, l) => sum + l.credit, 0);
  return {
    totalDebit,
    totalCredit,
    isBalanced: Math.abs(totalDebit - totalCredit) < 0.01,
  };
}
