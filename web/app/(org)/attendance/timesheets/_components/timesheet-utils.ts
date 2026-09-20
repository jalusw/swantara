import type { Timesheet } from "@/lib/services/swantara";

export function totalHours(timesheets: Timesheet[]): number {
  return timesheets.reduce((sum, ts) => sum + ts.hours, 0);
}

export function hoursByDate(timesheets: Timesheet[]): Map<string, number> {
  const byDate = new Map<string, number>();
  for (const ts of timesheets) {
    byDate.set(ts.date, (byDate.get(ts.date) ?? 0) + ts.hours);
  }
  return byDate;
}
