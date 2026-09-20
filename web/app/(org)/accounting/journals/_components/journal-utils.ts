import type { Journal } from "@/lib/services/swantara";

export type JournalType = Journal["type"];

export const journalTypeLabels: Record<JournalType, string> = {
  sale: "Sale",
  purchase: "Purchase",
  bank: "Bank",
  cash: "Cash",
  general: "General",
};

export function journalTypeTone(
  type: JournalType,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (type) {
    case "sale":
      return "success";
    case "purchase":
      return "info";
    case "bank":
    case "cash":
      return "warning";
    case "general":
      return "neutral";
    default:
      return "neutral";
  }
}
