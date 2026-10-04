import { describe, expect, it } from "vitest";
import { formatBytes, formatDate, formatDateTime, formatDuration, formatNumber } from "@/lib/utils";

describe("formatNumber", () => {
  it("formats integers with grouping", () => {
    expect(formatNumber(12500)).toMatch(/12\.500/);
  });

  it("respects maximumFractionDigits", () => {
    expect(formatNumber(12.5, { maximumFractionDigits: 0 })).toBe("13");
  });
});

describe("formatDate", () => {
  it("formats a date with medium date style", () => {
    expect(formatDate("2026-01-02", { locale: "en-US" })).toBe("Jan 2, 2026");
  });
});

describe("formatDateTime", () => {
  it("includes a time component", () => {
    expect(
      formatDateTime("2026-01-02T10:30:00Z", {
        locale: "en-US",
        timeZone: "UTC",
      }),
    ).toBe("Jan 2, 2026, 10:30 AM");
  });
});

describe("formatBytes", () => {
  it("formats bytes as B", () => {
    expect(formatBytes(512)).toBe("512 B");
  });

  it("formats kilobytes", () => {
    expect(formatBytes(2048)).toBe("2 KB");
  });

  it("formats megabytes", () => {
    expect(formatBytes(5 * 1024 * 1024)).toBe("5 MB");
  });
});

describe("formatDuration", () => {
  it("formats sub-second runs in milliseconds", () => {
    expect(formatDuration(250)).toBe("250.00ms");
  });

  it("formats runs over one second in seconds", () => {
    expect(formatDuration(1500)).toBe("1.50s");
  });
});
