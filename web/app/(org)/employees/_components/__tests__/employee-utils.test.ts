import { describe, expect, it } from "vitest";
import { formatDate } from "@/lib/utils";
import {
  canEdit,
  canTerminate,
  employeeStateLabel,
  employeeStateTone,
  employmentTypeLabel,
} from "../employee-utils";

const activeEmployee = {
  id: 1,
  active: true,
  employmentType: "full_time" as const,
  hireDate: "2024-03-15",
  employeeNumber: "EMP-001",
  departmentId: null,
  jobPositionId: null,
  managerId: null,
  organizationId: 1,
  contactId: 1,
  userId: null,
  terminationDate: null,
  workLocation: null,
};

const inactiveEmployee = { ...activeEmployee, active: false };

describe("employee-utils", () => {
  describe("canEdit", () => {
    it("should allow editing active employees", () => {
      expect(canEdit(activeEmployee)).toBe(true);
    });

    it("should not allow editing inactive employees", () => {
      expect(canEdit(inactiveEmployee)).toBe(false);
    });
  });

  describe("canTerminate", () => {
    it("should allow terminating active employees", () => {
      expect(canTerminate(activeEmployee)).toBe(true);
    });

    it("should not allow terminating inactive employees", () => {
      expect(canTerminate(inactiveEmployee)).toBe(false);
    });
  });

  describe("employeeStateLabel", () => {
    it("should return active for active employees", () => {
      expect(employeeStateLabel(true)).toBe("active");
    });

    it("should return inactive for inactive employees", () => {
      expect(employeeStateLabel(false)).toBe("inactive");
    });
  });

  describe("employeeStateTone", () => {
    it("should return success for active", () => {
      expect(employeeStateTone(true)).toBe("success");
    });

    it("should return neutral for inactive", () => {
      expect(employeeStateTone(false)).toBe("neutral");
    });
  });

  describe("employmentTypeLabel", () => {
    it("should format full_time", () => {
      expect(employmentTypeLabel("full_time")).toBe("Full time");
    });

    it("should format part_time", () => {
      expect(employmentTypeLabel("part_time")).toBe("Part time");
    });

    it("should format contract", () => {
      expect(employmentTypeLabel("contract")).toBe("Contract");
    });
  });

  describe("formatDate", () => {
    it("should return dash for null", () => {
      expect(formatDate(null, { nullFallback: "—" })).toBe("—");
    });

    it("should format a valid date", () => {
      expect(formatDate("2024-03-15")).toContain("Mar");
      expect(formatDate("2024-03-15")).toContain("15");
      expect(formatDate("2024-03-15")).toContain("2024");
    });
  });
});
