import type { ServiceContract } from "@/lib/services/swantara";

export type ServiceContractState = ServiceContract["state"];

export function canActivate(state: ServiceContractState): boolean {
  return state === "draft";
}

export function canCancel(state: ServiceContractState): boolean {
  return state === "draft" || state === "active";
}

export function serviceContractStateTone(
  state: ServiceContractState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "active":
      return "success";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}
