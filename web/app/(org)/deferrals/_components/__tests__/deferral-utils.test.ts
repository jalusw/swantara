import { describe, expect, it } from "vitest";
import { formatDate, formatMoney } from "@/lib/utils";
import { canCancel, canRecognize, deferralStateTone, formatDeferralType } from "../deferral-utils";

describe("deferral-utils", () => {
  describe("canRecognize", () => {
    it("should allow recognizing running deferrals", () => {
      expect(canRecognize("running")).toBe(true);
    });

    it("should not allow recognizing non-running deferrals", () => {
      expect(canRecognize("draft")).toBe(false);
      expect(canRecognize("done")).toBe(false);
      expect(canRecognize("cancelled")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("should allow cancelling draft deferrals", () => {
      expect(canCancel("draft")).toBe(true);
    });

    it("should allow cancelling running deferrals", () => {
      expect(canCancel("running")).toBe(true);
    });

    it("should not allow cancelling done or cancelled deferrals", () => {
      expect(canCancel("done")).toBe(false);
      expect(canCancel("cancelled")).toBe(false);
    });
  });

  describe("deferralStateTone", () => {
    it("should return neutral for draft", () => {
      expect(deferralStateTone("draft")).toBe("neutral");
    });

    it("should return success for running", () => {
      expect(deferralStateTone("running")).toBe("success");
    });

    it("should return info for done", () => {
      expect(deferralStateTone("done")).toBe("info");
    });

    it("should return danger for cancelled", () => {
      expect(deferralStateTone("cancelled")).toBe("danger");
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

  describe("formatDeferralType", () => {
    it("should format deferred_revenue", () => {
      expect(formatDeferralType("deferred_revenue")).toBe("Deferred Revenue");
    });

    it("should format deferred_expense", () => {
      expect(formatDeferralType("deferred_expense")).toBe("Deferred Expense");
    });

    it("should format prepaid", () => {
      expect(formatDeferralType("prepaid")).toBe("Prepaid");
    });
  });
});
