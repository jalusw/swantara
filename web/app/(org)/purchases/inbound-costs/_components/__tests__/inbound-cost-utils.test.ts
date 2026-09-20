import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import { canCancel, canPost, formatSplitMethod, inboundCostStateTone } from "../inbound-cost-utils";

describe("inbound-cost-utils", () => {
  describe("canPost", () => {
    it("should allow posting draft landed costs", () => {
      expect(canPost("draft")).toBe(true);
    });

    it("should not allow posting non-draft landed costs", () => {
      expect(canPost("posted")).toBe(false);
      expect(canPost("cancelled")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("should allow cancelling draft landed costs", () => {
      expect(canCancel("draft")).toBe(true);
    });

    it("should not allow cancelling non-draft landed costs", () => {
      expect(canCancel("posted")).toBe(false);
      expect(canCancel("cancelled")).toBe(false);
    });
  });

  describe("inboundCostStateTone", () => {
    it("should return neutral for draft", () => {
      expect(inboundCostStateTone("draft")).toBe("neutral");
    });

    it("should return success for posted", () => {
      expect(inboundCostStateTone("posted")).toBe("success");
    });

    it("should return danger for cancelled", () => {
      expect(inboundCostStateTone("cancelled")).toBe("danger");
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

  describe("formatCurrency", () => {
    it("should format as currency", () => {
      const result = formatMoney(5000, { currency: "USD" });
      expect(result).toContain("5");
      expect(result).toContain("000");
    });
  });

  describe("formatSplitMethod", () => {
    it("should format by_quantity", () => {
      expect(formatSplitMethod("by_quantity")).toBe("By quantity");
    });

    it("should format by_value", () => {
      expect(formatSplitMethod("by_value")).toBe("By value");
    });

    it("should format by_weight", () => {
      expect(formatSplitMethod("by_weight")).toBe("By weight");
    });

    it("should format by_volume", () => {
      expect(formatSplitMethod("by_volume")).toBe("By volume");
    });
  });
});
