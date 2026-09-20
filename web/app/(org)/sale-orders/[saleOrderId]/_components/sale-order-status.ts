import type { SaleOrder } from "@/lib/services/swantara";

export const orderSteps = ["draft", "sent", "confirmed", "done"] as const;

export function orderStateIndex(state: SaleOrder["state"]): number {
  const idx = orderSteps.indexOf(state as (typeof orderSteps)[number]);
  return idx >= 0 ? idx : 0;
}

export function getOrderStatuses() {
  return {
    draft: { label: "Draft", tone: "neutral" as const },
    sent: { label: "Sent", tone: "info" as const },
    confirmed: { label: "Confirmed", tone: "warning" as const },
    done: { label: "Done", tone: "success" as const },
    cancelled: { label: "Cancelled", tone: "danger" as const },
  };
}
