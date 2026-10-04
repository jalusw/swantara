import type { SaleOrder } from "@/lib/services/swantara";

export const orderSteps = ["draft", "sent", "confirmed", "done"] as const;

export function orderStateIndex(state: SaleOrder["state"]): number {
  const idx = orderSteps.indexOf(state as (typeof orderSteps)[number]);
  return idx >= 0 ? idx : 0;
}

export function getOrderStatuses(label?: (state: string) => string) {
  const text = (state: string, fallback: string) => (label ? label(state) : fallback);
  return {
    draft: { label: text("draft", "Draft"), tone: "neutral" as const },
    sent: { label: text("sent", "Terkirim"), tone: "info" as const },
    confirmed: { label: text("confirmed", "Dikonfirmasi"), tone: "warning" as const },
    done: { label: text("done", "Selesai"), tone: "success" as const },
    cancelled: { label: text("cancelled", "Dibatalkan"), tone: "danger" as const },
  };
}
