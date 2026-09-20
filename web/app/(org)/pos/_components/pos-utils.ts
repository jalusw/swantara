import type { StatusConfig } from "@/components/state-badge";
import type { PosConfig, PosOrder, PosOrderLine, PosSession } from "@/lib/services/swantara";

export type PosSessionState = PosSession["state"];
export type PosOrderState = PosOrder["state"];

export function posSessionStateLabel(state: PosSessionState): string {
  switch (state) {
    case "opened":
      return "Opened";
    case "closing":
      return "Closing";
    case "closed":
      return "Closed";
    default:
      return state;
  }
}

export function posSessionStateTone(
  state: PosSessionState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "opened":
      return "success";
    case "closing":
      return "warning";
    case "closed":
      return "neutral";
    default:
      return "neutral";
  }
}

export function posOrderStateLabel(state: PosOrderState): string {
  switch (state) {
    case "done":
      return "Done";
    case "refunded":
      return "Refunded";
    default:
      return state;
  }
}

export function posOrderStateTone(
  state: PosOrderState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "done":
      return "success";
    case "refunded":
      return "danger";
    default:
      return "neutral";
  }
}

export function posSessionStateConfig(state: PosSessionState): StatusConfig {
  return {
    label: posSessionStateLabel(state),
    tone: posSessionStateTone(state),
  };
}

export function posOrderStateConfig(state: PosOrderState): StatusConfig {
  return {
    label: posOrderStateLabel(state),
    tone: posOrderStateTone(state),
  };
}

export function canCloseSession(state: PosSessionState): boolean {
  return state === "opened";
}

export function canStartClosing(state: PosSessionState): boolean {
  return state === "opened";
}

export function canPlaceOrder(state: PosSessionState): boolean {
  return state === "opened";
}

export function canInvoiceOrder(order: PosOrder): boolean {
  return order.state === "done" && order.invoiceId == null;
}

export function canRefundOrder(order: PosOrder): boolean {
  return order.state === "done";
}

export function orderLineSubtotal(line: PosOrderLine): number {
  const afterDiscount = line.unitPrice * line.qty * (1 - line.discountPct / 100);
  return afterDiscount;
}

export function orderTotal(lines: PosOrderLine[]): number {
  return lines.reduce((sum, line) => sum + orderLineSubtotal(line), 0);
}

export function orderTaxTotal(lines: PosOrderLine[]): number {
  return lines.reduce((sum, line) => sum + line.priceTax, 0);
}

export function sessionOrderTotal(orders: PosOrder[]): number {
  return orders.filter((o) => o.state === "done").reduce((sum, o) => sum + o.amountTotal, 0);
}

export function sessionPaymentSummary(
  payments: { method: string; amount: number }[],
): { method: string; total: number }[] {
  const map = new Map<string, number>();
  for (const p of payments) {
    map.set(p.method, (map.get(p.method) ?? 0) + p.amount);
  }
  return Array.from(map.entries()).map(([method, total]) => ({
    method,
    total,
  }));
}

export function configDisplayName(config: PosConfig): string {
  return config.name || `POS-${config.id}`;
}
