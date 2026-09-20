import type { Rma, RmaLine } from "@/lib/services/swantara";

export type RmaType = Rma["type"];
export type RmaState = Rma["state"];
export type Disposition = RmaLine["disposition"];

export function rmaTypeLabel(type: RmaType): string {
  switch (type) {
    case "customer_return":
      return "Customer Return";
    case "vendor_return":
      return "Supplier Return";
    default:
      return type;
  }
}

export function rmaStateLabel(state: RmaState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "confirmed":
      return "Confirmed";
    case "received":
      return "Received";
    case "refunded":
      return "Refunded";
    case "done":
      return "Done";
    case "cancelled":
      return "Cancelled";
    default:
      return state;
  }
}

export function dispositionLabel(disposition: Disposition): string {
  switch (disposition) {
    case "restock":
      return "Restock";
    case "scrap":
      return "Scrap";
    case "repair":
      return "Repair";
    case "replace":
      return "Replace";
    default:
      return disposition;
  }
}

export const rmaValidStates: RmaState[] = [
  "draft",
  "confirmed",
  "received",
  "refunded",
  "done",
  "cancelled",
];
export const dispositionValues: Disposition[] = ["restock", "scrap", "repair", "replace"];

export function canConfirm(state: RmaState): boolean {
  return state === "draft";
}

export function canReceive(state: RmaState): boolean {
  return state === "confirmed";
}

export function canRefund(state: RmaState): boolean {
  return state === "received";
}

export function canDone(state: RmaState): boolean {
  return state === "refunded";
}

export function canCancel(state: RmaState): boolean {
  return state === "draft" || state === "confirmed";
}

export function rmaStateTone(
  state: RmaState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "confirmed":
      return "info";
    case "received":
      return "info";
    case "refunded":
      return "success";
    case "done":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function dispositionTone(
  disposition: Disposition,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (disposition) {
    case "restock":
      return "success";
    case "scrap":
      return "danger";
    case "repair":
      return "warning";
    case "replace":
      return "info";
    default:
      return "neutral";
  }
}

export function rmaSteps(state: RmaState): string[] {
  const allSteps = ["draft", "confirmed", "received", "refunded", "done"];
  const stateIndex = allSteps.indexOf(state);
  if (state === "cancelled") return ["draft", "cancelled"];
  return allSteps.slice(0, stateIndex + 1);
}

export function totalQty(lines: RmaLine[]): number {
  return lines.reduce((sum, line) => sum + line.qty, 0);
}
