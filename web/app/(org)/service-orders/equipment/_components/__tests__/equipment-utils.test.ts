import { describe, expect, it } from "vitest";
import { formatDate } from "@/lib/utils";
import { equipmentStateTone, formatEquipmentState } from "../equipment-utils";

describe("equipment-utils", () => {
  describe("equipmentStateTone", () => {
    it("should return success for active", () => {
      expect(equipmentStateTone("active")).toBe("success");
    });

    it("should return neutral for inactive", () => {
      expect(equipmentStateTone("inactive")).toBe("neutral");
    });

    it("should return warning for maintenance", () => {
      expect(equipmentStateTone("maintenance")).toBe("warning");
    });

    it("should return danger for retired", () => {
      expect(equipmentStateTone("retired")).toBe("danger");
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

  describe("formatEquipmentState", () => {
    it("should format active", () => {
      expect(formatEquipmentState("active")).toBe("Active");
    });

    it("should format inactive", () => {
      expect(formatEquipmentState("inactive")).toBe("Inactive");
    });

    it("should format maintenance", () => {
      expect(formatEquipmentState("maintenance")).toBe("Maintenance");
    });

    it("should format retired", () => {
      expect(formatEquipmentState("retired")).toBe("Retired");
    });
  });
});
