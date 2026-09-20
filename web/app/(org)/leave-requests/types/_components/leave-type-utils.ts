import type { LeaveRequest } from "@/lib/services/swantara";

export type LeaveRequestState = LeaveRequest["state"];

export function canSubmit(state: LeaveRequestState): boolean {
  return state === "draft";
}

export function canApprove(state: LeaveRequestState): boolean {
  return state === "submitted";
}

export function canRefuse(state: LeaveRequestState): boolean {
  return state === "submitted";
}

export function leaveRequestStateTone(
  state: LeaveRequestState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "submitted":
      return "info";
    case "approved":
      return "success";
    case "refused":
      return "danger";
    default:
      return "neutral";
  }
}
