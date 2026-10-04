import type { Employee } from "@/lib/services/swantara";

export type EmployeeState = Employee["active"];

export function canEdit(employee: Employee): boolean {
  return employee.active;
}

export function canTerminate(employee: Employee): boolean {
  return employee.active;
}

export function employeeStateLabel(active: boolean): "active" | "inactive" {
  return active ? "active" : "inactive";
}

export function employeeStateTone(
  active: boolean,
): "neutral" | "success" | "warning" | "danger" | "info" {
  return active ? "success" : "neutral";
}

export function employmentTypeLabel(type: Employee["employmentType"]): string {
  switch (type) {
    case "full_time":
      return "Full time";
    case "part_time":
      return "Part time";
    case "contract":
      return "Kontrak";
    default:
      return type;
  }
}
