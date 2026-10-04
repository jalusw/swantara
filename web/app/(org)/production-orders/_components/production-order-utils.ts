import type { MOComponent, ProductionOrder } from "@/lib/services/swantara";

export type MoState = ProductionOrder["state"];

export function moStateLabel(state: MoState): string {
  switch (state) {
    case "draft":
      return "Draft";
    case "confirmed":
      return "Dikonfirmasi";
    case "planned":
      return "Direncanakan";
    case "in_progress":
      return "Berjalan";
    case "done":
      return "Selesai";
    case "cancelled":
      return "Dibatalkan";
    default:
      return state;
  }
}

export function moOriginLabel(origin: string): string {
  switch (origin) {
    case "manual":
      return "Manual";
    case "sale_order":
      return "Sale Order";
    case "reorder":
      return "Reorder";
    default:
      return origin;
  }
}

export function canConfirmMo(state: MoState): boolean {
  return state === "draft";
}

export function canPlanMo(state: MoState): boolean {
  return state === "confirmed";
}

export function canStartMo(state: MoState): boolean {
  return state === "planned";
}

export function canProduceMo(state: MoState): boolean {
  return state === "in_progress";
}

export function canSettleMo(state: MoState): boolean {
  return state === "in_progress";
}

export function canCancelMo(state: MoState): boolean {
  return state !== "done" && state !== "cancelled";
}

export function moStateTone(state: MoState): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "confirmed":
      return "info";
    case "planned":
      return "info";
    case "in_progress":
      return "warning";
    case "done":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function moProgress(
  productionOrder: Pick<ProductionOrder, "qtyToProduce" | "qtyProduced">,
): number {
  if (productionOrder.qtyToProduce === 0) return 0;
  return Math.min(
    100,
    Math.round((productionOrder.qtyProduced / productionOrder.qtyToProduce) * 100),
  );
}

export function moProgressColor(progress: number): string {
  if (progress >= 100) return "text-success";
  if (progress >= 50) return "text-warning";
  return "text-muted-foreground";
}

export function componentConsumptionProgress(
  component: Pick<MOComponent, "qtyPlanned" | "qtyConsumed">,
): number {
  if (component.qtyPlanned === 0) return 0;
  return Math.min(100, Math.round((component.qtyConsumed / component.qtyPlanned) * 100));
}

export function moSteps(state: MoState): string[] {
  const allSteps = ["draft", "confirmed", "planned", "in_progress", "done"];
  const stateIndex = allSteps.indexOf(state);
  if (state === "cancelled") return ["draft", "cancelled"];
  return allSteps.slice(0, stateIndex + 1);
}
