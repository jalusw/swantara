import type { GiftCard } from "@/lib/services/swantara";

export type GiftCardState = GiftCard["state"];

export function canRedeem(state: GiftCardState): boolean {
  return state === "active";
}

export function canRefund(state: GiftCardState): boolean {
  return state === "active" || state === "used";
}

export function giftCardStateTone(
  state: GiftCardState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "active":
      return "success";
    case "used":
      return "info";
    case "expired":
      return "warning";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function formatGiftCardState(state: GiftCardState): string {
  switch (state) {
    case "active":
      return "Active";
    case "used":
      return "Used";
    case "expired":
      return "Expired";
    case "cancelled":
      return "Cancelled";
    default:
      return state;
  }
}
