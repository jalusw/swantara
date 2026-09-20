import type { Attendance } from "@/lib/services/swantara";

export function isCheckedIn(attendance: Attendance): boolean {
  return attendance.checkIn != null && attendance.checkOut == null;
}

export function workedHoursTone(hours: number): "neutral" | "success" | "warning" | "info" {
  if (hours >= 8) return "success";
  if (hours > 0) return "warning";
  return "neutral";
}
