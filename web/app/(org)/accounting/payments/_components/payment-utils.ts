import type { Payment } from "@/lib/services/swantara";

export type PaymentState = Payment["state"];
export type PaymentDirection = Payment["type"];

export function paymentStateLabel(state: PaymentState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "posted":
      return "Diposting";
    case "reconciled":
      return "Direkonsiliasi";
    case "cancelled":
      return "Dibatalkan";
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
