import { describe, expect, it } from "vitest";
import type { Attendance } from "@/lib/services/swantara";
import { formatDateTime, formatHours } from "@/lib/utils";
import { isCheckedIn, workedHoursTone } from "../attendance-utils";

const checkedIn: Attendance = {
  id: 1,
  employeeId: 1,
  checkIn: "2024-03-15T08:00:00Z",
  checkOut: null,
  workedHours: 0,
};

const checkedOut: Attendance = {
  id: 2,
  employeeId: 1,
  checkIn: "2024-03-15T08:00:00Z",
  checkOut: "2024-03-15T17:00:00Z",
  workedHours: 9,
};

describe("attendance-utils", () => {
  describe("isCheckedIn", () => {
    it("should return true when checked in but not out", () => {
      expect(isCheckedIn(checkedIn)).toBe(true);
    });

    it("should return false when checked out", () => {
      expect(isCheckedIn(checkedOut)).toBe(false);
    });
  });

  describe("formatDateTime", () => {
    it("should return dash for null", () => {
      expect(formatDateTime(null, { nullFallback: "—" })).toBe("—");
    });

    it("should format a valid date time", () => {
      const result = formatDateTime("2024-03-15T08:00:00Z", { nullFallback: "—" });
      expect(result).toContain("Mar");
      expect(result).toContain("15");
    });
  });

  describe("formatHours", () => {
    it("should format hours with one decimal", () => {
      expect(formatHours(9)).toBe("9.0h");
      expect(formatHours(7.5)).toBe("7.5h");
    });
  });

  describe("workedHoursTone", () => {
    it("should return success for 8+ hours", () => {
      expect(workedHoursTone(8)).toBe("success");
      expect(workedHoursTone(9)).toBe("success");
    });

    it("should return warning for 1-7 hours", () => {
      expect(workedHoursTone(4)).toBe("warning");
      expect(workedHoursTone(7.5)).toBe("warning");
    });

    it("should return neutral for 0 hours", () => {
      expect(workedHoursTone(0)).toBe("neutral");
    });
  });
});
