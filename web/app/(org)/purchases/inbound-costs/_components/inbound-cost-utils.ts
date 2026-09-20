import type { InboundCost } from "@/lib/services/swantara";

export type InboundCostState = InboundCost["state"];

export function canPost(state: InboundCostState): boolean {
  return state === "draft";
}

export function canCancel(state: InboundCostState): boolean {
  return state === "draft";
}

export function inboundCostStateTone(
  state: InboundCostState,
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

export function formatSplitMethod(method: string): string {
  switch (method) {
    case "by_quantity":
      return "By quantity";
    case "by_weight":
      return "By weight";
    case "by_volume":
      return "By volume";
    case "by_value":
      return "By value";
    case "equal":
      return "Equal";
    default:
      return method;
  }
}
