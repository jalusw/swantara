import type { Payment } from "@/lib/services/swantara";

export type PaymentState = Payment["state"];
export type PaymentDirection = Payment["type"];

export function paymentStateLabel(state: PaymentState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "posted":
      return "Posted";
    case "reconciled":
      return "Reconciled";
    case "cancelled":
      return "Cancelled";
    default:
      return state;
  }
}

export function paymentDirectionLabel(direction: PaymentDirection): string {
  switch (direction) {
    case "inbound":
      return "Inbound";
    case "outbound":
      return "Outbound";
    default:
      return direction;
  }
}

export function paymentStateTone(
  state: PaymentState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "posted":
      return "success";
    case "reconciled":
      return "info";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}
