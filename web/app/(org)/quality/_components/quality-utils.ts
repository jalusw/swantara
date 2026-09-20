import type { QualityAlert, QualityCheck, QualityPoint } from "@/lib/services/swantara";

export type CheckResult = QualityCheck["result"];
export type AlertState = QualityAlert["state"];
export type TestType = QualityPoint["testType"];

export function checkResultLabel(result: CheckResult): string {
  switch (result) {
    case "pending":
      return "Pending";
    case "pass":
      return "Pass";
    case "fail":
      return "Fail";
    default:
      return result;
  }
}

export function alertStateLabel(state: AlertState): string {
  switch (state) {
    case "open":
      return "Open";
    case "in_progress":
      return "In Progress";
    case "solved":
      return "Solved";
    case "cancelled":
      return "Cancelled";
    default:
      return state;
  }
}

export function testTypeLabel(testType: TestType): string {
  switch (testType) {
    case "pass_fail":
      return "Pass fail";
    case "measure":
      return "Measure";
    case "instruction":
      return "Instruction";
    default:
      return testType;
  }
}

export function severityLabel(severity: string): string {
  switch (severity) {
    case "low":
      return "Low";
    case "medium":
      return "Medium";
    case "high":
      return "High";
    case "critical":
      return "Critical";
    default:
      return severity;
  }
}

export function checkResultTone(
  result: CheckResult,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (result) {
    case "pending":
      return "info";
    case "pass":
      return "success";
    case "fail":
      return "danger";
    default:
      return "neutral";
  }
}

export function alertStateTone(
  state: AlertState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "open":
      return "danger";
    case "in_progress":
      return "warning";
    case "solved":
      return "success";
    case "cancelled":
      return "neutral";
    default:
      return "neutral";
  }
}

export function canRecordResult(check: QualityCheck): boolean {
  return check.result === "pending";
}

export function canTransitionAlert(current: AlertState, next: AlertState): boolean {
  if (current === "open" && next === "in_progress") return true;
  if (current === "open" && next === "cancelled") return true;
  if (current === "in_progress" && next === "solved") return true;
  if (current === "in_progress" && next === "cancelled") return true;
  return false;
}

export const alertStateSteps: AlertState[] = ["open", "in_progress", "solved"];

export function alertStateIndex(state: AlertState): number {
  return alertStateSteps.indexOf(state);
}

export function checksWithFailedCount(checks: QualityCheck[]): number {
  return checks.filter((c) => c.result === "fail").length;
}

export function checksPendingCount(checks: QualityCheck[]): number {
  return checks.filter((c) => c.result === "pending").length;
}

export function checksPassRate(checks: QualityCheck[]): number {
  const resolved = checks.filter((c) => c.result !== "pending");
  if (resolved.length === 0) return 0;
  const passed = resolved.filter((c) => c.result === "pass").length;
  return Math.round((passed / resolved.length) * 100);
}

export function isMeasureType(point: QualityPoint): boolean {
  return point.testType === "measure";
}

export function isPassFailType(point: QualityPoint): boolean {
  return point.testType === "pass_fail";
}

export function normRangeLabel(point: QualityPoint): string {
  if (point.normMin == null && point.normMax == null) return "—";
  if (point.normMin != null && point.normMax != null) return `${point.normMin} – ${point.normMax}`;
  if (point.normMin != null) return `≥ ${point.normMin}`;
  return `≤ ${point.normMax}`;
}

export function alertTraceability(alert: QualityAlert): string {
  const parts: string[] = [];
  if (alert.checkId) parts.push(`Check #${alert.checkId}`);
  if (alert.itemId) parts.push(`Item #${alert.itemId}`);
  if (alert.batchId) parts.push(`Batch #${alert.batchId}`);
  return parts.join(" · ") || "—";
}

const toneToBadgeVariantMap: Record<string, "default" | "destructive" | "secondary"> = {
  success: "default",
  danger: "destructive",
};

export function toneToBadgeVariant(tone: string): "default" | "destructive" | "secondary" {
  return toneToBadgeVariantMap[tone] ?? "secondary";
}
