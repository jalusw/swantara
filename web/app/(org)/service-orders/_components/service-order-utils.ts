import type { ServiceOrder } from "@/lib/services/swantara";

export type ServiceOrderState = ServiceOrder["state"];

export function canSchedule(state: ServiceOrderState): boolean {
  return state === "new";
}

export function canStart(state: ServiceOrderState): boolean {
  return state === "scheduled";
}

export function canComplete(state: ServiceOrderState): boolean {
  return state === "in_progress";
}

export function canBill(state: ServiceOrderState): boolean {
  return state === "done";
}

export function canCancel(state: ServiceOrderState): boolean {
  return state === "new" || state === "scheduled";
}

export function serviceOrderStateTone(
  state: ServiceOrderState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "new":
      return "neutral";
    case "scheduled":
      return "info";
    case "in_progress":
      return "warning";
    case "done":
      return "success";
    case "invoiced":
      return "info";
    case "cancelled":
      return "danger";
    default:
      return "neutral";
  }
}

export function formatServiceType(type: ServiceOrder["type"]): string {
  switch (type) {
    case "repair":
      return "Repair";
    case "maintenance":
      return "Maintenance";
    case "installation":
      return "Installation";
    case "inspection":
      return "Inspection";
    default:
      return type;
  }
}

export function formatPriority(priority: number): string {
  switch (priority) {
    case 1:
      return "Low";
    case 2:
      return "Normal";
    case 3:
      return "High";
    case 4:
      return "Urgent";
    default:
      return String(priority);
  }
}

export function formatLineType(lineType: string): string {
  switch (lineType) {
    case "part":
      return "Part";
    case "labor":
      return "Labor";
    case "expense":
      return "Expense";
    default:
      return lineType;
  }
}
