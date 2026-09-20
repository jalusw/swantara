type PerfLogLevel = "info" | "warn" | "error";

const slowThresholdMs = 1000;

export function perfLogLevel(durationMs: number, status: number, hasError = false): PerfLogLevel {
  if (hasError) {
    return "error";
  }
  if (durationMs > slowThresholdMs) {
    return "warn";
  }
  if (status >= 500) {
    return "error";
  }
  return "info";
}

export function formatPerfLog(
  method: string,
  target: string,
  outcome: number | "ERROR",
  duration: string,
): string {
  return `[PERF] ${method} ${target} → ${outcome} (${duration})`;
}
