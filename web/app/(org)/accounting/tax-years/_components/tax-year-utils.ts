import type { TaxYear } from "@/lib/services/swantara";

export function isYearLocked(year: TaxYear): boolean {
  return year.state === "done" || year.state === "locked";
}

export function isPeriodLocked(year: TaxYear, date: string): boolean {
  if (year.state === "done") return true;
  if (year.state === "locked") return true;
  const d = new Date(date);
  const start = year.dateStart;
  const end = year.dateEnd;
  if (start && d < start) return true;
  if (end && d > end) return true;
  return false;
}

export function yearDurationMonths(year: TaxYear): number {
  const start = year.dateStart;
  const end = year.dateEnd;
  if (!start || !end) return 0;
  const s = new Date(start);
  const e = new Date(end);
  return (e.getFullYear() - s.getFullYear()) * 12 + (e.getMonth() - s.getMonth()) + 1;
}
