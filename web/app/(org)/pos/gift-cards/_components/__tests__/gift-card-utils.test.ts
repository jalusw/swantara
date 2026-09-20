import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import { canRedeem, canRefund, formatGiftCardState, giftCardStateTone } from "../gift-card-utils";

describe("gift-card-utils", () => {
  describe("canRedeem", () => {
    it("should allow redeeming active cards", () => {
      expect(canRedeem("active")).toBe(true);
    });

    it("should not allow redeeming non-active cards", () => {
      expect(canRedeem("used")).toBe(false);
      expect(canRedeem("expired")).toBe(false);
      expect(canRedeem("cancelled")).toBe(false);
    });
  });

  describe("canRefund", () => {
    it("should allow refunding active or used cards", () => {
      expect(canRefund("active")).toBe(true);
      expect(canRefund("used")).toBe(true);
    });

    it("should not allow refunding other states", () => {
      expect(canRefund("expired")).toBe(false);
      expect(canRefund("cancelled")).toBe(false);
    });
  });

  describe("giftCardStateTone", () => {
    it("should return success for active", () => {
      expect(giftCardStateTone("active")).toBe("success");
    });

    it("should return info for used", () => {
      expect(giftCardStateTone("used")).toBe("info");
    });

    it("should return warning for expired", () => {
      expect(giftCardStateTone("expired")).toBe("warning");
    });

    it("should return danger for cancelled", () => {
      expect(giftCardStateTone("cancelled")).toBe("danger");
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
    it("should format as currency with default USD", () => {
      const result = formatMoney(5000, { currency: "USD" });
      expect(result).toContain("5");
      expect(result).toContain("000");
    });

    it("should format as currency with explicit code", () => {
      const result = formatMoney(5000, { currency: "USD" });
      expect(result).toContain("5");
      expect(result).toContain("000");
    });
  });

  describe("formatGiftCardState", () => {
    it("should format active", () => {
      expect(formatGiftCardState("active")).toBe("Active");
    });

    it("should format used", () => {
      expect(formatGiftCardState("used")).toBe("Used");
    });

    it("should format expired", () => {
      expect(formatGiftCardState("expired")).toBe("Expired");
    });

    it("should format cancelled", () => {
      expect(formatGiftCardState("cancelled")).toBe("Cancelled");
    });
  });
});
