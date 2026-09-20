import { describe, expect, it } from "vitest";
import { formatDate } from "@/lib/utils";
import { frequencyLabel } from "../maintenance-plan-utils";

describe("maintenance-plan-utils", () => {
  describe("frequencyLabel", () => {
    it("should format daily", () => {
      expect(frequencyLabel("daily")).toBe("Daily");
    });

    it("should format weekly", () => {
      expect(frequencyLabel("weekly")).toBe("Weekly");
    });

    it("should format monthly", () => {
      expect(frequencyLabel("monthly")).toBe("Monthly");
    });

    it("should format quarterly", () => {
      expect(frequencyLabel("quarterly")).toBe("Quarterly");
    });

    it("should format yearly", () => {
      expect(frequencyLabel("yearly")).toBe("Yearly");
    });
  });

  describe("formatDate", () => {
    it("should format a valid date", () => {
      const result = formatDate(new Date("2024-03-15"));
      expect(result).toContain("Mar");
      expect(result).toContain("15");
      expect(result).toContain("2024");
    });

    it("should return dash for null", () => {
      expect(formatDate(null, { nullFallback: "—" })).toBe("—");
    });
  });
});
