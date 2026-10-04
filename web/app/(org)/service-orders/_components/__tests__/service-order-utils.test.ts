import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canBill,
  canCancel,
  canComplete,
  canSchedule,
  canStart,
  formatLineType,
  formatPriority,
  formatServiceType,
  serviceOrderStateTone,
} from "../service-order-utils";

describe("service-order-utils", () => {
  describe("canSchedule", () => {
    it("should allow scheduling new orders", () => {
      expect(canSchedule("new")).toBe(true);
    });

    it("should not allow scheduling non-new orders", () => {
      expect(canSchedule("scheduled")).toBe(false);
      expect(canSchedule("in_progress")).toBe(false);
      expect(canSchedule("done")).toBe(false);
      expect(canSchedule("invoiced")).toBe(false);
      expect(canSchedule("cancelled")).toBe(false);
    });
  });

  describe("canStart", () => {
    it("should allow starting scheduled orders", () => {
      expect(canStart("scheduled")).toBe(true);
    });

    it("should not allow starting non-scheduled orders", () => {
      expect(canStart("new")).toBe(false);
      expect(canStart("in_progress")).toBe(false);
      expect(canStart("done")).toBe(false);
      expect(canStart("invoiced")).toBe(false);
      expect(canStart("cancelled")).toBe(false);
    });
  });

  describe("canComplete", () => {
    it("should allow completing in-progress orders", () => {
      expect(canComplete("in_progress")).toBe(true);
    });

    it("should not allow completing non-in-progress orders", () => {
      expect(canComplete("new")).toBe(false);
      expect(canComplete("scheduled")).toBe(false);
      expect(canComplete("done")).toBe(false);
      expect(canComplete("invoiced")).toBe(false);
      expect(canComplete("cancelled")).toBe(false);
    });
  });

  describe("canBill", () => {
    it("should allow billing done orders", () => {
      expect(canBill("done")).toBe(true);
    });

    it("should not allow billing non-done orders", () => {
      expect(canBill("new")).toBe(false);
      expect(canBill("scheduled")).toBe(false);
      expect(canBill("in_progress")).toBe(false);
      expect(canBill("invoiced")).toBe(false);
      expect(canBill("cancelled")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("should allow cancelling new or scheduled orders", () => {
      expect(canCancel("new")).toBe(true);
      expect(canCancel("scheduled")).toBe(true);
    });

    it("should not allow cancelling other states", () => {
      expect(canCancel("in_progress")).toBe(false);
      expect(canCancel("done")).toBe(false);
      expect(canCancel("invoiced")).toBe(false);
      expect(canCancel("cancelled")).toBe(false);
    });
  });

  describe("serviceOrderStateTone", () => {
    it("should return neutral for new", () => {
      expect(serviceOrderStateTone("new")).toBe("neutral");
    });

    it("should return info for scheduled", () => {
      expect(serviceOrderStateTone("scheduled")).toBe("info");
    });

    it("should return warning for in_progress", () => {
      expect(serviceOrderStateTone("in_progress")).toBe("warning");
    });

    it("should return success for done", () => {
      expect(serviceOrderStateTone("done")).toBe("success");
    });

    it("should return info for invoiced", () => {
      expect(serviceOrderStateTone("invoiced")).toBe("info");
    });

    it("should return danger for cancelled", () => {
      expect(serviceOrderStateTone("cancelled")).toBe("danger");
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

  describe("formatCurrency", () => {
    it("should format as currency", () => {
      const result = formatMoney(5000, { currency: "USD" });
      expect(result).toContain("5");
      expect(result).toContain("000");
    });
  });

  describe("formatServiceType", () => {
    it("should format repair", () => {
      expect(formatServiceType("repair")).toBe("Perbaikan");
    });

    it("should format maintenance", () => {
      expect(formatServiceType("maintenance")).toBe("Perawatan");
    });

    it("should format installation", () => {
      expect(formatServiceType("installation")).toBe("Installation");
    });

    it("should format inspection", () => {
      expect(formatServiceType("inspection")).toBe("Inspection");
    });
  });

  describe("formatPriority", () => {
    it("should format low priority", () => {
      expect(formatPriority(1)).toBe("Rendah");
    });

    it("should format normal priority", () => {
      expect(formatPriority(2)).toBe("Normal");
    });

    it("should format high priority", () => {
      expect(formatPriority(3)).toBe("Tinggi");
    });

    it("should format urgent priority", () => {
      expect(formatPriority(4)).toBe("Mendesak");
    });
  });

  describe("formatLineType", () => {
    it("should format part", () => {
      expect(formatLineType("part")).toBe("Part");
    });

    it("should format labor", () => {
      expect(formatLineType("labor")).toBe("Labor");
    });

    it("should format expense", () => {
      expect(formatLineType("expense")).toBe("Biaya");
    });
  });
});
