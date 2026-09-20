import { describe, expect, it } from "vitest";
import { formatDate, formatHours } from "@/lib/utils";
import { hoursByDate, totalHours } from "../timesheet-utils";

const timesheets = [
  {
    id: 1,
    employeeId: 1,
    date: "2024-03-15",
    hours: 8,
    projectId: null,
    taskId: null,
    dimensionId: null,
    description: null,
  },
  {
    id: 2,
    employeeId: 1,
    date: "2024-03-15",
    hours: 2,
    projectId: null,
    taskId: null,
    dimensionId: null,
    description: null,
  },
  {
    id: 3,
    employeeId: 1,
    date: "2024-03-16",
    hours: 6,
    projectId: null,
    taskId: null,
    dimensionId: null,
    description: null,
  },
];

describe("timesheet-utils", () => {
  describe("totalHours", () => {
    it("should sum all hours", () => {
      expect(totalHours(timesheets)).toBe(16);
    });

    it("should return 0 for empty list", () => {
      expect(totalHours([])).toBe(0);
    });
  });

  describe("hoursByDate", () => {
    it("should group hours by date", () => {
      const result = hoursByDate(timesheets);
      expect(result.get("2024-03-15")).toBe(10);
      expect(result.get("2024-03-16")).toBe(6);
    });

    it("should return empty map for empty list", () => {
      expect(hoursByDate([]).size).toBe(0);
    });
  });

  describe("formatDate", () => {
    it("should format a valid date", () => {
      const result = formatDate("2024-03-15");
      expect(result).toContain("Mar");
      expect(result).toContain("15");
      expect(result).toContain("2024");
    });
  });

  describe("formatHours", () => {
    it("should format hours with one decimal", () => {
      expect(formatHours(8)).toBe("8.0h");
      expect(formatHours(7.5)).toBe("7.5h");
    });
  });
});
