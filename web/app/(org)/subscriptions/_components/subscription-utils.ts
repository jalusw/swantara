import type { Subscription } from "@/lib/services/swantara";

export type SubscriptionState = Subscription["state"];

export function canActivate(state: SubscriptionState): boolean {
  return state === "draft";
}

export function canPause(state: SubscriptionState): boolean {
  return state === "active";
}

export function canResume(state: SubscriptionState): boolean {
  return state === "paused";
}

export function canChurn(state: SubscriptionState): boolean {
  return state === "active" || state === "paused";
}

export function canClose(state: SubscriptionState): boolean {
  return state === "active" || state === "paused";
}

export function subscriptionStateTone(
  state: SubscriptionState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "active":
      return "success";
    case "paused":
      return "warning";
    case "churned":
      return "danger";
    case "closed":
      return "info";
    default:
      return "neutral";
  }
}

export function formatInterval(interval: string, count: number): string {
  if (count === 1) return interval;
  return `${interval} (x${count})`;
}
